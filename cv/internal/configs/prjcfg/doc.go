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

// Package prjcfg handles project-scoped CV config.
//
// Configs are ingested and kept up to date using `ProjectConfigRefresher`,
// which is supposed to be called frequently, typically by a cron job.
//
// Every time config is changed, corresponding Project Manager is notified.
// Additionally, Project managers is "poked" probabilistically for reliability
// even if there were no changes so long as corresponding LUCI project's config
// remains active.
//
// TODO(crbug/1221908): implement pruning of old configs.
package prjcfg
