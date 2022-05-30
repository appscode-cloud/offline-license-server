/*
Copyright AppsCode Inc.

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

package cmds

import (
	"flag"

	"github.com/spf13/cobra"
	v "gomodules.xyz/x/version"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               "offline-license-server [command]",
		Short:             `offline-license-server by AppsCode - Offline License server for AppsCode products`,
		DisableAutoGenTag: true,
	}

	flags := rootCmd.PersistentFlags()
	flags.AddGoFlagSet(flag.CommandLine)

	rootCmd.AddCommand(NewCmdQuotation())
	rootCmd.AddCommand(NewCmdCreate())
	rootCmd.AddCommand(NewCmdGet())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdIssueFullLicense())
	rootCmd.AddCommand(NewCmdGenerateAccessLogCSV())
	rootCmd.AddCommand(NewCmdQA())
	rootCmd.AddCommand(v.NewCmdVersion())
	return rootCmd
}
