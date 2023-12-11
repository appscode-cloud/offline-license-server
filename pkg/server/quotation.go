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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/davegardnerisme/phonegeocode"
	"golang.org/x/net/context"
	. "gomodules.xyz/email-providers"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	gdrive "gomodules.xyz/gdrive-utils"
	listmonkclient "gomodules.xyz/listmonk-client-go"
	"gomodules.xyz/mailer"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"gopkg.in/macaron.v1"
	"k8s.io/klog/v2"
)

type QuoteInfo struct {
	TemplateDocId string
	MailingLists  []string
}

var templateIds = map[string]QuoteInfo{
	"kubedb-enterprise-50": {
		TemplateDocId: "1oD9_jpzRL5djK7i9jQ74PvzFx2xN3O867DrWSiAZrSg",
		MailingLists:  []string{listmonkclient.MailingList_KubeDB, listmonkclient.MailingList_Stash},
	},
	"kubedb-enterprise-60": {
		TemplateDocId: "1Li9ERfgYEXL80_kE-Mg1PKCQXt2cs6aK-xzhP9j45Q4",
		MailingLists:  []string{listmonkclient.MailingList_KubeDB, listmonkclient.MailingList_Stash},
	},
	"kubedb-payg": {
		TemplateDocId: "1w0EeXotjL6PWNbFdlGH4cnPb_tckf_uZeWp14UwALRA",
		MailingLists:  []string{listmonkclient.MailingList_KubeDB, listmonkclient.MailingList_Stash},
	},
	"kubedb-reseller": {
		TemplateDocId: "1w46SFq9kA8ciibINOv7a4frzoogt-yQxmFegr9BWMcc",
		MailingLists:  []string{listmonkclient.MailingList_KubeDB, listmonkclient.MailingList_Stash},
	},
	"kubedb-unlimited": {
		TemplateDocId: "13_Z2EGGdS8WASXqjMusojum0Do3U4nXytXxgZmQkxRU",
		MailingLists:  []string{listmonkclient.MailingList_KubeDB, listmonkclient.MailingList_Stash},
	},
	"stash-enterprise": {
		TemplateDocId: "1PDwas0A119L232ZyLi4reg3-sOXVagB5bhIz8acwaKY",
		MailingLists:  []string{listmonkclient.MailingList_Stash},
	},
	"stash-payg": {
		TemplateDocId: "1ao-gnhco2KY6ETvgLEZBEjLDtA_O5t7yM-WKiT1--kM",
		MailingLists:  []string{listmonkclient.MailingList_Stash},
	},
	"stash-unlimited": {
		TemplateDocId: "1iqaj1GOzo4Bj_kb3y4eondlqD7PfTmVQx4l_ECczILM",
		MailingLists:  []string{listmonkclient.MailingList_Stash},
	},
	"kubeform-enterprise": {
		TemplateDocId: "1ERPvo8KrTL6Cmk067guyqPYyrGuKTQsZUUJAakeMYa4",
		MailingLists:  []string{listmonkclient.MailingList_Kubeform},
	},
	"kubeform-payg": {
		TemplateDocId: "1w0KkMno6HIe33iefjTfijCKGBQZ3LIrVRAJRAdGYhVo",
		MailingLists:  []string{listmonkclient.MailingList_Kubeform},
	},
	"kubevault-enterprise": {
		TemplateDocId: "1iDocKQPUDADVMj3cBcReW6fE-1RD6usS9MjjS8RF0EY",
		MailingLists:  []string{listmonkclient.MailingList_KubeVault},
	},
	"kubevault-payg": {
		TemplateDocId: "1Z_z3VxxiBDHF4aB9UkbjvH74P6PGGJSzhZuMsOGxJ1o",
		MailingLists:  []string{listmonkclient.MailingList_KubeVault},
	},
	"voyager-enterprise": {
		TemplateDocId: "1GQ8UocSIgYhWRD-Zulnz4USOP9yg0K4pTRAp1fWjg8E",
		MailingLists:  []string{listmonkclient.MailingList_Voyager},
	},
	"voyager-payg": {
		TemplateDocId: "1NuOO2cpH89GKFNBNxvayLXcXOcbfcH6eIMJF2t-JCeo",
		MailingLists:  []string{listmonkclient.MailingList_Voyager},
	},
	"guard-enterprise": {
		TemplateDocId: "12a3NFvdgbfVbmmNEMt0IFbotiUfGsDX8_-NIEsbKS5w",
		MailingLists:  []string{listmonkclient.MailingList_Console},
	},
	"config-syncer-enterprise": {
		TemplateDocId: "1091KacS4i6fO8m11i8rrL825uFJpKZwHeP8uXQfukbM",
		MailingLists:  []string{listmonkclient.MailingList_Console},
	},
}

