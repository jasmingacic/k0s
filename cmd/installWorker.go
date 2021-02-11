/*
Copyright 2021 Mirantis, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	installWorkerCmd = &cobra.Command{
		Use:   "worker",
		Short: "Helper command for setting up k0s as a worker node on a brand-new system. Must be run as root (or with sudo)",
		Example: `Worker subcommand allows you to pass in all available worker parameters. 
All default values of worker command will be passed to the service stub unless overriden.

Windows flags like "--api-server", "--cidr-range" and "--cluster-dns" will be ignored since install command doesn't yet support Windows services`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flagsAndVals := []string{"worker"}
			cmd.Flags().Visit(func(f *pflag.Flag) {
				if f.Value.Type() == "stringSlice" { // this is a workaround to parse f.Value correctly
					labels := f.Value.String()
					flagsAndVals = append(flagsAndVals, fmt.Sprintf(`--%s="%s"`, f.Name, strings.Trim(labels, "[]")))
				} else {
					flagsAndVals = append(flagsAndVals, fmt.Sprintf("--%s=%s", f.Name, f.Value.String()))
				}
			})

			return setup("worker", flagsAndVals)
		},
	}
)
