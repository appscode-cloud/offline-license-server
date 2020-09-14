/*
Copyright The Kubepack Authors.

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
	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/spf13/cobra"
)

func NewCmdIssueFullLicense() *cobra.Command {
	opts := server.NewOptions()
	info := server.LicenseForm{
		Name:    "",
		Email:   "",
		Product: "",
		Cluster: "",
	}
	cmd := &cobra.Command{
		Use:               "issue-full-license",
		Short:             `Issue full license`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := server.New(opts)
			if err != nil {
				return err
			}
			return s.IssueEnterpriseLicense(info)
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&info.Name, "name", info.Name, "Name of the user receiving the license")
	cmd.Flags().StringVar(&info.Email, "email", info.Email, "Email of the user receiving the license")
	cmd.Flags().StringVar(&info.Product, "product", info.Product, "Product for which license will be issued")
	cmd.Flags().StringVar(&info.Cluster, "cluster", info.Cluster, "Cluster ID for which license will be issued")

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("product")
	_ = cmd.MarkFlagRequired("cluster")

	return cmd
}
