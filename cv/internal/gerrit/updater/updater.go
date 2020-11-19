// Copyright 2020 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package updater fetches latest CL data from Gerrit.
package updater

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/grpc/grpcutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/gerrit"
	"go.chromium.org/luci/cv/internal/gerrit/gobmap"
)

// UpdateCL fetches latest info from Gerrit.
//
// If datastore already contains snapshot with Gerrit-reported update time equal
// to or after updatedHint, then no updating or querying will be performed.
// To force an update, provide as time.Time{} as updatedHint.
func UpdateCL(ctx context.Context, luciProject, host string, change int64, updatedHint time.Time) (err error) {
	fetcher := fetcher{
		luciProject: luciProject,
		host:        host,
		change:      change,
		updatedHint: updatedHint,
	}
	if fetcher.g, err = gerrit.CurrentClient(ctx, host, luciProject); err != nil {
		return err
	}
	return fetcher.update(ctx)
}

// fetcher efficiently computes new snapshot by fetching data from Gerrit.
//
// It ensures each dependency is resolved to an existing CLID,
// creating CLs in datastore as needed. Schedules tasks to update
// dependencies but doesn't wait for them to complete.
//
// The prior Snapshot, if given, can reduce RPCs made to Gerrit.
type fetcher struct {
	luciProject string
	host        string
	change      int64
	updatedHint time.Time

	g gerrit.CLReaderClient

	externalID changelist.ExternalID
	priorCL    *changelist.CL

	newSnapshot *changelist.Snapshot
	newAcfg     *changelist.ApplicableConfig
}

func (f *fetcher) shouldSkip(ctx context.Context) (skip bool, err error) {
	switch f.priorCL, err = f.externalID.Get(ctx); {
	case err == datastore.ErrNoSuchEntity:
		return false, nil
	case err != nil:
		return false, err
	case f.priorCL.Snapshot == nil:
		// CL is likely created as a dependency and not yet populated.
		return false, nil

	case f.priorCL.Snapshot.GetGerrit().GetInfo() == nil:
		panic(errors.Reason("%s has snapshot without Gerrit Info", f).Err())
	case f.priorCL.ApplicableConfig == nil:
		panic(errors.Reason("%s has snapshot but not ApplicableConfig", f).Err())

	case !f.updatedHint.IsZero() && f.priorCL.Snapshot.IsUpToDate(f.luciProject, f.updatedHint):
		ci := f.priorCL.Snapshot.GetGerrit().Info
		switch acfg, err := gobmap.Lookup(ctx, f.host, ci.GetProject(), ci.GetRef()); {
		case err != nil:
			return false, err
		case acfg.HasProject(f.luciProject):
			logging.Debugf(ctx, "Updating %s to %s skipped, already at %s", f, f.updatedHint,
				f.priorCL.Snapshot.GetExternalUpdateTime().AsTime())
			return true, nil
		default:
			// CL is no longer watched by the given luciProject, even though
			// snapshot is considered up-to-date.
			return true, changelist.Update(ctx, "", f.priorCL.ID, nil /*keep snapshot as is*/, acfg)
		}
	}
	return false, nil
}

func (f *fetcher) update(ctx context.Context) (err error) {
	f.externalID, err = changelist.GobID(f.host, f.change)
	if err != nil {
		return err
	}

	switch skip, err := f.shouldSkip(ctx); {
	case err != nil:
		return err
	case skip:
		return nil
	}

	f.newSnapshot = &changelist.Snapshot{Kind: &changelist.Snapshot_Gerrit{Gerrit: &changelist.Gerrit{}}}
	// TODO(tandrii): optimize for existing CL case.
	if err := f.new(ctx); err != nil {
		return err
	}

	min, cur, err := gerrit.EquivalentPatchsetRange(f.newSnapshot.GetGerrit().GetInfo())
	if err != nil {
		return err
	}
	f.newSnapshot.MinEquivalentPatchset = int32(min)
	f.newSnapshot.Patchset = int32(cur)
	return changelist.Update(ctx, f.externalID, f.clidIfKnown(), f.newSnapshot, f.newAcfg)
}

