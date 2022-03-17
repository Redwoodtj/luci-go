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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

const (
	ensureFileHeader = "$ServiceURL https://chrome-infra-packages.appspot.com/\n$ParanoidMode CheckPresence\n"
	kitchenCheckout  = "kitchen-checkout"
)

// resultsFilePath is the path to the generated file from cipd ensure command.
// Placing it here to allow to replace it during testing.
var resultsFilePath = filepath.Join(os.TempDir(), "cipd_ensure_results.json")

// execCommandContext to allow to replace it during testing.
var execCommandContext = exec.CommandContext

type cipdPkg struct {
	Package    string `json:"package"`
	InstanceID string `json:"instance_id"`
}

// cipdOut corresponds to the structure of the generated result json file from cipd ensure command.
type cipdOut struct {
	Result map[string][]*cipdPkg `json:"result"`
}

// installCipdPackages installs cipd packages defined in build.Infra.Buildbucket.Agent.Input
// and build exe. It also prepends desired value to $PATH env var.
//
// Note: It assumes `cipd` client tool binary is already in path.
func installCipdPackages(ctx context.Context, build *bbpb.Build, workDir string) (map[string]*bbpb.ResolvedDataRef, error) {
	logging.Infof(ctx, "Installing cipd packages into %s", workDir)
	inputData := build.Infra.Buildbucket.Agent.Input.Data

	ensureFileBuilder := strings.Builder{}
	ensureFileBuilder.WriteString(ensureFileHeader)
	extraPathEnv := stringset.Set{}
	for dir, pkgs := range inputData {
		if pkgs.GetCipd() == nil {
			continue
		}
		extraPathEnv.AddAll(pkgs.OnPath)
		fmt.Fprintf(&ensureFileBuilder, "@Subdir %s\n", dir)
		for _, spec := range pkgs.GetCipd().Specs {
			fmt.Fprintf(&ensureFileBuilder, "%s %s\n", spec.Package, spec.Version)
		}
	}
	if build.Exe != nil {
		fmt.Fprintf(&ensureFileBuilder, "@Subdir %s\n", kitchenCheckout)
		fmt.Fprintf(&ensureFileBuilder, "%s %s\n", build.Exe.CipdPackage, build.Exe.CipdVersion)
	}

	// TODO(crbug.com/1297809): Remove this redundant log once this feature development is done.
	logging.Infof(ctx, "===ensure file===\n%s\n=========", ensureFileBuilder.String())

	// Install packages
	cmd := execCommandContext(ctx, "cipd", "ensure", "-root", workDir, "-ensure-file", "-", "-json-output", resultsFilePath)
	cmd.Stdin = strings.NewReader(ensureFileBuilder.String())
	logging.Infof(ctx, "Running command: %s", cmd.String())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Annotate(err, "Errors in running command: %s\nOutput: %s", cmd.String(), string(out)).Err()
	}
	logging.Debugf(ctx, "Output of the cipd ensure command:\n%s", string(out))

	resultsFile, err := os.Open(resultsFilePath)
	if err != nil {
		return nil, err
	}
	defer resultsFile.Close()
	cipdOutputs := cipdOut{}
	jsonResults, err := ioutil.ReadAll(resultsFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonResults, &cipdOutputs); err != nil {
		return nil, err
	}

	// Prepend to $PATH
	original := os.Getenv("PATH")
	if err := os.Setenv("PATH", strings.Join(append(extraPathEnv.ToSlice(), original), string(os.PathListSeparator))); err != nil {
		return nil, err
	}

	resolvedDataMap := map[string]*bbpb.ResolvedDataRef{}
	for p, pkgs := range cipdOutputs.Result {
		resolvedPkgs := make([]*bbpb.ResolvedDataRef_CIPD_PkgSpec, 0, len(pkgs))
		for _, pkg := range pkgs {
			resolvedPkgs = append(resolvedPkgs, &bbpb.ResolvedDataRef_CIPD_PkgSpec{
				Package: pkg.Package,
				Version: pkg.InstanceID,
			})
		}
		resolvedDataMap[p] = &bbpb.ResolvedDataRef{
			DataType: &bbpb.ResolvedDataRef_Cipd{Cipd: &bbpb.ResolvedDataRef_CIPD{
				Specs: resolvedPkgs,
			}},
		}
	}

	return resolvedDataMap, nil
}