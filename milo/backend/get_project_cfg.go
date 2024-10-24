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

package backend

import (
	"context"

	"go.chromium.org/luci/auth/identity"
	"go.chromium.org/luci/grpc/appstatus"
	milo "go.chromium.org/luci/milo/api/config"
	milopb "go.chromium.org/luci/milo/api/service/v1"
	"go.chromium.org/luci/milo/common"
	"go.chromium.org/luci/server/auth"
	"google.golang.org/grpc/codes"
)

// GetProjectCfg implements milopb.MiloInternal service
func (s *MiloInternalService) GetProjectCfg(ctx context.Context, req *milopb.GetProjectCfgRequest) (_ *milo.Project, err error) {
	projectName := req.GetProject()
	if projectName == "" {
		return nil, appstatus.Error(codes.InvalidArgument, "project must be specified")
	}

	allowed, err := common.IsAllowed(ctx, projectName)
	if err != nil {
		return nil, err
	}
	if !allowed {
		if auth.CurrentIdentity(ctx) == identity.AnonymousIdentity {
			return nil, appstatus.Error(codes.Unauthenticated, "not logged in ")
		}
		return nil, appstatus.Error(codes.PermissionDenied, "no access to the project")
	}

	project, err := common.GetProject(ctx, projectName)
	if err != nil {
		return nil, err
	}
	return &milo.Project{
		LogoUrl:        project.LogoURL,
		BugUrlTemplate: project.BugURLTemplate,
	}, nil
}
