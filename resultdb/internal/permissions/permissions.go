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

package permissions

import (
	"context"

	"google.golang.org/grpc/codes"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/trace"
	"go.chromium.org/luci/grpc/appstatus"
	"go.chromium.org/luci/resultdb/internal/invocations"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/realms"
	"go.chromium.org/luci/server/span"
)

// VerifyInvocation checks if the caller has the specified permissions on the
// realm that the invocation with the specified id belongs to.
// There must not already be a transaction in the given context.
func VerifyInvocation(ctx context.Context, id invocations.ID, permissions ...realms.Permission) error {
	return VerifyInvocations(ctx, invocations.NewIDSet(id), permissions...)
}

// VerifyInvocations is checks multiple invocations' realms for the specified
// permissions.
// There must not already be a transaction in the given context.
func VerifyInvocations(ctx context.Context, ids invocations.IDSet, permissions ...realms.Permission) (err error) {
	if len(ids) == 0 {
		return nil
	}
	ctx, ts := trace.StartSpan(ctx, "resultdb.permissions.VerifyInvocations")
	defer func() { ts.End(err) }()

	realms, err := invocations.ReadRealms(span.Single(ctx), ids)
	if err != nil {
		return err
	}

	checked := stringset.New(1)
	for id, realm := range realms {
		if !checked.Add(realm) {
			continue
		}
		// Note: HasPermission does not make RPCs.
		for _, permission := range permissions {
			switch allowed, err := auth.HasPermission(ctx, permission, realm, nil); {
			case err != nil:
				return err
			case !allowed:
				return appstatus.Errorf(codes.PermissionDenied, `caller does not have permission %s in realm of invocation %s`, permission, id)
			}
		}
	}
	return nil
}

// VerifyInvocationsByName does the same as VerifyInvocations but accepts
// an invocation name instead of an invocations.ID.
// There must not already be a transaction in the given context.
func VerifyInvocationsByName(ctx context.Context, invNames []string, permissions ...realms.Permission) error {
	ids, err := invocations.ParseNames(invNames)
	if err != nil {
		return appstatus.BadRequest(err)
	}
	return VerifyInvocations(ctx, ids, permissions...)
}

// VerifyInvocationByName does the same as VerifyInvocation but accepts
// invocation names instead of an invocations.IDSet.
// There must not already be a transaction in the given context.
func VerifyInvocationByName(ctx context.Context, invName string, permissions ...realms.Permission) error {
	return VerifyInvocationsByName(ctx, []string{invName}, permissions...)
}
