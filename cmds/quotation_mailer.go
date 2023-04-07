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

	"go.bytebuilders.dev/offline-license-server/pkg/server"

	"github.com/spf13/cobra"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/mailer"
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

			for _, product := range opts.Contact.Product {
				gen := server.NewQuotationGenerator(client, opts.Complete())
				gen.Contact = server.ProductQuotation{
					Name:      opts.Contact.Name,
					Email:     opts.Contact.Email,
					CC:        opts.Contact.CC,
					Title:     opts.Contact.Title,
					Telephone: opts.Contact.Telephone,
					Product:   product,
					Company:   opts.Contact.Company,
				}
				quote, docId, err := gen.Generate()
				if err != nil {
					return err
				}

				mm := gen.GetMailer()
				mm.GoogleDocIds = map[string]string{
					gen.DocName(quote) + ".pdf": docId,
				}

				mg, err := mailer.NewSMTPServiceFromEnv()
				if err != nil {
					return err
				}
				err = mm.SendMail(mg, opts.Contact.Email, opts.Contact.CC, gen.DriveService)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&opts.AccountsFolderId, "accounts-folder-id", opts.AccountsFolderId, "Parent folder id where generated docs will be stored under a folder with matching email domain")
	cmd.Flags().StringVar(&opts.TemplateDocId, "template-doc-id", opts.TemplateDocId, "Template document id")
	cmd.Flags().StringVar(&outDir, "out-dir", outDir, "Path to directory where output files are stored")
	cmd.Flags().StringVar(&opts.LicenseSpreadsheetId, "spreadsheet-id", opts.LicenseSpreadsheetId, "Google Spreadsheet Id used to store quotation log")

	cmd.Flags().StringVar(&opts.Contact.Name, "contact.name", opts.Contact.Name, "Name of contact")
	cmd.Flags().StringVar(&opts.Contact.Email, "contact.email", opts.Contact.Email, "Email of contact")
	cmd.Flags().StringVar(&opts.Contact.CC, "contact.cc", opts.Contact.CC, "CC the quotation to these command separated emails")
	cmd.Flags().StringVar(&opts.Contact.Title, "contact.title", opts.Contact.Title, "Job title of contact")
	cmd.Flags().StringVar(&opts.Contact.Telephone, "contact.telephone", opts.Contact.Telephone, "Telephone number of contact")
	cmd.Flags().StringSliceVar(&opts.Contact.Product, "contact.product", opts.Contact.Product, "Name of product for which quotation is requested")
	cmd.Flags().StringVar(&opts.Contact.Company, "contact.company", opts.Contact.Company, "Name of company of the contact")

	return cmd
}
