// Copyright 2022 The LUCI Authors.
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

package tasks

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/buildbucket/appengine/model"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestBackendTask(t *testing.T) {
	ctx := context.Background()
	build := &model.Build{}

	Convey("assert createBackendTask", t, func() {
		err := createBackendTask(ctx, build)
		So(err, ShouldErrLike, "Method not implemented")
	})
}
