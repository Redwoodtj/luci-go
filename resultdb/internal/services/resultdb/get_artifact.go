// Copyright 2020 The LUCI Authors.
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

package resultdb

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/resultdb/internal/artifacts"
	"go.chromium.org/luci/resultdb/internal/invocations"
	"go.chromium.org/luci/resultdb/internal/permissions"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/v1"
	"go.chromium.org/luci/resultdb/rdbperms"
	"go.chromium.org/luci/server/span"
)

func verifyReadArtifactPermission(ctx context.Context, name string) error {
	invIDStr, _, _, _, inputErr := pbutil.ParseArtifactName(name)
	if inputErr != nil {
		return appstatus.BadRequest(inputErr)
	}

	return permissions.VerifyInvocation(ctx, invocations.ID(invIDStr), rdbperms.PermGetArtifact)
}

func validateGetArtifactRequest(req *pb.GetArtifactRequest) error {
	if err := pbutil.ValidateArtifactName(req.Name); err != nil {
		return errors.Annotate(err, "name").Err()
	}

	return nil
}

// GetArtifact implements pb.ResultDBServer.
func (s *resultDBServer) GetArtifact(ctx context.Context, in *pb.GetArtifactRequest) (*pb.Artifact, error) {
	if err := verifyReadArtifactPermission(ctx, in.Name); err != nil {
		return nil, err
	}

	if err := validateGetArtifactRequest(in); err != nil {
		return nil, appstatus.BadRequest(err)
	}

	ctx, cancel := span.ReadOnlyTransaction(ctx)
	defer cancel()

	art, err := artifacts.Read(ctx, in.Name)
	if err != nil {
		return nil, err
	}

	if err := s.populateFetchURLs(ctx, art); err != nil {
		return nil, err
	}

	return art, nil
}

// populateFetchURLs populates FetchUrl and FetchUrlExpiration fields
// of the artifacts.
//
// Must be called from within some gRPC request handler.
func (s *resultDBServer) populateFetchURLs(ctx context.Context, artifacts ...*pb.Artifact) error {
	// Extract Host header (may be empty) from the request to use it as a basis
	// for generating artifact URLs.
	requestHost := ""
	md, _ := metadata.FromIncomingContext(ctx)
	if val := md.Get("host"); len(val) > 0 {
		requestHost = val[0]
	}

	for _, a := range artifacts {

		if a.GcsUri != "" {
			now := clock.Now(ctx).UTC()

			// TODO: Validate whether the path belongs to the realm and change this to
			// use signed url instead.
			// Currently return the GCS url directly (gated by the bucket ACL)
			url, exp := strings.Replace(a.GcsUri, "gs://",
				"https://console.developers.google.com/storage/browser/", 1), now.Add(7*24*time.Hour)

			a.FetchUrl = url
			a.FetchUrlExpiration = pbutil.MustTimestampProto(exp)
		} else {
			url, exp, err := s.generateArtifactURL(ctx, requestHost, a.Name)
			if err != nil {
				return err
			}

			a.FetchUrl = url
			a.FetchUrlExpiration = pbutil.MustTimestampProto(exp)
		}
	}
	return nil
}