// new efficiently fetches new snapshot from Gerrit.
func (f *fetcher) new(ctx context.Context) error {
	req := &gerritpb.GetChangeRequest{
		Number:  f.change,
		Project: f.gerritProjectIfKnown(),
		Options: []gerritpb.QueryOption{
			// These are expensive to compute for Gerrit,
			// CV should not do this needlessly.
			gerritpb.QueryOption_ALL_REVISIONS,
			gerritpb.QueryOption_CURRENT_COMMIT,
			gerritpb.QueryOption_DETAILED_LABELS,
			gerritpb.QueryOption_DETAILED_ACCOUNTS,
			gerritpb.QueryOption_MESSAGES,
			gerritpb.QueryOption_SUBMITTABLE,
			// Avoid asking Gerrit to perform expensive operation.
			gerritpb.QueryOption_SKIP_MERGEABLE,
		},
	}
	ci, err := f.g.GetChange(ctx, req)
	switch grpcutil.Code(err) {
	case codes.OK:
		if err := f.ensureNotStale(ctx, ci.GetUpdated()); err != nil {
			return err
		}
		f.newSnapshot.GetGerrit().Info = ci
	case codes.NotFound:
		// Either no access OR CL was deleted.
		return errors.New("not implemented")
	case codes.PermissionDenied:
		return errors.New("not implemented")
	default:
		return unhandledError(ctx, err, "failed to fetch %s/%d", f.host, f.change)
	}

	// TODO(tandrii): implement files & deps.
	return nil
}

// ensureNotStale returns error if given Gerrit updated timestamp is older than
// the updateHint or existing CL state.
func (f *fetcher) ensureNotStale(ctx context.Context, externalUpdateTime *timestamppb.Timestamp) error {
	t := externalUpdateTime.AsTime()
	storedTS := f.priorSnapshot().GetExternalUpdateTime()

	switch {
	case !f.updatedHint.IsZero() && f.updatedHint.After(t):
		logging.Errorf(ctx, "Fetched last Gerrit update of %s, but %s expected", t, f.updatedHint)
	case storedTS != nil && storedTS.AsTime().Before(t):
		logging.Errorf(ctx, "Fetched last Gerrit update of %s, but %s was already seen & stored", t, storedTS.AsTime())
	default:
		return nil
	}
	return errors.Reason("Fetched stale Gerrit data").Tag(transient.Tag).Err()
}

// unhandledError is used to process and annotate Gerrit errors.
func unhandledError(ctx context.Context, err error, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	ann := errors.Annotate(err, msg)
	switch code := grpcutil.Code(err); code {
	case
		codes.OK,
		codes.PermissionDenied,
		codes.NotFound,
		codes.FailedPrecondition:
		// These must be handled before.
		logging.Errorf(ctx, "FIXME unhandled Gerrit error: %s while %s", err, msg)
		return ann.Err()

	case
		codes.InvalidArgument,
		codes.Unauthenticated:
		// This must not happen in practice unless there is a bug in CV or Gerrit.
		logging.Errorf(ctx, "FIXME bug in CV: %s while %s", err, msg)
		return ann.Err()

	case codes.Unimplemented:
		// This shouldn't happen in production, but may happen in development
		// if gerrit.NewRESTClient doesn't actually implement fully the option
		// or entire method that CV is coded to work with.
		logging.Errorf(ctx, "FIXME likely bug in CV: %s while %s", err, msg)
		return ann.Err()

	default:
		// Assume transient. If this turns out non-transient, then its code must be
		// handled explicitly above.
		return ann.Tag(transient.Tag).Err()
	}
}

func (f *fetcher) gerritProjectIfKnown() string {
	if project := f.priorSnapshot().GetGerrit().GetInfo().GetProject(); project != "" {
		return project
	}
	if project := f.newSnapshot.GetGerrit().GetInfo().GetProject(); project != "" {
		return project
	}
	return ""
}

func (f *fetcher) clidIfKnown() changelist.CLID {
	if f.priorCL != nil {
		return f.priorCL.ID
	}
	return 0
}

func (f *fetcher) priorSnapshot() *changelist.Snapshot {
	if f.priorCL != nil {
		return f.priorCL.Snapshot
	}
	return nil
}

func (f *fetcher) priorAcfg() *changelist.ApplicableConfig {
	if f.priorCL != nil {
		return f.priorCL.ApplicableConfig
	}
	return nil
}

// String is used for debug identification of a fetch in errors and logs.
func (f *fetcher) String() string {
	if f.priorCL == nil {
		return fmt.Sprintf("CL(%s/%d)", f.host, f.change)
	}
	return fmt.Sprintf("CL(%s/%d [%d])", f.host, f.change, f.priorCL.ID)
}