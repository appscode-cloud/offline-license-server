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

package server

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/mailgun/mailgun-go/v4"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/x/log"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

const (
	PaymentTermAnnual = "annual"
	PaymentTermPAYG   = "payg"
)

const (
	SupportPlanBasic    = "basic"
	SupportPlanGold     = "gold"
	SupportPlanPlatinum = "platinum"
)

type EULAInfo struct {
	Company     string `form:"company" binding:"Required" json:"company" csv:"company"`
	Domain      string `form:"domain" binding:"Required" json:"domain" csv:"domain"`
	Address     string `form:"address" binding:"Required" json:"address" csv:"address"`
	Quotation   string `form:"quotation" binding:"Required" json:"quotation" csv:"quotation"`
	Product     string `form:"product" binding:"Required" json:"product" csv:"product"`
	PaymentTerm string `form:"payment-term" binding:"Required" json:"payment-term" csv:"payment-term"`
	SupportPlan string `form:"support-plan" binding:"Required" json:"support-plan" csv:"support-plan"`

	TermYears   int       `form:"-" json:"-" csv:"term-years"`
	EULADocLink string    `form:"-" json:"-" csv:"eula"`
	PreparedOn  OfferDate `form:"-" json:"-" csv:"prepared-on"`
}

func (form EULAInfo) Data() map[string]string {
	return map[string]string{
		"{{company}}":      form.Company,
		"{{domain}}":       form.Domain,
		"{{address}}":      form.Address,
		"{{quotation}}":    form.Quotation,
		"{{product}}":      form.Product,
		"{{payment-term}}": form.PaymentTerm,
		"{{support-plan}}": form.SupportPlan,
		"{{term-years}}":   numfmt.Sprintf("%d", form.TermYears),
		// "{{prepared-on}}":  string(form.PreparedOn),
		// "{{eula}}":         form.EULADocLink,
	}
}

func (form *EULAInfo) Complete() error {
	now := time.Now()
	form.PreparedOn = NewOfferOfferDate(now)
	form.TermYears = 1 // for now only support 1 year

	u, err := url.Parse(form.Domain)
	if err != nil {
		return err
	}
	form.Domain = u.Hostname()
	return nil
}

func (form EULAInfo) Validate() error {
	if form.PaymentTerm == PaymentTermPAYG && form.SupportPlan == SupportPlanPlatinum {
		return errors.New("support plan Platinum is not offered with PAYG contract")
	}
	return nil
}

type EULATemplateKey struct {
	PaymentTerm string
	SupportPlan string
}

var eulaTemplates = map[EULATemplateKey]string{
	{
		PaymentTerm: "annual",
		SupportPlan: "basic",
	}: "1QBIjMA-hqfnm5H879Y9zSLGufVeQrHD3cLO0mVQNxPo",
	{
		PaymentTerm: "annual",
		SupportPlan: "gold",
	}: "16LawcDIbeNNIGJeS0pKmXYFukMQMw0olqL5gVWlmD3c",
	{
		PaymentTerm: "annual",
		SupportPlan: "platinum",
	}: "1bcMuWQBdDT8I4XMAdHyyYJSMjlEHjxuIWclvvF5vajQ",
	{
		PaymentTerm: "payg",
		SupportPlan: "basic",
	}: "16LB_aBLWn44MMn5gDyOTL7cehLpvsQAtVO71gDVY3Gs",
	{
		PaymentTerm: "payg",
		SupportPlan: "gold",
	}: "10_tM2wUTxWRKhyGIOLWK1TuZ2ivRFdCM69iRFU5FgNI",
}

func (s *Server) GenerateEULA(info *EULAInfo) (string, error) {
	var domainFolderId string

	// https://developers.google.com/drive/api/v3/search-files
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents", info.Domain, AccountFolderId)
	files, err := s.srvDrive.Files.List().Q(q).Spaces("drive").Do()
	if err != nil {
		return "", err
	}
	if len(files.Files) > 0 {
		domainFolderId = files.Files[0].Id
	} else {
		// https://developers.google.com/drive/api/v3/folder#java
		folderMetadata := &drive.File{
			Name:     info.Domain,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{AccountFolderId},
		}
		folder, err := s.srvDrive.Files.Create(folderMetadata).Fields("id").Do()
		if err != nil {
			return "", err
		}
		domainFolderId = folder.Id
	}
	fmt.Println("Using domain folder id:", domainFolderId)

	go func() {
		docId, err := s.generateEULADoc(info, domainFolderId)
		if err != nil {
			log.Warningln(err)
			return
		}
		info.EULADocLink = fmt.Sprintf("https://docs.google.com/document/d/%s/edit", docId)
		fmt.Println("EULA docId:", docId)

		// record in spreadsheet
		clients := []*EULAInfo{
			info,
		}
		writer := gdrive.NewWriter(s.srvSheets, LicenseSpreadsheetId, "EULA Log")
		err = gocsv.MarshalCSV(clients, writer)
		if err != nil {
			log.Warningln(err)
			return
		}

		// mail HR
		mailer := NewEULAMailer(info)
		fmt.Println("sending email for generated EULA", info.Domain)
		mg, err := mailgun.NewMailgunFromEnv()
		if err != nil {
			log.Warningln(err)
			return
		}
		err = mailer.SendMail(mg, MailSales, "", nil)
		if err != nil {
			log.Warningln(err)
			return
		}
	}()

	return domainFolderId, nil
}

func (s *Server) generateEULADoc(info *EULAInfo, domainFolderId string) (string, error) {
	date, err := info.PreparedOn.Parse()
	if err != nil {
		return "", err
	}
	docName := fmt.Sprintf("%s EULA %s - %s", info.Domain, info.Quotation, date.Format("2006-01-02"))

	docKey := EULATemplateKey{
		PaymentTerm: info.PaymentTerm,
		SupportPlan: info.SupportPlan,
	}
	templateDocId, ok := eulaTemplates[docKey]
	if !ok {
		return "", fmt.Errorf("no template doc founder for %+v", docKey)
	}

	// https://developers.google.com/docs/api/how-tos/documents#copying_an_existing_document
	copyMetadata := &drive.File{
		Name:    docName,
		Parents: []string{domainFolderId},
	}
	copyFile, err := s.srvDrive.Files.Copy(templateDocId, copyMetadata).Fields("id", "parents").Do()
	if err != nil {
		return "", err
	}

	// https://developers.google.com/docs/api/how-tos/merge
	replacements := info.Data()
	req := &docs.BatchUpdateDocumentRequest{
		Requests: make([]*docs.Request, 0, len(replacements)),
	}
	for k, v := range replacements {
		req.Requests = append(req.Requests, &docs.Request{
			ReplaceAllText: &docs.ReplaceAllTextRequest{
				ContainsText: &docs.SubstringMatchCriteria{
					MatchCase: true,
					Text:      k,
				},
				ReplaceText: v,
			},
		})
	}
	doc, err := s.srvDoc.Documents.BatchUpdate(copyFile.Id, req).Do()
	if err != nil {
		return "", err
	}
	return doc.DocumentId, nil
}
