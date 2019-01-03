// Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package version

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

// Information will overridden by values which are passed via ldflags during build time.
var (
	gitCommit  = "unknown"
	gitVersion = "0.0.0-dev"
	buildDate  = time.RFC3339
)

// GetVersionCmd returns a pointer to a Cobra command, which expose version and build information.
func GetVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`git version : %s
git commit  : %s
build date  : %s
go version  : %s
go compiler : %s
platform    : %s/%s`, gitVersion, gitCommit, buildDate, runtime.Version(), runtime.Compiler, runtime.GOOS, runtime.GOARCH)
		},
	}
}
