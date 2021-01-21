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

package migration

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/retry/transient"
	"go.chromium.org/luci/common/sync/parallel"
	"go.chromium.org/luci/gae/service/datastore"

	cvbqpb "go.chromium.org/luci/cv/api/bigquery/v1"
	migrationpb "go.chromium.org/luci/cv/api/migration"
	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/common"
	"go.chromium.org/luci/cv/internal/prjmanager"
	"go.chromium.org/luci/cv/internal/run"
	runImpl "go.chromium.org/luci/cv/internal/run/impl"
)

func fetchActiveRuns(ctx context.Context, project string) ([]*migrationpb.Run, error) {
	runs := []run.Run{}
	q := run.NewQueryWithLUCIProject(ctx, project).Eq("Status", run.Status_RUNNING)
	if err := datastore.GetAll(ctx, q, &runs); err != nil {
		return nil, errors.Annotate(err, "fetch Run entities").Tag(transient.Tag).Err()
	}
	numRuns := len(runs)
	if numRuns == 0 {
		return nil, nil
	}
	poolSize := numRuns
	if poolSize > 20 {
		poolSize = 20
	}
	ret := make([]*migrationpb.Run, numRuns)
	err := parallel.WorkPool(poolSize, func(workCh chan<- func() error) {
		for i, r := range runs {
			i, r := i, r
			workCh <- func() error {
				runKey := datastore.MakeKey(ctx, run.RunKind, string(r.ID))
				runCLs := make([]run.RunCL, len(r.CLs))
				for i, cl := range r.CLs {
					runCLs[i] = run.RunCL{
						ID:  cl,
						Run: runKey,
					}
				}
				if err := datastore.Get(ctx, runCLs); err != nil {
					return errors.Annotate(err, "fetch CLs for run %q", r.ID).Tag(transient.Tag).Err()
				}
				mcls := make([]*migrationpb.RunCL, len(runCLs))
				mode := cvbqpb.Mode_FULL_RUN
				if r.Mode == run.DryRun {
					mode = cvbqpb.Mode_DRY_RUN
				}
				for i, cl := range runCLs {
					trigger := &migrationpb.RunCL_Trigger{
						Email:     cl.Trigger.GetEmail(),
						Time:      cl.Trigger.GetTime(),
						AccountId: cl.Trigger.GetGerritAccountId(),
					}
					mcl := &migrationpb.RunCL{
						Id: int64(cl.ID),
						Gc: &cvbqpb.GerritChange{
							Host:                       cl.Detail.GetGerrit().GetHost(),
							Project:                    cl.Detail.GetGerrit().GetInfo().GetProject(),
							Change:                     cl.Detail.GetGerrit().GetInfo().GetNumber(),
							Patchset:                   int64(cl.Detail.GetPatchset()),
							EarliestEquivalentPatchset: int64(cl.Detail.GetMinEquivalentPatchset()),
							Mode:                       mode,
						},
						Files:   cl.Detail.GetGerrit().GetFiles(),
						Info:    cl.Detail.GetGerrit().GetInfo(),
						Trigger: trigger,
						Deps:    make([]*migrationpb.RunCL_Dep, len(cl.Detail.GetDeps())),
					}
					for i, dep := range cl.Detail.GetDeps() {
						mcl.Deps[i] = &migrationpb.RunCL_Dep{
							Id: dep.GetClid(),
						}
						if dep.GetKind() == changelist.DepKind_HARD {
							mcl.Deps[i].Hard = true
						}
					}
					mcls[i] = mcl
				}
				ret[i] = &migrationpb.Run{
					Attempt: &cvbqpb.Attempt{
						LuciProject: project,
					},
					Id:  string(r.ID),
					Cls: mcls,
				}
				return nil
			}
		}
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

var terminalStatusMapping = map[cvbqpb.AttemptStatus]run.Status{
	cvbqpb.AttemptStatus_SUCCESS:       run.Status_SUCCEEDED,
	cvbqpb.AttemptStatus_FAILURE:       run.Status_FAILED,
	cvbqpb.AttemptStatus_INFRA_FAILURE: run.Status_FAILED,
	cvbqpb.AttemptStatus_ABORTED:       run.Status_CANCELLED,
}

// finalizeRun updates the Run in CV with the reported finished Run from
// CQDaemon and then notifies Project Manager.
func finalizeRun(ctx context.Context, mr *migrationpb.Run) error {
	attempt := mr.GetAttempt()
	terminalStatus, ok := terminalStatusMapping[attempt.GetStatus()]
	if !ok {
		return errors.Reason("expected terminal status for attempt %q; got %s", attempt.GetKey(), attempt.GetStatus()).Err()
	}
	err := datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		r := &run.Run{ID: common.RunID(mr.Id)}
		switch err := datastore.Get(ctx, r); {
		case err == datastore.ErrNoSuchEntity:
			return errors.Reason("run doesn't exist in CV").Err()
		case err != nil:
			return errors.Annotate(err, "failed to fetch Run").Tag(transient.Tag).Err()
		}
		if run.IsEnded(r.Status) {
			logging.Warningf(ctx, "Run %q has already ended", r.ID)
			return nil
		}
		r.Mutate(ctx, func(r *run.Run) (updated bool) {
			r.Status = terminalStatus
			r.EndTime = attempt.GetEndTime().AsTime()
			return true
		})
		// TODO(yiwzhang): Store rest of the data (e.g. tryjobs) reported by
		// CQDaemon. This will help CV accumulate more historical data.
		if err := datastore.Put(ctx, r); err != nil {
			return errors.Annotate(err, "failed to put Run").Tag(transient.Tag).Err()
		}
		if err := runImpl.RemoveRunFromCLs(ctx, r); err != nil {
			return err
		}
		return prjmanager.NotifyRunFinished(ctx, r.ID)
	}, nil)
	if err != nil {
		return err
	}
	return nil
}