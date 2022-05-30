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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	csvtypes "gomodules.xyz/encoding/csv/types"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func NewCmdQATestConfigure() *cobra.Command {
	cwd, _ := os.Getwd()
	opts := struct {
		GoogleCredentialDir string
		ConfigDocId         string
		QATemplateDocId     string
		StartDate           string
		TestDays            int
		Duration            time.Duration
	}{
		GoogleCredentialDir: cwd,
		StartDate:           time.Now().UTC().Format(time.RFC3339),
		TestDays:            3,
		Duration:            90 * time.Minute,
	}

	cmd := &cobra.Command{
		Use:               "configure",
		Short:             "Configure Test",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.ConfigDocId == "" {
				return errors.New("missing config doc id.")
			}
			if opts.QATemplateDocId == "" {
				return errors.New("missing QA template doc id.")
			}
			startDate, err := time.Parse(time.RFC3339, opts.StartDate)
			if err != nil {
				return errors.Wrap(err, "failed to parse start date")
			}

			client, err := gdrive.DefaultClient(opts.GoogleCredentialDir)
			if err != nil {
				return errors.Wrap(err, "Error creating Google client")
			}
			svcSheets, err := sheets.NewService(context.TODO(), option.WithHTTPClient(client))
			if err != nil {
				return errors.Wrap(err, "Error creating Sheets client")
			}

			cfg := server.QuestionConfig{
				ConfigType:            server.ConfigTypeQuestion,
				QuestionTemplateDocId: opts.QATemplateDocId,
				StartDate:             csvtypes.Date{Time: startDate},
				EndDate:               csvtypes.Date{Time: startDate.Add(time.Duration(opts.TestDays) * 24 * time.Hour)},
				Duration:              csvtypes.Duration{Duration: opts.Duration},
			}
			err = server.SaveConfig(svcSheets, opts.ConfigDocId, cfg)
			if err != nil {
				return errors.Wrap(err, "failed to save config")
			}

			fmt.Println()
			fmt.Println("Email the following link to candidates:")
			fmt.Printf("https://x.appscode.com/_/qa/%s/\n", opts.ConfigDocId)
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.GoogleCredentialDir, "google.credential-dir", opts.GoogleCredentialDir, "Directory used to store Google credential")
	cmd.Flags().StringVar(&opts.ConfigDocId, "test.config-doc-id", opts.ConfigDocId, "Google sheets id for config spread sheet")
	cmd.Flags().StringVar(&opts.QATemplateDocId, "test.qa-template-doc-id", opts.QATemplateDocId, "Google docs id for QA template")
	cmd.Flags().StringVar(&opts.StartDate, "test.start-date", opts.StartDate, "Start date for test")
	cmd.Flags().IntVar(&opts.TestDays, "test.days-to-take-test", opts.TestDays, "Number of days available to take the test")
	cmd.Flags().DurationVar(&opts.Duration, "test.duration", opts.Duration, "Test duration")

	return cmd
}
