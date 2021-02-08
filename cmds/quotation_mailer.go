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
	"os"
	"path/filepath"

	"github.com/appscodelabs/offline-license-server/pkg/server"
	"github.com/mailgun/mailgun-go/v4"
	"github.com/spf13/cobra"
	gdrive "gomodules.xyz/gdrive-utils"
)

func NewCmdEmailQuotation() *cobra.Command {
	opts := server.QuotationGeneratorOptions{
		AccountsFolderId:     server.AccountFolderId,
		TemplateDocId:        "",
		LicenseSpreadsheetId: server.LicenseSpreadsheetId,
	}
	outDir := filepath.Join("/personal", "AppsCode", "quotes")
	cmd := &cobra.Command{
		Use:               "mail",
		Short:             "Email Quotation",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return err
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			client, err := gdrive.DefaultClient(dir)
			if err != nil {
				return err
			}

			gen := server.NewQuotationGenerator(client, opts.Complete())
			gen.Lead = opts.Lead
			quote, docId, err := gen.Generate()
			if err != nil {
				return err
			}

			mailer := gen.GetMailer()
			mailer.GoogleDocIds = map[string]string{
				gen.DocName(quote) + ".pdf": docId,
			}

			mg, err := mailgun.NewMailgunFromEnv()
			if err != nil {
				return err
			}
			return mailer.SendMail(mg, opts.Lead.Email, opts.Lead.CC, gen.DriveService)
		},
	}

	cmd.Flags().StringVar(&opts.AccountsFolderId, "accounts-folder-id", opts.AccountsFolderId, "Parent folder id where generated docs will be stored under a folder with matching email domain")
	cmd.Flags().StringVar(&opts.TemplateDocId, "template-doc-id", opts.TemplateDocId, "Template document id")
	cmd.Flags().StringVar(&outDir, "out-dir", outDir, "Path to directory where output files are stored")
	cmd.Flags().StringVar(&opts.LicenseSpreadsheetId, "spreadsheet-id", opts.LicenseSpreadsheetId, "Google Spreadsheet Id used to store quotation log")

	cmd.Flags().StringVar(&opts.Lead.Name, "lead.name", opts.Lead.Name, "Name of lead")
	cmd.Flags().StringVar(&opts.Lead.Email, "lead.email", opts.Lead.Email, "Email of lead")
	cmd.Flags().StringVar(&opts.Lead.CC, "lead.cc", opts.Lead.CC, "CC the quotation to these command separated emails")
	cmd.Flags().StringVar(&opts.Lead.Title, "lead.title", opts.Lead.Title, "Job title of lead")
	cmd.Flags().StringVar(&opts.Lead.Telephone, "lead.telephone", opts.Lead.Telephone, "Telephone number of lead")
	cmd.Flags().StringVar(&opts.Lead.Product, "lead.product", opts.Lead.Product, "Name of product for which quotation is requested")
	cmd.Flags().StringVar(&opts.Lead.Company, "lead.company", opts.Lead.Company, "Name of company of the lead")

	return cmd
}
