// Copyright 2021 The LUCI Authors.
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

package bq

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/gae/service/datastore"

	cvbqpb "go.chromium.org/luci/cv/api/bigquery/v1"
	cfgpb "go.chromium.org/luci/cv/api/config/v2"
	"go.chromium.org/luci/cv/internal/common"
	cvbq "go.chromium.org/luci/cv/internal/common/bq"
	"go.chromium.org/luci/cv/internal/configs/prjcfg"
	"go.chromium.org/luci/cv/internal/metrics"
	"go.chromium.org/luci/cv/internal/migration"
	"go.chromium.org/luci/cv/internal/run"
	"go.chromium.org/luci/cv/internal/tryjob"
)

const (
	// CV's own dataset/table.
	CVDataset = "raw"
	CVTable   = "attempts_cv"

	// Legacy CQ dataset.
	legacyProject    = "commit-queue"
	legacyProjectDev = "commit-queue-dev"
	legacyDataset    = "raw"
	legacyTable      = "attempts"
)

func send(ctx context.Context, env *common.Env, client cvbq.Client, id common.RunID) error {
	r := &run.Run{ID: id}
	switch err := datastore.Get(ctx, r); {
	case err == datastore.ErrNoSuchEntity:
		return errors.Reason("Run not found").Err()
	case err != nil:
		return errors.Annotate(err, "failed to fetch Run").Tag(transient.Tag).Err()
	case !run.IsEnded(r.Status):
		panic(fmt.Errorf("the Run status must be final before sending to BQ"))
	}

	switch r.Mode {
	case run.DryRun, run.FullRun, run.QuickDryRun:
	case run.NewPatchsetRun:
		// New patchset runs need not be exported to BQ at the moment.
		return nil
	default:
		panic(fmt.Errorf("unknown run mode: %s", r.Mode))
	}
	// Load CLs and convert them to GerritChanges including submit status.
	cls, err := run.LoadRunCLs(ctx, r.ID, r.CLs)
	if err != nil {
		return err
	}

	a, err := makeAttempt(ctx, r, cls)
	if err != nil {
		return errors.Annotate(err, "failed to make Attempt").Err()
	}

	// During the migration period when CQDaemon does most checks and triggers
	// builds, CV can't populate all of the fields of Attempt without the
	// information from CQDaemon; so for finished Attempts reported by
	// CQDaemon, we can fill in the remaining fields.
	//
	// TODO(crbug/1225047): After CQDaemon turn-down, this will be unnecessary.
	switch cqda, err := fetchCQDAttempt(ctx, r); {
	case err != nil:
		return err
	case cqda != nil:
		a = reconcileAttempts(a, cqda)
	}

	var wg sync.WaitGroup
	var exportErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		logging.Debugf(ctx, "CV exporting Run to CQ BQ table")
		project := legacyProject
		if env.IsGAEDev {
			project = legacyProjectDev
		}
		exportErr = client.SendRow(ctx, cvbq.Row{
			CloudProject: project,
			Dataset:      legacyDataset,
			Table:        legacyTable,
			OperationID:  "run-" + string(id),
			Payload:      a,
		})
		if exportErr == nil {
			delay := clock.Since(ctx, r.EndTime).Milliseconds()
			metrics.Internal.BigQueryExportDelay.Add(ctx, float64(delay),
				r.ID.LUCIProject(),
				r.ConfigGroupID.Name(),
				string(r.Mode))
		}
	}()

	go func() {
		defer wg.Done()
		// *Always* export to the local CV dataset but the error won't fail the
		// task.
		err := client.SendRow(ctx, cvbq.Row{
			Dataset:     CVDataset,
			Table:       CVTable,
			OperationID: "run-" + string(id),
			Payload:     a,
		})
		if err != nil {
			logging.Errorf(ctx, "failed to export the Run to CV dataset: %s", err)
		}
	}()
	wg.Wait()
	return exportErr
}

func makeAttempt(ctx context.Context, r *run.Run, cls []*run.RunCL) (*cvbqpb.Attempt, error) {
	builds, err := computeAttemptBuilds(ctx, r)
	if err != nil {
		return nil, err
	}
	// TODO(crbug/1173168, crbug/1105669): We want to change the BQ
	// schema so that StartTime is processing start time and CreateTime is
	// trigger time.
	a := &cvbqpb.Attempt{
		Key:                  r.ID.AttemptKey(),
		LuciProject:          r.ID.LUCIProject(),
		ConfigGroup:          r.ConfigGroupID.Name(),
		ClGroupKey:           run.ComputeCLGroupKey(cls, false),
		EquivalentClGroupKey: run.ComputeCLGroupKey(cls, true),
		// Run.CreateTime is trigger time, which corresponds to what CQD sends for
		// StartTime.
		StartTime:            timestamppb.New(r.CreateTime),
		EndTime:              timestamppb.New(r.EndTime),
		Builds:               builds,
		HasCustomRequirement: len(r.Options.GetIncludedTryjobs()) > 0,
	}
	submittedSet := common.MakeCLIDsSet(r.Submission.GetSubmittedCls()...)
	failedSet := common.MakeCLIDsSet(r.Submission.GetFailedCls()...)
	a.GerritChanges = make([]*cvbqpb.GerritChange, len(cls))
	for i, cl := range cls {
		a.GerritChanges[i] = toGerritChange(cl, submittedSet, failedSet, r.Mode)
	}
	a.Status, a.Substatus = attemptStatus(ctx, r)
	return a, nil
}

