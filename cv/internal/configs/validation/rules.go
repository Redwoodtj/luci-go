// Copyright 2018 The LUCI Authors.
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

package validation

import (
	"context"
	"regexp"
	"time"

	"go.chromium.org/luci/common/data/caching/lru"
	"go.chromium.org/luci/config/validation"
)

// Config validation rules go here.

func init() {
	addRules(&validation.Rules)
}

// TODO(crbug.com/1252545): Use a dev-specific configuration for a dev instance
// of the service after CQD is deleted.
func addRules(r *validation.RuleSet) {
	r.Add("regex:projects/[^/]+", "commit-queue.cfg", validateProject)
	r.Add("services/commit-queue", "migration-settings.cfg", validateMigrationSettings)
	r.Add("services/${appid}", "listener-settings.cfg", validateListenerSettings)
}

// regexpCompileCached is the caching version of regexp.Compile.
//
// Most config files use the same regexp many times.
func regexpCompileCached(pattern string) (*regexp.Regexp, error) {
	cached, err := regexpCache.GetOrCreate(context.Background(), pattern, func() (interface{}, time.Duration, error) {
		r, err := regexp.Compile(pattern)
		return regexpCacheValue{r, err}, 0, nil
	})
	if err != nil {
		panic(err)
	}
	v := cached.(regexpCacheValue)
	return v.r, v.err
}

var regexpCache = lru.New(1024)

type regexpCacheValue struct {
	r   *regexp.Regexp
	err error
}

func enter(vctx *validation.Context, kind string, i int, name string) {
	if name == "" {
		vctx.Enter(kind+" #%d", i+1)
	} else {
		vctx.Enter(kind+" #%d %q", i+1, name)
	}
}
