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

package buildcron

import (
	"context"

	"go.chromium.org/luci/buildbucket/appengine/model"
)

// TimeoutExpiredBuilds marks incomplete builds that were created longer than
// model.BuildMaxCompletionTime w/ INFRA_FAILURE.
func TimeoutExpiredBuilds(ctx context.Context) error {
	// TODO: implement me
	_ = model.BuildMaxCompletionTime
	return nil
}

// ResetExpiredLeases resets expired leases.
func ResetExpiredLeases(ctx context.Context) error {
	// TODO: implement me
	return nil
}