type QuotationForm struct {
	Name      string   `form:"name" binding:"Required" json:"name"`
	Email     string   `form:"email" binding:"Required" json:"email"`
	CC        string   `form:"cc" json:"cc"`
	Title     string   `form:"title" binding:"Required" json:"title"`
	Telephone string   `form:"telephone" binding:"Required" json:"telephone"`
	Product   []string `form:"product" binding:"Required" json:"product"`
	Company   string   `form:"company" binding:"Required" json:"company"`
	Tos       string   `form:"tos" binding:"Required" json:"tos"`
}

type ProductQuotation struct {
	Name      string `form:"name" binding:"Required" json:"name"`
	Email     string `form:"email" binding:"Required" json:"email"`
	CC        string `form:"cc" json:"cc"`
	Title     string `form:"title" binding:"Required" json:"title"`
	Telephone string `form:"telephone" binding:"Required" json:"telephone"`
	Product   string `form:"product" binding:"Required" json:"product"`
	Company   string `form:"company" binding:"Required" json:"company"`
}

func (form QuotationForm) Validate() error {
	for _, product := range form.Product {
		if _, ok := templateIds[product]; !ok {
			return fmt.Errorf("unknown plan: %s", form.Product)
		}
	}
	if agree, _ := strconv.ParseBool(form.Tos); !agree {
		return fmt.Errorf("user must agree to terms and services")
	}
	return nil
}

func (form ProductQuotation) Replacements() map[string]string {
	data, err := json.Marshal(form)
	if err != nil {
		panic(err)
	}
	fields := map[string]string{}
	err = json.Unmarshal(data, &fields)
	if err != nil {
		panic(err)
	}
	replacements := map[string]string{}
	for k, v := range fields {
		replacements["{{"+k+"}}"] = v
	}

	if IsPublicEmail(form.Email) {
		replacements["{{website}}"] = ""
	} else {
		replacements["{{website}}"] = Domain(form.Email)
	}

	now := time.Now()
	replacements["{{prep-date}}"] = now.Format("Jan 2, 2006")
	replacements["{{expiry-date}}"] = now.Add(30 * 24 * time.Hour).Format("Jan 2, 2006")

	return replacements
}

type QuotationGeneratorOptions struct {
	AccountsFolderId string
	TemplateDocId    string
	// ReplacementInput     map[string]string
	LicenseSpreadsheetId string

	Contact QuotationForm
}

func (opts QuotationGeneratorOptions) Validate() error {
	if opts.AccountsFolderId == "" {
		return errors.New("missing parent folder id")
	}
	if opts.TemplateDocId == "" {
		return errors.New("missing template doc id")
	}
	return nil
}

func (opts QuotationGeneratorOptions) Complete() QuotationGeneratorConfig {
	cfg := QuotationGeneratorConfig{
		AccountsFolderId:     opts.AccountsFolderId,
		TemplateDocId:        opts.TemplateDocId,
		TemplateDoc:          opts.TemplateDocId,
		LicenseSpreadsheetId: opts.LicenseSpreadsheetId,
	}

	if id, ok := templateIds[cfg.TemplateDocId]; ok {
		cfg.TemplateDocId = id.TemplateDocId
	}

	return cfg
}

type QuotationGeneratorConfig struct {
	AccountsFolderId     string
	TemplateDocId        string
	TemplateDoc          string
	LicenseSpreadsheetId string
}

