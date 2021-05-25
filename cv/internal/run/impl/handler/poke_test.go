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
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/gae/service/datastore"

	cfgpb "go.chromium.org/luci/cv/api/config/v2"
	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/common"
	"go.chromium.org/luci/cv/internal/config"
	"go.chromium.org/luci/cv/internal/cvtesting"
	gf "go.chromium.org/luci/cv/internal/gerrit/gerritfake"
	"go.chromium.org/luci/cv/internal/run"
	"go.chromium.org/luci/cv/internal/run/impl/state"
	"go.chromium.org/luci/cv/internal/run/runtest"
	"go.chromium.org/luci/cv/internal/tree"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestPokeRecheckTree(t *testing.T) {
	t.Parallel()

	Convey("Poke", t, func() {
		ct := cvtesting.Test{}
		ctx, cancel := ct.SetUp()
		defer cancel()

		const lProject = "infra"
		rid := common.MakeRunID(lProject, ct.Clock.Now(), 1, []byte("deadbeef"))
		rs := &state.RunState{
			Run: run.Run{
				ID:         rid,
				CreateTime: ct.Clock.Now().UTC().Add(-2 * time.Minute),
				StartTime:  ct.Clock.Now().UTC().Add(-1 * time.Minute),
				CLs:        common.CLIDs{1},
			},
		}
		So(datastore.Put(ctx,
			&run.RunCL{
				ID:  1,
				Run: datastore.MakeKey(ctx, run.RunKind, string(rid)),
				Detail: &changelist.Snapshot{
					Kind: &changelist.Snapshot_Gerrit{
						Gerrit: &changelist.Gerrit{
							Host: "example.com",
							Info: gf.CI(1),
						},
					},
				},
			},
		), ShouldBeNil)
		h := &Impl{
			RM:         run.NewNotifier(ct.TQDispatcher),
			TreeClient: ct.TreeFake.Client(),
		}

		now := ct.Clock.Now()
		ctx = context.WithValue(ctx, &fakeTaskIDKey, "task-foo")
		Convey("Check Tree if condition matches", func() {
			rs.Run.Status = run.Status_WAITING_FOR_SUBMISSION
			rs.Run.Submission = &run.Submission{
				TreeOpen:          false,
				LastTreeCheckTime: timestamppb.New(now.Add(-1 * time.Minute)),
			}
			cfg := &cfgpb.Config{
				ConfigGroups: []*cfgpb.ConfigGroup{
					{
						Name: "main",
						Verifiers: &cfgpb.Verifiers{
							TreeStatus: &cfgpb.Verifiers_TreeStatus{
								Url: "tree.example.com",
							},
						},
					},
				},
			}
			ct.Cfg.Create(ctx, lProject, cfg)
			meta, err := config.GetLatestMeta(ctx, lProject)
			So(err, ShouldBeNil)
			So(meta.ConfigGroupIDs, ShouldHaveLength, 1)
			rs.Run.ConfigGroupID = meta.ConfigGroupIDs[0]

			Convey("Open", func() {
				res, err := h.Poke(ctx, rs)
				So(err, ShouldBeNil)
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldNotBeNil)
				// proceed to submission right away
				So(res.State.Run.Status, ShouldEqual, run.Status_SUBMITTING)
				So(res.State.Run.Submission, ShouldResembleProto, &run.Submission{
					Deadline:          timestamppb.New(now.Add(submissionDuration)),
					Cls:               []int64{1},
					TaskId:            "task-foo",
					TreeOpen:          true,
					LastTreeCheckTime: timestamppb.New(now),
				})
			})

			Convey("Close", func() {
				ct.TreeFake.ModifyState(ctx, tree.Closed)
				res, err := h.Poke(ctx, rs)
				So(err, ShouldBeNil)
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldBeNil)
				So(res.State.Run.Status, ShouldEqual, run.Status_WAITING_FOR_SUBMISSION)
				// record the result and check again after 1 minute.
				So(res.State.Run.Submission, ShouldResembleProto, &run.Submission{
					TreeOpen:          false,
					LastTreeCheckTime: timestamppb.New(now),
				})
				runtest.AssertReceivedPoke(ctx, rid, now.Add(1*time.Minute))
			})
		})

		Convey("No-op if condition doesn't match", func() {
			verifyNoOp := func() {
				res, err := h.Poke(ctx, rs)
				So(err, ShouldBeNil)
				So(res.State, ShouldEqual, rs)
				So(res.SideEffectFn, ShouldBeNil)
				So(res.PreserveEvents, ShouldBeFalse)
				So(res.PostProcessFn, ShouldBeNil)
			}
			Convey("Not in WAITING_FOR_SUBMISSION status", func() {
				rs.Run.Status = run.Status_RUNNING
				verifyNoOp()
			})

			Convey("Tree is open in the previous check", func() {
				rs.Run.Status = run.Status_WAITING_FOR_SUBMISSION
				rs.Run.Submission = &run.Submission{
					TreeOpen:          true,
					LastTreeCheckTime: timestamppb.New(now.Add(-2 * time.Minute)),
				}
				verifyNoOp()
			})

			Convey("Last Tree check is too recent", func() {
				rs.Run.Status = run.Status_WAITING_FOR_SUBMISSION
				rs.Run.Submission = &run.Submission{
					TreeOpen:          false,
					LastTreeCheckTime: timestamppb.New(now.Add(-1 * time.Second)),
				}
				verifyNoOp()
			})
		})
	})
}