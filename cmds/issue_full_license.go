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
	"fmt"
	"os"

	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/rickb777/date/period"
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
	d, _ := period.NewOf(server.DefaultTTLForEnterpriseProduct)
	cmd := &cobra.Command{
		Use:               "issue-full-license",
		Short:             `Issue full license`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(d.Duration())
			d2, _ := d.Duration()
			if d2 > server.DefaultTTLForEnterpriseProduct {
				// ask for confirmation
				fmt.Printf("Do you want to issue license for %v? [Y/N]", d)
				if !askForConfirmation() {
					fmt.Println("GoodBye!")
					os.Exit(1)
				}
			}
			s, err := server.New(opts)
			if err != nil {
				return err
			}
			defer func() {
				s.Close()
			}()
			return s.IssueEnterpriseLicense(info, d2)
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&info.Name, "name", info.Name, "Name of the user receiving the license")
	cmd.Flags().StringVar(&info.Email, "email", info.Email, "Email of the user receiving the license")
	cmd.Flags().StringVar(&info.Product, "product", info.Product, "Product for which license will be issued")
	cmd.Flags().StringVar(&info.Cluster, "cluster", info.Cluster, "Cluster ID for which license will be issued")
	cmd.Flags().Var(&d, "duration", "Duration for the new license")

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("product")
	_ = cmd.MarkFlagRequired("cluster")

	return cmd
}

// ref: https://gist.github.com/albrow/5882501

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		panic(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

// You might want to put the following two functions in a separate utility package.

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}