type QuotationGenerator struct {
	cfg     QuotationGeneratorConfig
	Contact ProductQuotation

	Location GeoLocation
	UA       *uasurfer.UserAgent

	DriveService *drive.Service
	DocService   *docs.Service
	SheetService *gdrive.Spreadsheet

	FolderChan chan<- string
}

func NewQuotationGenerator(client *http.Client, cfg QuotationGeneratorConfig) *QuotationGenerator {
	srvDrive, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		klog.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	srvDoc, err := docs.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		klog.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	srvSheets, err := sheets.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		klog.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	srvSheet, err := gdrive.NewSpreadsheet(srvSheets, cfg.LicenseSpreadsheetId)
	if err != nil {
		klog.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return &QuotationGenerator{
		cfg:          cfg,
		DriveService: srvDrive,
		DocService:   srvDoc,
		SheetService: srvSheet,
	}
}

func (gen *QuotationGenerator) Generate() (string, string, error) {
	if gen.Contact.Telephone != "" && gen.Location.Country == "" {
		tel := SanitizeTelNumber(gen.Contact.Telephone)
		if !strings.HasPrefix(tel, "+") && len(tel) == 10 {
			tel = "+1" + tel
		}
		if cc, err := phonegeocode.New().Country(tel); err == nil {
			gen.Location.Country = cc
		}
	}

	replacements := gen.Contact.Replacements()
	var clientOS, clientDevice string
	if gen.UA != nil {
		clientOS = gen.UA.OS.Name.StringTrimPrefix()
		clientDevice = gen.UA.DeviceType.StringTrimPrefix()
	}
	quote, err := logQuotation(gen.SheetService, []string{
		"Quotation #",
		"Name",
		"Title",
		"Email",
		"Telephone",
		"Company",
		"Website",
		"Pricing Template",
		"Preparation Date",
		"Expiration Date",
		"IP",
		"Timezone",
		"City",
		"Country",
		"Coordinates",
		"Client OS",
		"Client Device",
	}, []string{
		"AC_DETECT_QUOTE",
		gen.Contact.Name,
		gen.Contact.Title,
		gen.Contact.Email,
		gen.Contact.Telephone,
		gen.Contact.Company,
		replacements["{{website}}"],
		gen.cfg.TemplateDoc,
		replacements["{{prep-date}}"],
		replacements["{{expiry-date}}"],
		gen.Location.IP,
		gen.Location.Timezone,
		gen.Location.City,
		gen.Location.Country,
		gen.Location.Coordinates,
		clientOS,
		clientDevice,
	})
	if err != nil {
		return "", "", fmt.Errorf("unable to append quotation: %v", err)
	}
	replacements["{{quote}}"] = quote

	var domainFolderId string

	// https://developers.google.com/drive/api/v3/search-files
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents", FolderName(gen.Contact.Email), gen.cfg.AccountsFolderId)
	files, err := gen.DriveService.Files.List().Q(q).Spaces("drive").Do()
	if err != nil {
		return "", "", err
	}
	if len(files.Files) > 0 {
		domainFolderId = files.Files[0].Id
	} else {
		// https://developers.google.com/drive/api/v3/folder#java
		folderMetadata := &drive.File{
			Name:     FolderName(gen.Contact.Email),
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{gen.cfg.AccountsFolderId},
		}
		folder, err := gen.DriveService.Files.Create(folderMetadata).Fields("id").Do()
		if err != nil {
			return "", "", err
		}
		domainFolderId = folder.Id
	}
	fmt.Println("Using domain folder id:", domainFolderId)
	if gen.FolderChan != nil {
		gen.FolderChan <- domainFolderId
	}

	// https://developers.google.com/docs/api/how-tos/documents#copying_an_existing_document
	copyMetadata := &drive.File{
		Name:    gen.DocName(quote),
		Parents: []string{domainFolderId},
	}
	copyFile, err := gen.DriveService.Files.Copy(gen.cfg.TemplateDocId, copyMetadata).Fields("id", "parents").Do()
	if err != nil {
		return "", "", err
	}
	fmt.Println("doc id:", copyFile.Id)

	// https://developers.google.com/docs/api/how-tos/merge
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
	doc, err := gen.DocService.Documents.BatchUpdate(copyFile.Id, req).Do()
	if err != nil {
		return "", "", err
	}
	return quote, doc.DocumentId, nil
}

func (gen *QuotationGenerator) DocName(quote string) string {
	return fmt.Sprintf("%s QUOTE #%s", FolderName(gen.Contact.Email), quote)
}

func (gen *QuotationGenerator) GetMailer() mailer.Mailer {
	switch strings.ToLower(gen.cfg.TemplateDoc) {
	case "kubedb-payg":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "KubeDB",
			FullPlan:         "Pay-As-You-Go (PAYG)",
			Plan:             "PAYG",
		})
	case "kubedb-enterprise":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "KubeDB",
			FullPlan:         "Enterprise",
			Plan:             "Enterprise",
		})
	case "stash-payg":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Stash",
			FullPlan:         "Pay-As-You-Go (PAYG)",
			Plan:             "PAYG",
		})
	case "stash-enterprise":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Stash",
			FullPlan:         "Enterprise",
			Plan:             "Enterprise",
		})
	case "kubeform-payg":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Kubeform",
			FullPlan:         "Pay-As-You-Go (PAYG)",
			Plan:             "PAYG",
		})
	case "kubeform-enterprise":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Kubeform",
			FullPlan:         "Enterprise",
			Plan:             "Enterprise",
		})
	case "kubevault-payg":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "KubeVault",
			FullPlan:         "Pay-As-You-Go (PAYG)",
			Plan:             "PAYG",
		})
	case "kubevault-enterprise":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "KubeVault",
			FullPlan:         "Enterprise",
			Plan:             "Enterprise",
		})
	case "voyager-payg":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Voyager",
			FullPlan:         "Pay-As-You-Go (PAYG)",
			Plan:             "PAYG",
		})
	case "voyager-enterprise":
		return NewQuotationMailer(QuotationEmailData{
			ProductQuotation: gen.Contact,
			Offer:            "Voyager",
			FullPlan:         "Enterprise",
			Plan:             "Enterprise",
		})
	default:
		panic(fmt.Errorf("unknown template doc %s", gen.cfg.TemplateDoc))
	}
}

