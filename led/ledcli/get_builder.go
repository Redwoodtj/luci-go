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

package ledcli

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/buildbucket/protoutil"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/flag/stringlistflag"
	"go.chromium.org/luci/led/job"
	"go.chromium.org/luci/led/ledcmd"
)

func getBuilderCmd(opts cmdBaseOptions) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: "get-builder bucket_name:builder_name or project/bucket/builder",
		ShortDesc: "obtain a JobDefinition from a buildbucket builder",
		LongDesc:  `Obtains the builder definition from buildbucket and produces a JobDefinition.`,

		CommandRun: func() subcommands.CommandRun {
			ret := &cmdGetBuilder{}
			ret.initFlags(opts)
			return ret
		},
	}
}

type cmdGetBuilder struct {
	cmdBase

	tags         stringlistflag.Flag
	bbHost       string
	canary       bool
	priorityDiff int
	realBuild    bool

	project string
	bucket  string
	builder string
}

func (c *cmdGetBuilder) initFlags(opts cmdBaseOptions) {
	c.Flags.Var(&c.tags, "t",
		"(repeatable) set tags for this build. Buildbucket expects these to be `key:value`.")
	c.Flags.StringVar(&c.bbHost, "B", "cr-buildbucket.appspot.com",
		"The buildbucket hostname to grab the definition from.")
	c.Flags.BoolVar(&c.canary, "canary", false,
		"Get a 'canary' build, rather than a 'prod' build.")
	c.Flags.IntVar(&c.priorityDiff, "adjust-priority", 10,
		"Increase or decrease the priority of the generated job. Note: priority works like Unix 'niceness'; Higher values indicate lower priority.")
	c.Flags.BoolVar(&c.realBuild, "real-build", false,
		"Get a synthesized build for the builder, instead of the swarmbucket template.")
	c.cmdBase.initFlags(opts)
}

func (c *cmdGetBuilder) jobInput() bool                  { return false }
func (c *cmdGetBuilder) positionalRange() (min, max int) { return 1, 1 }

type builder struct {
	project  string
	v1Bucket string
	v2Bucket string
	builder  string
}

// parseBuilder parses the builder string in the format of "luci.project.bucket:builder".
func parseV1Builder(builderStr string) *builder {
	v1BuilderRe := regexp.MustCompile(`^(luci\.(\w*)\.(\w*)):(.*)$`)
	match := v1BuilderRe.FindStringSubmatch(builderStr)
	if len(match) != 5 {
		return nil
	}
	return &builder{
		project:  match[2],
		v1Bucket: match[1],
		v2Bucket: match[3],
		builder:  match[4],
	}
}

// parseBuilder parses the builder string in the format of "project/bucket/builder".
func parseV2Builder(builderStr string) *builder {
	builderID, err := protoutil.ParseBuilderID(builderStr)
	if err != nil {
		return nil
	}
	return &builder{
		project:  builderID.Project,
		v1Bucket: fmt.Sprintf("luci.%s.%s", builderID.Project, builderID.Bucket),
		v2Bucket: builderID.Bucket,
		builder:  builderID.Builder,
	}
}

func (c *cmdGetBuilder) validateFlags(ctx context.Context, positionals []string, env subcommands.Env) (err error) {
	if err := pingHost(c.bbHost); err != nil {
		return errors.Annotate(err, "buildbucket host").Err()
	}

	bldr := parseV1Builder(positionals[0])
	if bldr == nil {
		bldr = parseV2Builder(positionals[0])
	}
	if bldr == nil {
		err = errors.Reason("cannot parse builder: %q", positionals[0]).Err()
		return
	}

	c.builder = bldr.builder
	if c.realBuild {
		c.project = bldr.project
		c.bucket = bldr.v2Bucket
		if c.project == "" {
			return errors.New("empty project")
		}
	} else {
		c.bucket = bldr.v1Bucket
	}
	if c.bucket == "" {
		return errors.New("empty bucket")
	}
	if c.builder == "" {
		return errors.New("empty builder")
	}
	return nil
}

func (c *cmdGetBuilder) execute(ctx context.Context, authClient *http.Client, _ auth.Options, inJob *job.Definition) (out interface{}, err error) {
	return ledcmd.GetBuilder(ctx, authClient, ledcmd.GetBuildersOpts{
		BuildbucketHost: c.bbHost,
		Project:         c.project,
		Bucket:          c.bucket,
		Builder:         c.builder,
		Canary:          c.canary,
		ExtraTags:       c.tags,
		PriorityDiff:    c.priorityDiff,

		KitchenSupport: c.kitchenSupport,
		RealBuild:      c.realBuild,
	})
}

func (c *cmdGetBuilder) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	return c.doContextExecute(a, c, args, env)
}
