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

package aggrmetrics

import (
	"testing"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/cvtesting"
	"go.chromium.org/luci/cv/internal/prjmanager"
	"go.chromium.org/luci/cv/internal/prjmanager/prjpb"
	"go.chromium.org/luci/cv/internal/run"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPMReporter(t *testing.T) {
	t.Parallel()

	Convey("pmReporter works", t, func() {
		ct := cvtesting.Test{}
		ctx, cancel := ct.SetUp()
		defer cancel()

		So(datastore.Put(ctx,
			&prjmanager.Project{
				ID:         "small",
				EVersion:   1,
				UpdateTime: ct.Clock.Now().UTC(),
				State: &prjpb.PState{
					Status: prjpb.Status_STARTED,
					Pcls: []*prjpb.PCL{
						{Clid: 1, Eversion: 1, Submitted: true},
						{Clid: 2, Eversion: 2, Deps: []*changelist.Dep{{Clid: 1, Kind: changelist.DepKind_HARD}},
							Trigger: &run.Trigger{Mode: "DryRun", Time: timestamppb.New(ct.Clock.Now())},
						},
					},
				},
			},
			&prjmanager.Project{
				ID:         "empty",
				EVersion:   1,
				UpdateTime: ct.Clock.Now().UTC(),
				State: &prjpb.PState{
					Status: prjpb.Status_STARTED,
				},
			},
		), ShouldBeNil)

		r := pmReporter{}
		reportFunc, err := r.prepare(ctx, stringset.NewFromSlice("small", "empty"))
		So(err, ShouldBeNil)
		reportFunc(ctx)
		So(ct.TSMonSentValue(ctx, metricPMEntitySize, "small"), ShouldBeBetween, 80, 90)
		So(ct.TSMonSentValue(ctx, metricPMEntitySize, "empty"), ShouldBeBetween, 40, 50)
	})
}