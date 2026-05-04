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
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	gdrive "gomodules.xyz/gdrive-utils"
	"k8s.io/klog/v2"
)

type KubeDBSalesQAInfo struct {
	Distro      string `json:"distro" csv:"distro"`
	DistroLabel string `json:"distroLabel" csv:"distro-label"`

	ContactName    string `json:"contactName" csv:"contact-name"`
	ContactEmail   string `json:"contactEmail" csv:"contact-email"`
	ContactCompany string `json:"contactCompany" csv:"contact-company"`
	ContactPhone   string `json:"contactPhone" csv:"contact-phone"`
	ContactTitle   string `json:"contactTitle" csv:"contact-title"`
	ContactCountry string `json:"contactCountry" csv:"contact-country"`
	ContactAddress string `json:"contactAddress" csv:"contact-address"`
	ContactNotes   string `json:"contactNotes" csv:"contact-notes"`

	HotCount  int `json:"hotCount" csv:"hot-count"`
	WarmCount int `json:"warmCount" csv:"warm-count"`
	ColdCount int `json:"coldCount" csv:"cold-count"`

	Verdict     string `json:"verdict" csv:"verdict"`
	VerdictText string `json:"verdictText" csv:"verdict-text"`
	NotesJSON   string `json:"notesJson" csv:"notes-json"`

	SubmittedOn OfferDate `json:"-" csv:"submitted-on"`
}

func (form *KubeDBSalesQAInfo) Complete() {
	now := time.Now()
	form.SubmittedOn = NewOfferOfferDate(now)

	form.Distro = strings.TrimSpace(form.Distro)
	form.DistroLabel = strings.TrimSpace(form.DistroLabel)
	form.ContactName = strings.TrimSpace(form.ContactName)
	form.ContactEmail = strings.TrimSpace(form.ContactEmail)
	form.ContactCompany = strings.TrimSpace(form.ContactCompany)
	form.ContactPhone = strings.TrimSpace(form.ContactPhone)
	form.ContactTitle = strings.TrimSpace(form.ContactTitle)
	form.ContactCountry = strings.TrimSpace(form.ContactCountry)
	form.ContactAddress = strings.TrimSpace(form.ContactAddress)
	form.ContactNotes = strings.TrimSpace(form.ContactNotes)
	form.Verdict = strings.TrimSpace(form.Verdict)
	form.VerdictText = strings.TrimSpace(form.VerdictText)
	form.NotesJSON = strings.TrimSpace(form.NotesJSON)
}

func (form KubeDBSalesQAInfo) Validate() error {
	if form.Distro == "" {
		return fmt.Errorf("kubernetes distribution is required")
	}
	if form.ContactName == "" {
		return fmt.Errorf("contact name is required")
	}
	if form.ContactEmail == "" || !strings.Contains(form.ContactEmail, "@") {
		return fmt.Errorf("invalid contact email: %s", form.ContactEmail)
	}
	if form.ContactCompany == "" {
		return fmt.Errorf("contact company is required")
	}
	if form.ContactCountry == "" {
		return fmt.Errorf("contact country is required")
	}
	return nil
}

func (s *Server) HandleKubeDBSalesQA(info *KubeDBSalesQAInfo) error {
	go func() {
		entries := []*KubeDBSalesQAInfo{info}
		writer := gdrive.NewWriter(s.srvSheets, DealSpreadsheetId, "KubeDB Sales QA")
		err := gocsv.MarshalCSV(entries, writer)
		if err != nil {
			klog.Warningln(err)
			return
		}

		mailer := NewKubeDBSalesQAMailer(info)
		fmt.Println("sending email for kubedb sales qa", info.ContactCompany)
		err = mailer.SendMail(s.mg, MailIncomingDeals, "", nil)
		if err != nil {
			klog.Warningln(err)
			return
		}
	}()

	return nil
}
