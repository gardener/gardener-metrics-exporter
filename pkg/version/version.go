// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
