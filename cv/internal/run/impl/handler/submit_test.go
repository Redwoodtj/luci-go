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

package handler

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/types/known/timestamppb"

	cfgpb "go.chromium.org/luci/cv/api/config/v2"
	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/common"
	"go.chromium.org/luci/cv/internal/config"
	"go.chromium.org/luci/cv/internal/cvtesting"
	gf "go.chromium.org/luci/cv/internal/gerrit/gerritfake"
	"go.chromium.org/luci/cv/internal/run"
	"go.chromium.org/luci/cv/internal/run/impl/state"
	"go.chromium.org/luci/cv/internal/run/impl/submit"
	"go.chromium.org/luci/cv/internal/run/runtest"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestOnReadyForSubmission(t *testing.T) {
	t.Parallel()

	Convey("OnReadyForSubmission", t, func() {
		ct := cvtesting.Test{}
		ctx, cancel := ct.SetUp()
		defer cancel()
		rid := common.MakeRunID("infra", ct.Clock.Now(), 1, []byte("deadbeef"))
		runCLs := common.CLIDs{1, 2}
		r := run.Run{
			ID:         rid,
			Status:     run.Status_RUNNING,
			CreateTime: ct.Clock.Now().UTC().Add(-2 * time.Minute),
			StartTime:  ct.Clock.Now().UTC().Add(-1 * time.Minute),
			CLs:        runCLs,
		}
		ct.Cfg.Create(ctx, rid.LUCIProject(), &cfgpb.Config{
			ConfigGroups: []*cfgpb.ConfigGroup{
				{Name: "main"},
			},
		})
		meta, err := config.GetLatestMeta(ctx, rid.LUCIProject())
		So(err, ShouldBeNil)
		So(meta.ConfigGroupIDs, ShouldHaveLength, 1)
		r.ConfigGroupID = meta.ConfigGroupIDs[0]
		So(datastore.Put(ctx, &r,
			&run.RunCL{
				ID:  runCLs[0],
				Run: datastore.MakeKey(ctx, run.RunKind, string(rid)),
				Detail: &changelist.Snapshot{
					Kind: &changelist.Snapshot_Gerrit{
						Gerrit: &changelist.Gerrit{
							Host: "example.com",
							Info: gf.CI(1111),
						},
					},
					Deps: []*changelist.Dep{
						{Clid: 2, Kind: changelist.DepKind_HARD},
					},
				},
			},
			&run.RunCL{
				ID:  runCLs[1],
				Run: datastore.MakeKey(ctx, run.RunKind, string(rid)),
				Detail: &changelist.Snapshot{
					Kind: &changelist.Snapshot_Gerrit{
						Gerrit: &changelist.Gerrit{
							Host: "example.com",
							Info: gf.CI(2222),
						},
					},
				},
			},
		), ShouldBeNil)
		rs := &state.RunState{Run: r}

		h := &Impl{}

		statuses := []run.Status{
			run.Status_SUCCEEDED,
			run.Status_FAILED,
			run.Status_CANCELLED,
		}
		for _, status := range statuses {
			Convey(fmt.Sprintf("Release submit queue when Run is %s", status), func() {
				So(datastore.RunInTransaction(ctx, func(ctx context.Context) error {
					waitlisted, err := submit.TryAcquire(ctx, rs.Run.ID, nil)
					So(waitlisted, ShouldBeFalse)
					return err
				}, nil), ShouldBeNil)
				rs.Run.Status = status
				res, err := h.OnReadyForSubmission(ctx, rs)
				So(err, ShouldBeNil)
				So(res.State, ShouldEqual, rs)
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldBeNil)
				current, waitlist, err := submit.LoadCurrentAndWaitlist(ctx, rs.Run.ID)
				So(err, ShouldBeNil)
				So(current, ShouldBeEmpty)
				So(waitlist, ShouldBeEmpty)
			})
		}

		now := ct.Clock.Now().UTC()
		ctx, cancel = clock.WithDeadline(ctx, now.Add(1*time.Minute))
		defer cancel()
		Convey("When status is SUBMITTING", func() {
			rs.Run.Status = run.Status_SUBMITTING

			Convey("Sends Poke if within deadline", func() {
				rs.Run.Submission = &run.Submission{
					Deadline:     timestamppb.New(now.Add(30 * time.Second)), // with in deadline
					AttemptCount: 1,
				}
				res, err := h.OnReadyForSubmission(ctx, rs)
				So(err, ShouldBeNil)
				So(res.State, ShouldEqual, rs)
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldBeNil)
				runtest.AssertReceivedPoke(ctx, rs.Run.ID, rs.Run.Submission.Deadline.AsTime())
			})

			Convey("Re-acquire submit queue if deadline is exceeded", func() {
				rs.Run.Submission = &run.Submission{
					Deadline:     timestamppb.New(now.Add(-30 * time.Second)), // passed deadline
					AttemptCount: 1,
				}

				Convey("And if waitlisted, fall back to WAITING_FOR_SUBMISSION status", func() {
					// submit queue is taken by another run.
					So(datastore.RunInTransaction(ctx, func(ctx context.Context) error {
						waitlisted, err := submit.TryAcquire(ctx, common.MakeRunID("infra", now, 1, []byte("another-run")), nil)
						So(waitlisted, ShouldBeFalse)
						return err
					}, nil), ShouldBeNil)
					res, err := h.OnReadyForSubmission(ctx, rs)
					So(err, ShouldBeNil)
					So(res.State.Run.Status, ShouldEqual, run.Status_WAITING_FOR_SUBMISSION)
					So(res.State.Run.Submission.Deadline, ShouldBeNil)
					So(res.SideEffectFn, ShouldBeNil)
					So(res.PreserveEvents, ShouldBeFalse)
					So(res.PostProcessFn, ShouldBeNil)
				})

				Convey("And if not waitlisted, try submitting again", func() {
					res, err := h.OnReadyForSubmission(ctx, rs)
					So(err, ShouldBeNil)
					So(res.State.Run.Status, ShouldEqual, run.Status_SUBMITTING)
					So(res.State.Run.Submission.Deadline, ShouldResembleProto, timestamppb.New(now.Add(1*time.Minute))) // set to ctx deadline
					So(res.State.Run.Submission.AttemptCount, ShouldEqual, 2)
					So(res.SideEffectFn, ShouldBeNil)
					So(res.PreserveEvents, ShouldBeFalse)
					So(res.PostProcessFn, ShouldNotBeNil)
					// event sent when successfully acquiring the submit queue.
					runtest.AssertReceivedReadyForSubmission(ctx, rs.Run.ID, now.Add(10*time.Second))
				})
			})
		})

		for _, status := range []run.Status{run.Status_RUNNING, run.Status_WAITING_FOR_SUBMISSION} {
			Convey(fmt.Sprintf("Mark submitting when status is %s", status), func() {
				rs.Run.Status = status
				res, err := h.OnReadyForSubmission(ctx, rs)
				So(err, ShouldBeNil)
				So(res.State.Run.Status, ShouldEqual, run.Status_SUBMITTING)
				So(res.State.Run.Submission, ShouldResembleProto, &run.Submission{
					Deadline:     timestamppb.New(now.Add(1 * time.Minute)), // use deadline in ctx
					AttemptCount: 1,
					Cls:          []int64{2, 1}, // in submission order
				})
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldNotBeNil)
			})
		}
	})
}