func SanitizeTelNumber(tel string) string {
	var buf bytes.Buffer
	for _, r := range tel {
		if r == '+' || (r >= '0' && r <= '9') {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func logQuotation(si *gdrive.Spreadsheet, headers, data []string) (string, error) {
	const sheetName = "Quotation Log"

	sheetId, err := si.EnsureSheet(sheetName, headers)
	if err != nil {
		return "", err
	}

	lastQuote, err := si.FindEmptyCell(sheetName)
	if err != nil {
		return "", err
	}

	var quote string
	now := time.Now().UTC()
	if strings.HasPrefix(lastQuote, "AC") {
		y, err := strconv.Atoi(lastQuote[2:4])
		if err != nil {
			return "", fmt.Errorf("failed to detect YY from quote %s", lastQuote)
		}
		m, err := strconv.Atoi(lastQuote[4:6])
		if err != nil {
			return "", fmt.Errorf("failed to detect MM from quote %s", lastQuote)
		}
		sl, err := strconv.Atoi(lastQuote[6:])
		if err != nil {
			return "", fmt.Errorf("failed to detect Serial# from quote %s", lastQuote)
		}

		if (now.Year()-2000) == y && m == int(now.Month()) {
			quote = fmt.Sprintf("AC%02d%02d%03d", now.Year()-2000, now.Month(), sl+1)
		} else {
			quote = fmt.Sprintf("AC%02d%02d%03d", now.Year()-2000, now.Month(), 1)
		}
	} else {
		quote = fmt.Sprintf("AC%02d%02d%03d", now.Year()-2000, now.Month(), 1)
	}
	data[0] = quote

	return quote, si.AppendRowData(sheetId, data, false)
}

func FolderName(email string) string {
	if IsPublicEmail(email) {
		return email
	}
	parts := strings.Split(email, "@")
	return parts[len(parts)-1]
}

func (s *Server) HandleEmailQuotation(ctx *macaron.Context, contact QuotationForm) error {
	folderChan := make(chan string)

	for idx, product := range contact.Product {
		cfg := QuotationGeneratorConfig{
			AccountsFolderId:     AccountFolderId,
			TemplateDocId:        templateIds[product].TemplateDocId,
			TemplateDoc:          product,
			LicenseSpreadsheetId: LicenseSpreadsheetId,
		}

		gen := NewQuotationGenerator(s.driveClient, cfg)
		gen.Contact = ProductQuotation{
			Name:      contact.Name,
			Email:     contact.Email,
			CC:        contact.CC,
			Title:     contact.Title,
			Telephone: contact.Telephone,
			Product:   product,
			Company:   contact.Company,
		}
		gen.UA = uasurfer.Parse(ctx.Req.UserAgent())
		location := GeoLocation{
			IP: GetIP(ctx.Req.Request),
		}
		DecorateGeoData(s.geodb, &location)
		gen.Location = location
		if idx == 0 {
			gen.FolderChan = folderChan
		}

		go func() {
			sendEmail := ctx.QueryBool("send_email")
			if err := s.processQuotationRequest(gen, sendEmail); err != nil {
				// email support@appscode.com failed to process request
				mailer := NewQuotationProcessFailedMailer(gen, err)
				e2 := mailer.SendMail(s.mg, MailSupport, "", nil)
				if e2 != nil {
					_, _ = fmt.Fprintf(os.Stderr, "failed send email %v", e2)
				}
			}
		}()
	}

	select {
	case folderId := <-folderChan:
		ctx.Redirect(fmt.Sprintf("https://drive.google.com/drive/folders/%s", folderId))
	case <-time.After(30 * time.Second):
		// can't wait too long. Cloudflare does not like that
		// respond(ctx, []byte("Thank you! Please check your email in a few minutes for price quotation. Don't forget to check spam folder."))
		ctx.Redirect(ctx.Req.URL.String(), http.StatusSeeOther)
	}
	return nil
}

func (s *Server) processQuotationRequest(gen *QuotationGenerator, sendEmail bool) error {
	quote, docId, err := gen.Generate()
	if err != nil {
		return err
	}

	mailer := gen.GetMailer()
	mailer.GoogleDocIds = map[string]string{
		gen.DocName(quote) + ".pdf": docId,
	}

	srvDrive, err := drive.NewService(context.TODO(), option.WithHTTPClient(s.driveClient))
	if err != nil {
		return err
	}

	if sendEmail {
		fmt.Println("sending email to", gen.Contact.Email)
		err = mailer.SendMail(s.mg, gen.Contact.Email, gen.Contact.CC, srvDrive)
		if err != nil {
			return err
		}
	}

	err = s.listmonk.SubscribeToList(listmonkclient.SubscribeRequest{
		Email:        gen.Contact.Email,
		Name:         gen.Contact.Name,
		MailingLists: templateIds[gen.Contact.Product].MailingLists,
	})
	if err != nil {
		return err
	}

	return s.noteEventQuotation(gen.Contact, EventQuotationGenerated{
		BaseNoteDescription: freshsalesclient.BaseNoteDescription{
			Event: "quotation_generated",
			Client: freshsalesclient.ClientInfo{
				OS:     gen.UA.OS.Name.StringTrimPrefix(),
				Device: gen.UA.DeviceType.StringTrimPrefix(),
				Location: freshsalesclient.GeoLocation{
					IP:          gen.Location.IP,
					Timezone:    gen.Location.Timezone,
					City:        gen.Location.City,
					Country:     gen.Location.Country,
					Coordinates: gen.Location.Coordinates,
				},
			},
		},
		Quotation:     quote,
		TemplateDoc:   gen.cfg.TemplateDoc,
		TemplateDocId: gen.cfg.TemplateDocId,
	})
}
