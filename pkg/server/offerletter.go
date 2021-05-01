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
	"fmt"
	"strconv"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/mailgun/mailgun-go/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/x/log"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
)

const (
	offerLetterSpreadsheetId = "1fZ3KEebljFOPUW01Yf-I-_idsWz6OZ1RPlZzudgL1GE"
	offerLetterFolderId      = "1NuIlHI-IwtjJREkFhMeT9lAtOLCl7NKa"
	MailHR                   = "hr.bd@appscode.com"
)

var offerLetterTemplateDocIds = map[string]string{
	"offer":    "15WECv5bM1KgJ0t342pF9FoDBuX-cSJW338zMOosp_Nc",
	"nda":      "1Ed7xPomjh5JzRGzmOueY2EwyiL_6xW0z_wc_UbFcsHs",
	"handbook": "1nsoDDnM15ZTWgmyzcc71iIWQK5auiltCQogCghHBAQU",
}

var numfmt = message.NewPrinter(language.AmericanEnglish)

type CandidateInfo struct {
	Email        string `form:"email" binding:"Required;Email" csv:"email"`
	Name         string `form:"name" binding:"Required" csv:"name"`
	Tel          string `form:"tel" binding:"Required" csv:"tel"`
	AddressLine1 string `form:"address-line-1" binding:"Required" csv:"address-line-1"`
	AddressLine2 string `form:"address-line-2" binding:"Required" csv:"address-line-2"`
	AddressLine3 string `form:"address-line-3" binding:"Required" csv:"address-line-3"`
	Title        string `form:"title" binding:"Required" csv:"title"`
	Salary       int    `form:"salary" binding:"Required" csv:"salary"`

	// https://bulma-calendar.onrender.com/#content
	// format: MM/DD/YYYY 01/02/2006
	StartDate      OfferDate `form:"start-date" binding:"Required" csv:"start-date"`
	OfferStartDate OfferDate `form:"-" csv:"offer-start-date"`
	OfferEndDate   OfferDate `form:"-" csv:"offer-end-date"`
}

func (form CandidateInfo) Data() map[string]string {
	return map[string]string{
		"{{email}}":            form.Email,
		"{{name}}":             form.Name,
		"{{tel}}":              form.Tel,
		"{{address-line-1}}":   form.AddressLine1,
		"{{address-line-2}}":   form.AddressLine2,
		"{{address-line-3}}":   form.AddressLine3,
		"{{title}}":            form.Title,
		"{{salary}}":           numfmt.Sprintf("%d", form.Salary),
		"{{start-date}}":       string(form.StartDate),
		"{{offer-start-date}}": string(form.OfferStartDate),
		"{{offer-end-date}}":   string(form.OfferEndDate),
	}
}

func (form *CandidateInfo) Complete() {
	now := time.Now()
	form.OfferStartDate = NewOfferOfferDate(now)
	form.OfferEndDate = NewOfferOfferDate(now.Add(3 * 24 * time.Hour)) // 3 days
}

func (form CandidateInfo) Validate() error {
	_, err := form.StartDate.Parse()
	return err
}

type OfferDate string

func NewOfferOfferDate(t time.Time) OfferDate {
	return OfferDate(t.Format("Jan 2, 2006"))
}

func (date OfferDate) Parse() (time.Time, error) {
	return time.Parse("Jan 2, 2006", string(date))
}

func (s *Server) GenerateOfferLetter(info *CandidateInfo) (string, error) {
	var candidateFolderId string

	// https://developers.google.com/drive/api/v3/search-files
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents", info.Email, offerLetterFolderId)
	files, err := s.srvDrive.Files.List().Q(q).Spaces("drive").Do()
	if err != nil {
		return "", err
	}
	if len(files.Files) > 0 {
		candidateFolderId = files.Files[0].Id
	} else {
		// https://developers.google.com/drive/api/v3/folder#java
		folderMetadata := &drive.File{
			Name:     info.Email,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{offerLetterFolderId},
		}
		folder, err := s.srvDrive.Files.Create(folderMetadata).Fields("id").Do()
		if err != nil {
			return "", err
		}
		candidateFolderId = folder.Id
	}

	go func() {
		fmt.Println("Employee:", info.Name)
		fmt.Println("Email:", info.Email)
		fmt.Println("Using folder id:", candidateFolderId)

		docId, err := s.generateDoc(info, candidateFolderId, "offer")
		if err != nil {
			log.Warningln(err)
			return
		}
		fmt.Println("Offer docId:", docId)

		docId, err = s.generateDoc(info, candidateFolderId, "nda")
		if err != nil {
			log.Warningln(err)
			return
		}
		fmt.Println("NDA docId:", docId)

		docId, err = s.generateDoc(info, candidateFolderId, "handbook")
		if err != nil {
			log.Warningln(err)
			return
		}
		fmt.Println("Handbook docId:", docId)

		// record in spreadsheet
		startDate, err := info.StartDate.Parse()
		if err != nil {
			log.Warningln(err)
			return
		}
		sheetName := strconv.Itoa(startDate.Year())
		clients := []*CandidateInfo{
			info,
		}
		writer := gdrive.NewWriter(s.srvSheets, offerLetterSpreadsheetId, sheetName)
		err = gocsv.MarshalCSV(clients, writer)
		if err != nil {
			log.Warningln(err)
			return
		}

		// mail HR
		mailer := NewOfferLetterMailer(info, candidateFolderId)
		fmt.Println("sending email for generated offer letter", info.Email)
		mg, err := mailgun.NewMailgunFromEnv()
		if err != nil {
			log.Warningln(err)
			return
		}
		err = mailer.SendMail(mg, MailHR, "", nil)
		if err != nil {
			log.Warningln(err)
			return
		}
	}()

	return candidateFolderId, nil
}

func (s *Server) generateDoc(info *CandidateInfo, candidateFolderId string, templateKey string) (string, error) {
	var docName string
	switch templateKey {
	case "offer":
		docName = fmt.Sprintf("Offer - %s", info.Email)
	case "nda":
		docName = fmt.Sprintf("NDA - %s", info.Email)
	case "handbook":
		docName = fmt.Sprintf("Handbook - %s", info.Email)
	}

	// https://developers.google.com/docs/api/how-tos/documents#copying_an_existing_document
	copyMetadata := &drive.File{
		Name:    docName,
		Parents: []string{candidateFolderId},
	}
	copyFile, err := s.srvDrive.Files.Copy(offerLetterTemplateDocIds[templateKey], copyMetadata).Fields("id", "parents").Do()
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