// toGerritChange creates a GerritChange for the given RunCL.
//
// This includes the submit status of the CL.
func toGerritChange(cl *run.RunCL, submitted, failed common.CLIDsSet, mode run.Mode) *cvbqpb.GerritChange {
	detail := cl.Detail
	ci := detail.GetGerrit().GetInfo()
	gc := &cvbqpb.GerritChange{
		Host:                       detail.GetGerrit().Host,
		Project:                    ci.Project,
		Change:                     ci.Number,
		Patchset:                   int64(detail.Patchset),
		EarliestEquivalentPatchset: int64(detail.MinEquivalentPatchset),
		TriggerTime:                cl.Trigger.Time,
		Mode:                       mode.BQAttemptMode(),
		SubmitStatus:               cvbqpb.GerritChange_PENDING,
		Owner:                      ci.GetOwner().GetEmail(),
	}

	if mode == run.FullRun {
		// Mark the CL submit status as success if it appears in the submitted CLs
		// list, and failure if it does not.
		switch _, submitted := submitted[cl.ID]; {
		case submitted:
			gc.SubmitStatus = cvbqpb.GerritChange_SUCCESS
		case failed.Has(cl.ID):
			gc.SubmitStatus = cvbqpb.GerritChange_FAILURE
		default:
			gc.SubmitStatus = cvbqpb.GerritChange_PENDING
		}
	}
	return gc
}

// fetchCQDAttempt fetches an Attempt from CQDaemon if available.
//
// Returns nil if no Attempt is available.
func fetchCQDAttempt(ctx context.Context, r *run.Run) (*cvbqpb.Attempt, error) {
	v := migration.VerifiedCQDRun{ID: r.ID}
	switch err := datastore.Get(ctx, &v); {
	case err == datastore.ErrNoSuchEntity:
		// A Run may end without a VerifiedCQDRun stored if the Run is canceled.
		logging.Debugf(ctx, "no VerifiedCQDRun found for Run %q", r.ID)
	case err != nil:
		return nil, errors.Annotate(err, "failed to fetch VerifiedCQDRun").Tag(transient.Tag).Err()
	}
	return v.Payload.GetRun().GetAttempt(), nil
}

// reconcileAttempts merges the CV Attempt and CQDaemon Attempt.
//
// Modifies and returns the CV Attempt.
//
// Once CV does the relevant work (keeping track of builds, reading the CL
// description footers, and performing checks) these will no longer have to be
// filled in with the CQDaemon Attempt values.
func reconcileAttempts(a, cqda *cvbqpb.Attempt) *cvbqpb.Attempt {
	// The list of Builds will be known to CV after it starts triggering
	// and tracking builds; until then CQD is the source of truth.
	a.Builds = cqda.Builds
	// Substatus generally indicates a failure reason, which is
	// known once one of the checks fails. CQDaemon may specify
	// a substatus in the case of abort (substatus: MANUAL_CANCEL)
	// or failure (FAILED_TRYJOBS etc.).
	if a.Status == cvbqpb.AttemptStatus_ABORTED || a.Status == cvbqpb.AttemptStatus_FAILURE {
		a.Status = cqda.Status
		a.Substatus = cqda.Substatus
	}
	a.Status = cqda.Status
	a.Substatus = cqda.Substatus
	// The HasCustomRequirement is determined by CL description footers.
	a.HasCustomRequirement = cqda.HasCustomRequirement
	return a
}

// attemptStatus converts a Run status to Attempt status.
func attemptStatus(ctx context.Context, r *run.Run) (cvbqpb.AttemptStatus, cvbqpb.AttemptSubstatus) {
	switch r.Status {
	case run.Status_SUCCEEDED:
		return cvbqpb.AttemptStatus_SUCCESS, cvbqpb.AttemptSubstatus_NO_SUBSTATUS
	case run.Status_FAILED:
		switch {
		case r.Submission != nil && len(r.Submission.Cls) != len(r.Submission.SubmittedCls):
			// In the case that the checks passed but not all CLs were submitted
			// successfully, the Attempt will still have status set to SUCCESS for
			// backwards compatibility (See: crbug.com/1114686). Note that
			// r.Submission is expected to be set only if a submission is attempted,
			// 	meaning all checks passed.
			//
			// TODO(crbug/1114686): Add a new FAILED_SUBMIT substatus, which
			// should be used in the case that some CLs failed to submit after
			// passing checks. (In this case, for backwards compatibility, we
			// will set status = SUCCESS, substatus = FAILED_SUBMIT.)
			return cvbqpb.AttemptStatus_SUCCESS, cvbqpb.AttemptSubstatus_NO_SUBSTATUS
		case r.UseCVTryjobExecutor && r.Tryjobs.GetState().GetStatus() == tryjob.ExecutionState_FAILED:
			return cvbqpb.AttemptStatus_FAILURE, cvbqpb.AttemptSubstatus_FAILED_TRYJOBS
		default:
			// TODO(crbug/1342810): use the failure reason stored in Run entity to
			// decide accurate sub-status. For now, use unapproved because it is the
			// most common failure reason after failed tryjobs.
			return cvbqpb.AttemptStatus_FAILURE, cvbqpb.AttemptSubstatus_UNAPPROVED
		}
	case run.Status_CANCELLED:
		return cvbqpb.AttemptStatus_ABORTED, cvbqpb.AttemptSubstatus_MANUAL_CANCEL
	default:
		logging.Errorf(ctx, "Unexpected attempt status %q", r.Status)
		return cvbqpb.AttemptStatus_ATTEMPT_STATUS_UNSPECIFIED, cvbqpb.AttemptSubstatus_ATTEMPT_SUBSTATUS_UNSPECIFIED
	}
}

