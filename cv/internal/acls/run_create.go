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

package acls

import (
	"context"
	"strings"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/server/auth"

	"go.chromium.org/luci/cv/internal/configs/prjcfg"
	"go.chromium.org/luci/cv/internal/run"

	cfgpb "go.chromium.org/luci/cv/api/config/v2"
)

// CheckResult tells the result of an ACL check performed.
type CheckResult struct {
	// FailuresSummary is a summary of the check failures with the reasons.
	//
	// Provides a human friendly summary of the reasons for the decision
	// when OK == false.
	// Empty if OK == true.
	FailuresSummary string
}

// OK returns true if the result indicates no failures. False, otherwise.
func (res CheckResult) OK() bool {
	return res.FailuresSummary == ""
}

type runCreateFailures struct {
	neitherCommitterNorOwner []*run.RunCL
}

func (fs *runCreateFailures) length() int {
	return len(fs.neitherCommitterNorOwner)
}

func (fs *runCreateFailures) summary() string {
	if fs.length() == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("CV run can't continue due to the following CLs\n\n")
	if cls := fs.neitherCommitterNorOwner; len(cls) > 0 {
		sb.WriteString("* only the full committers or CL owner can trigger runs.\n")
		for _, cl := range cls {
			sb.WriteString(cl.ExternalID.MustURL())
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// CheckRunCreate verifies that the user(s) who triggered Run are authorized
// to create the Run for the CLs.
func CheckRunCreate(ctx context.Context, cg *prjcfg.ConfigGroup, cls []*run.RunCL) (CheckResult, error) {
	failures := &runCreateFailures{}
	for _, cl := range cls {
		triggerer, err := identity.MakeIdentity("user:" + cl.Trigger.Email)
		if err != nil {
			return CheckResult{}, errors.Annotate(
				err, "the triggerer identity %q of CL %q is invalid", cl.Trigger.Email, cl.ID).Err()
		}

		switch yes, err := isCommitter(ctx, triggerer, cg.Content.Verifiers); {
		case err != nil:
			return CheckResult{}, errors.Annotate(err, "failed to check committer").Err()
		case !yes:
			// Non-committer must be CL owner.
			owner, err := cl.Detail.OwnerIdentity()
			if err != nil {
				return CheckResult{}, errors.Annotate(
					err, "the owner identity of CL %q is invalid", cl.ID).Err()
			}
			if triggerer != owner {
				failures.neitherCommitterNorOwner = append(failures.neitherCommitterNorOwner, cl)
			}
		}
	}
	return CheckResult{failures.summary()}, nil
}

func isCommitter(ctx context.Context, one identity.Identity, v *cfgpb.Verifiers) (bool, error) {
	if groups := v.GetGerritCqAbility().GetCommitterList(); len(groups) > 0 {
		return auth.GetState(ctx).DB().IsMember(ctx, one, groups)
	}
	return false, nil
}