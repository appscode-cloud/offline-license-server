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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/avct/uasurfer"
	"github.com/davegardnerisme/phonegeocode"
	"github.com/mailgun/mailgun-go/v4"
	"golang.org/x/net/context"
	. "gomodules.xyz/email-providers"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	gdrive "gomodules.xyz/gdrive-utils"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"gopkg.in/macaron.v1"
)

var templateIds = map[string]string{
	"kubedb-payg":         "1n8zRoI5qjBaqa5hrogAey8OFd8-q7nCE9ysxwullb0g", // "kubedb-30"
	"stash-payg":          "1zvnJ6PNWqesnh9-33kF47k2jN2WSPwxrlPPRojSO1Y0",
	"stash-on-demand-100": "1U4GAovUia7K96PpBj0PWj4juTPaHMKB4nNK9LPyLgFk",
	"stash-50":            "1EXMmcztXGb-EOrebHCrPrhFwQuRB0RpTl0UVeMtcMNk",
	"stash-100":           "1Y2z7UZIIuvF3Twka6tXoovkbxyxXXz4qLnr9W43BIFs",
	"kubedb-30":           "1n8zRoI5qjBaqa5hrogAey8OFd8-q7nCE9ysxwullb0g",
	"kubedb-40":           "1s5751cd1SWZAy824njvTz2-iSC4V7NXRoFoCmZfoIcQ",
	"kubedb-45":           "1VN3C_fDdUG_-zgFwvPkASVYzVmVr9E2Scv1Z2uqBRrY",
	"kubedb-cluster-edu":  "18niPAUxB0OzsWTSln2OYuMqlXvHidozquqVwhtaFKYg",
	"kubedb-cluster-gov":  "11cfcXar6p9cuHQhecOu0MW6f5zTYBbXjY2P0KpQCZq0",
	"combined-cluster":    "14aeqhYnU88it0D8cXT65lfVefxmouaeu3buNnUppw-U",
}

type QuotationForm struct {
	Name      string `form:"name" binding:"Required" json:"name"`
	Email     string `form:"email" binding:"Required" json:"email"`
	Title     string `form:"title" binding:"Required" json:"title"`
	Telephone string `form:"telephone" binding:"Required" json:"telephone"`
	Product   string `form:"product" binding:"Required" json:"product"`
	Company   string `form:"company" binding:"Required" json:"company"`
	Tos       string `form:"tos" binding:"Required" json:"tos"`
}

func (form QuotationForm) Validate() error {
	if form.Product != "kubedb-payg" && form.Product != "stash-payg" {
		return fmt.Errorf("unknown plan: %s", form.Product)
	}
	if agree, _ := strconv.ParseBool(form.Tos); !agree {
		return fmt.Errorf("user must agree to terms and services")
	}
	return nil
}

func (form QuotationForm) Replacements() map[string]string {
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

	Lead QuotationForm
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
		cfg.TemplateDocId = id
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
	cfg  QuotationGeneratorConfig
	Lead QuotationForm

	Location GeoLocation
	UA       *uasurfer.UserAgent

	srvDrive *drive.Service
	srvDoc   *docs.Service
	srvSheet *gdrive.Spreadsheet
}