func computeAttemptBuilds(ctx context.Context, r *run.Run) ([]*cvbqpb.Build, error) {
	if r.UseCVTryjobExecutor {
		var ret []*cvbqpb.Build
		for i, execution := range r.Tryjobs.GetState().GetExecutions() {
			definition := r.Tryjobs.GetState().GetRequirement().GetDefinitions()[i]
			for _, executionAttempt := range execution.GetAttempts() {
				if executionAttempt.GetExternalId() == "" {
					// It's possible that CV fails to launch the tryjob against
					// buildbucket and has missing external ID.
					continue
				}
				host, buildID, err := tryjob.ExternalID(executionAttempt.GetExternalId()).ParseBuildbucketID()
				if err != nil {
					return nil, err
				}
				origin := cvbqpb.Build_NOT_REUSED
				switch {
				case executionAttempt.GetReused():
					origin = cvbqpb.Build_REUSED
				case definition.GetDisableReuse():
					origin = cvbqpb.Build_NOT_REUSABLE
				}
				ret = append(ret, &cvbqpb.Build{
					Host:     host,
					Id:       buildID,
					Critical: definition.GetCritical(),
					Origin:   origin,
				})
			}
		}
		sort.Slice(ret, func(i, j int) bool {
			return ret[i].Id < ret[j].Id
		})
		return ret, nil
	}

	runTryjobs := r.Tryjobs.GetTryjobs()
	if len(runTryjobs) == 0 {
		return nil, nil
	}
	ret := make([]*cvbqpb.Build, len(runTryjobs))
	cg, err := prjcfg.GetConfigGroup(ctx, r.ID.LUCIProject(), r.ConfigGroupID)
	if err != nil {
		return nil, err
	}
	buildersCfg := cg.Content.GetVerifiers().GetTryjob().GetBuilders()
	builderCfgsByName := make(map[string]*cfgpb.Verifiers_Tryjob_Builder, len(buildersCfg))
	for _, builderCfg := range buildersCfg {
		builderCfgsByName[builderCfg.Name] = builderCfg
		// Associate the builder config with the equivalent name as well because
		// when CQDaemon reports Tryjobs, it just reports the launched builder name.
		// Therefore, if CQDaemon decides to launch the equivalent builder, the
		// definition stored in CV will be the equivalent builder instead of the
		// main builder.
		if equiName := builderCfg.GetEquivalentTo().GetName(); equiName != "" {
			builderCfgsByName[equiName] = builderCfg
		}
	}
	for i, tj := range runTryjobs {
		var err error
		b := &cvbqpb.Build{}
		if b.Host, b.Id, err = tryjob.ExternalID(tj.ExternalId).ParseBuildbucketID(); err != nil {
			return nil, err
		}
		builderName := bbBuilderNameFromDef(tj.GetDefinition())
		builderCfg, ok := builderCfgsByName[builderName]
		if !ok {
			logging.Warningf(ctx, "CQDaemon reported tryjob with builder \""+
				builderName+"\" that is not present in the ConfigGroup. This "+
				"may happen when builder is removed from the config during the Run")
		}
		switch {
		case tj.GetReused():
			b.Origin = cvbqpb.Build_REUSED
		case builderCfg.GetDisableReuse():
			b.Origin = cvbqpb.Build_NOT_REUSABLE
		default:
			b.Origin = cvbqpb.Build_NOT_REUSED
		}
		b.Critical = tj.GetCritical()
		ret[i] = b
	}
	return ret, nil
}

// bbBuilderNameFromDef returns Buildbucket builder name from Tryjob Definition.
//
// Returns the builder name in the format of "$project/$bucket/$builder".
// Panics for non-buildbucket backend.
func bbBuilderNameFromDef(def *tryjob.Definition) string {
	if def.GetBuildbucket() == nil {
		panic(fmt.Errorf("non-buildbucket backend is not supported; got %T", def.GetBackend()))
	}
	builder := def.GetBuildbucket().GetBuilder()
	return fmt.Sprintf("%s/%s/%s", builder.Project, builder.Bucket, builder.Builder)
}
