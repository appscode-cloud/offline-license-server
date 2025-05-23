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
	"strings"
	"time"

	licenseapi "go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"
	"go.bytebuilders.dev/offline-license-server/pkg/server"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rickb777/date/period"
	"github.com/spf13/cobra"
)

func NewCmdIssueFullLicense() *cobra.Command {
	opts := server.NewOptions()
	info := server.LicenseForm{
		Name:         "",
		Email:        "",
		ProductAlias: "",
		Cluster:      "",
	}
	var clusters []string
	var ccList []string
	d, _ := period.NewOf(server.DefaultTTLForEnterpriseProduct)
	var expiryDate string
	var featureFlags map[string]string
	cmd := &cobra.Command{
		Use:               "issue-full-license",
		Short:             `Issue full license`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			info.CC = strings.Join(ccList, ",")

			var d2 time.Duration
			if expiryDate != "" {
				t, err := time.Parse("2006-1-2", expiryDate)
				if err != nil {
					return fmt.Errorf("failed to parse expiry date %s, err: %v", expiryDate, err)
				}
				d2 = time.Until(t) + 24*time.Hour
			} else {
				d2, _ = d.Duration()
			}
			fmt.Println(d2)
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

			if len(featureFlags) == 0 {
				// ask for confirmation
				fmt.Println("Do you want to disable analytics? [Y/N]")
				if askForConfirmation() {
					featureFlags = map[string]string{}
					featureFlags[string(licenseapi.FeatureDisableAnalytics)] = "true"
				}
			}
			ff := licenseapi.FeatureFlags{}
			for k, v := range featureFlags {
				ff[licenseapi.FeatureFlag(k)] = v
			}
			if err := ff.IsValid(); err != nil {
				panic(err)
			}

			for _, cluster := range clusters {
				fmt.Println("cluster:", cluster)
				if _, err := uuid.Parse(cluster); err != nil {
					return errors.Wrapf(err, "invalid cluster id %s", cluster)
				}
				info.Cluster = cluster
				if err := s.IssueEnterpriseLicense(info, d2, ff); err != nil {
					return errors.Wrapf(err, "failed to issue license for cluster %s", cluster)
				}
			}
			return nil
		},
	}
	opts.AddFlags(cmd.Flags())
	cmd.Flags().StringVar(&info.Name, "name", info.Name, "Name of the user receiving the license")
	cmd.Flags().StringVar(&info.Email, "email", info.Email, "Email of the user receiving the license")
	cmd.Flags().StringSliceVar(&ccList, "cc", ccList, "CC the license to these emails")
	cmd.Flags().StringVar(&info.ProductAlias, "product", info.ProductAlias, "Product for which license will be issued")
	cmd.Flags().StringSliceVar(&clusters, "cluster", clusters, "Cluster IDs for which license will be issued")
	cmd.Flags().Var(&d, "duration", "Duration for the new license")
	cmd.Flags().StringVar(&expiryDate, "expiry-date", expiryDate, "Expiry date in YYYY-MM-DD format")
	cmd.Flags().StringToStringVar(&featureFlags, "feature-flag", featureFlags, "List of feature flags")

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