func NewQuotationGenerator(client *http.Client, cfg QuotationGeneratorConfig) *QuotationGenerator {
	srvDrive, err := drive.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	srvDoc, err := docs.NewService(context.TODO(), option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Docs client: %v", err)
	}

	srvSheet, err := gdrive.NewSpreadsheet(cfg.LicenseSpreadsheetId, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	return &QuotationGenerator{
		cfg:      cfg,
		srvDrive: srvDrive,
		srvDoc:   srvDoc,
		srvSheet: srvSheet,
	}
}

func (gen *QuotationGenerator) Generate() (string, string, error) {
	if gen.Lead.Telephone != "" && gen.Location.Country == "" {
		tel := SanitizeTelNumber(gen.Lead.Telephone)
		if !strings.HasPrefix(tel, "+") && len(tel) == 10 {
			tel = "+1" + tel
		}
		if cc, err := phonegeocode.New().Country(tel); err == nil {
			gen.Location.Country = cc
		}
	}

	replacements := gen.Lead.Replacements()
	quote, err := logQuotation(gen.srvSheet, []string{
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
		gen.Lead.Name,
		gen.Lead.Title,
		gen.Lead.Email,
		gen.Lead.Telephone,
		gen.Lead.Company,
		replacements["{{website}}"],
		gen.cfg.TemplateDoc,
		replacements["{{prep-date}}"],
		replacements["{{expiry-date}}"],
		gen.Location.IP,
		gen.Location.Timezone,
		gen.Location.City,
		gen.Location.Country,
		gen.Location.Coordinates,
		gen.UA.OS.Name.StringTrimPrefix(),
		gen.UA.DeviceType.StringTrimPrefix(),
	})
	if err != nil {
		return "", "", fmt.Errorf("unable to append quotation: %v", err)
	}
	replacements["{{quote}}"] = quote

	var domainFolderId string

	// https://developers.google.com/drive/api/v3/search-files
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents", FolderName(gen.Lead.Email), gen.cfg.AccountsFolderId)
	files, err := gen.srvDrive.Files.List().Q(q).Spaces("drive").Do()
	if err != nil {
		return "", "", err
	}
	if len(files.Files) > 0 {
		domainFolderId = files.Files[0].Id
	} else {
		// https://developers.google.com/drive/api/v3/folder#java
		folderMetadata := &drive.File{
			Name:     FolderName(gen.Lead.Email),
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{gen.cfg.AccountsFolderId},
		}
		folder, err := gen.srvDrive.Files.Create(folderMetadata).Fields("id").Do()
		if err != nil {
			return "", "", err
		}
		domainFolderId = folder.Id
	}
	fmt.Println("Using domain folder id:", domainFolderId)

	// https://developers.google.com/docs/api/how-tos/documents#copying_an_existing_document
	copyMetadata := &drive.File{
		Name:    gen.DocName(quote),
		Parents: []string{domainFolderId},
	}
	copyFile, err := gen.srvDrive.Files.Copy(gen.cfg.TemplateDocId, copyMetadata).Fields("id", "parents").Do()
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
	doc, err := gen.srvDoc.Documents.BatchUpdate(copyFile.Id, req).Do()
	if err != nil {
		return "", "", err
	}
	return quote, doc.DocumentId, nil
}

func (gen *QuotationGenerator) DocName(quote string) string {
	return fmt.Sprintf("%s QUOTE #%s", FolderName(gen.Lead.Email), quote)
}

func (gen *QuotationGenerator) GetMailer() Mailer {
	doc := strings.ToLower(gen.cfg.TemplateDoc)
	if strings.Contains(doc, "kubedb") {
		return NewKubeDBQuotationMailer(gen.Lead)
	} else if strings.Contains(doc, "stash") {
		return NewStashQuotationMailer(gen.Lead)
	}
	panic(fmt.Errorf("unknown template doc %s", gen.cfg.TemplateDoc))
}

func ExportPDF(srvDrive *drive.Service, docId, filename string) error {
	resp, err := srvDrive.Files.Export(docId, "application/pdf").Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	// filename := filepath.Join(gen.cfg.OutDir, FolderName(gen.cfg.Email), docName+".pdf")
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}
	fmt.Println("writing file:", filename)
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
}

func DownloadFile(srvDrive *drive.Service, docId, filename string) error {
	resp, err := srvDrive.Files.Get(docId).Download()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, resp.Body)
	if err != nil {
		return err
	}
	// filename := filepath.Join(gen.cfg.OutDir, FolderName(gen.cfg.Email), docName+".pdf")
	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		return err
	}
	fmt.Println("writing file:", filename)
	return ioutil.WriteFile(filename, buf.Bytes(), 0644)
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

func (s *Server) HandleEmailQuotation(ctx *macaron.Context, lead QuotationForm) error {
	cfg := QuotationGeneratorConfig{
		AccountsFolderId:     AccountFolderId,
		TemplateDocId:        templateIds[lead.Product],
		TemplateDoc:          lead.Product,
		LicenseSpreadsheetId: LicenseSpreadsheetId,
	}

	gen := NewQuotationGenerator(s.driveClient, cfg)
	gen.Lead = lead
	gen.UA = uasurfer.Parse(ctx.Req.UserAgent())
	location := GeoLocation{
		IP: GetIP(ctx.Req.Request),
	}
	DecorateGeoData(s.geodb, &location)
	gen.Location = location

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

	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		return err
	}
	err = mailer.SendMail(mg, lead.Email, srvDrive)
	if err != nil {
		return err
	}

	return s.noteEventQuotation(lead, EventQuotationGenerated{
		BaseNoteDescription: freshsalesclient.BaseNoteDescription{
			Event: "quotation_generated",
			Client: freshsalesclient.ClientInfo{
				OS:     gen.UA.OS.Name.StringTrimPrefix(),
				Device: gen.UA.DeviceType.StringTrimPrefix(),
				Location: freshsalesclient.GeoLocation{
					City:    location.City,
					Country: location.Country,
				},
			},
		},
		Quotation:     quote,
		TemplateDoc:   gen.cfg.TemplateDoc,
		TemplateDocId: gen.cfg.TemplateDocId,
	})
}