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

package requirement

import (
	"context"
	"fmt"
	"regexp"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	cfgpb "go.chromium.org/luci/cv/api/config/v2"
	"go.chromium.org/luci/cv/internal/changelist"
	"go.chromium.org/luci/cv/internal/run"
)

// locationFilterMatch returns true if a builder is included, given
// the location filters and CLs (and their file paths).
func locationFilterMatch(ctx context.Context, locationFilters []*cfgpb.Verifiers_Tryjob_Builder_LocationFilter, cls []*run.RunCL) (bool, error) {
	if len(locationFilters) == 0 {
		// If there are no location filters, the builder is included.
		return true, nil
	}

	// For efficiency, pre-compile all regexes here. This also checks whether
	// the regexes are valid.
	compiled, err := compileLocationFilters(ctx, locationFilters)
	if err != nil {
		return false, err
	}

	for _, cl := range cls {
		gerrit := cl.Detail.GetGerrit()
		if gerrit == nil {
			// Could result from error or if there is an non-Gerrit backend.
			return false, errors.New("empty Gerrit detail")
		}
		host := gerrit.GetHost()
		project := gerrit.GetInfo().GetProject()

		if isMergeCommit(ctx, gerrit) {
			if hostAndProjectMatch(compiled, host, project) {
				// Gerrit treats CLs representing merged commits (i.e. CLs with with a
				// git commit with multiple parents) as having no file diff. There may
				// also be no file diff if there is no longer a diff after rebase.
				// For merge commits, we want to avoid inadvertently landing
				// such a CLs without triggering any builders.
				//
				// If there is a CL which is a merge commit, and the builder would
				// be triggered for some files in that repo, then trigger the builder.
				// See crbug/1006534.
				return true, nil
			}
			continue
		}
		// Iterate through all files to try to find a match.
		// If there are no files, but this is not a merge commit, then do
		// nothing for this CL.
		for _, path := range gerrit.GetFiles() {
			// If the first filter is an exclude filter, then include by default, and
			// vice versa.
			included := locationFilters[0].Exclude
			// Whether the file is included is determined by the last filter to match.
			// So we can iterate through the filters backwards and break when we have
			// a match.
			for i := len(compiled) - 1; i >= 0; i-- {
				f := compiled[i]
				// Check for inclusion; if it matches then this is the filter
				// that applies.
				if match(f.hostRE, host) && match(f.projectRE, project) && match(f.pathRE, path) {
					included = !f.exclude
					break
				}
			}
			// If at least one file in one CL is included, then the builder is included.
			if included {
				return true, nil
			}
		}
	}

	// After looping through all files in all CLs, all were considered
	// excluded, so the builder should not be triggered.
	return false, nil
}

// match is like re.MatchString, but also matches if the regex is nil.
func match(re *regexp.Regexp, str string) bool {
	return re == nil || re.MatchString(str)
}

// compileLocationFilters precompiles regexes in a LocationFilter.
//
// Returns an error if a regex is invalid.
func compileLocationFilters(ctx context.Context, locationFilters []*cfgpb.Verifiers_Tryjob_Builder_LocationFilter) ([]compiledLocationFilter, error) {
	ret := make([]compiledLocationFilter, len(locationFilters))
	for i, lf := range locationFilters {
		var err error
		if lf.GerritHostRegexp != "" {
			ret[i].hostRE, err = regexp.Compile(fmt.Sprintf("^%s$", lf.GerritHostRegexp))
			if err != nil {
				return nil, err
			}
		}
		if lf.GerritProjectRegexp != "" {
			ret[i].projectRE, err = regexp.Compile(fmt.Sprintf("^%s$", lf.GerritProjectRegexp))
			if err != nil {
				return nil, err
			}
		}
		if lf.PathRegexp != "" {
			ret[i].pathRE, err = regexp.Compile(fmt.Sprintf("^%s$", lf.PathRegexp))
			if err != nil {
				return nil, err
			}
		}
		ret[i].exclude = lf.Exclude
	}
	return ret, nil
}

// compiledLocationFilter stores the same information as the LocationFilter
// message in the config proto, but with compiled regexes.
type compiledLocationFilter struct {
	// Compiled regexes; nil if the regex is nil in the LocationFilter.
	hostRE, projectRE, pathRE *regexp.Regexp
	// Whether this filter is an exclude filter.
	exclude bool
}

// hostAndProjectMatch returns true if the Gerrit host and project could match
// the filters (for any possible files).
func hostAndProjectMatch(compiled []compiledLocationFilter, host, project string) bool {
	for _, f := range compiled {
		if !f.exclude && match(f.hostRE, host) && match(f.projectRE, project) {
			return true
		}
	}
	// If the first filter is an exclude filter, we include by default;
	// exclude by default.
	return compiled[0].exclude
}

// isMergeCommit checks whether the current revision of the change is a merge
// commit based on available Gerrit information.
func isMergeCommit(ctx context.Context, g *changelist.Gerrit) bool {
	i := g.GetInfo()
	rev, ok := i.GetRevisions()[i.GetCurrentRevision()]
	if !ok {
		logging.Errorf(ctx, "No current revision in ChangeInfo when checking isMergeCommit, got %+v", i)
		return false
	}
	return len(rev.GetCommit().GetParents()) > 1 && len(g.GetFiles()) == 0
}
