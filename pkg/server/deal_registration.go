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

const (
	MailIncomingDeals = "incoming-deal-alerts@appscode.com"
)

type DealRegistrationInfo struct {
	// Partner Information
	PartnerName    string `form:"partner-name" binding:"Required" json:"partner-name" csv:"partner-name"`
	PartnerEmail   string `form:"partner-email" binding:"Required;Email" json:"partner-email" csv:"partner-email"`
	PartnerCompany string `form:"partner-company" binding:"Required" json:"partner-company" csv:"partner-company"`
	Region         string `form:"region" binding:"Required" json:"region" csv:"region"`

	// Customer Information
	CustomerName    string `form:"customer-name" binding:"Required" json:"customer-name" csv:"customer-name"`
	CustomerEmail   string `form:"customer-email" binding:"Required;Email" json:"customer-email" csv:"customer-email"`
	CustomerCompany string `form:"customer-company" binding:"Required" json:"customer-company" csv:"customer-company"`
	CustomerPhone   string `form:"customer-phone" json:"customer-phone" csv:"customer-phone"`
	CustomerAddress string `form:"customer-address" binding:"Required" json:"customer-address" csv:"customer-address"`
	CustomerCountry string `form:"customer-country" binding:"Required" json:"customer-country" csv:"customer-country"`

	// Deal Information
	Product           string `form:"product" binding:"Required" json:"product" csv:"product"`
	EstimatedDealSize string `form:"estimated-deal-size" json:"estimated-deal-size" csv:"estimated-deal-size"`
	ProjectTimeline   string `form:"project-timeline" json:"project-timeline" csv:"project-timeline"`
	CompetitorProduct string `form:"competitor-product" json:"competitor-product" csv:"competitor-product"`
	Notes             string `form:"notes" json:"notes" csv:"notes"`

	// Internal fields
	RegisteredOn OfferDate `form:"-" json:"-" csv:"registered-on"`
	GeoLocation  `form:"-" json:",inline"`
}

func (form *DealRegistrationInfo) Complete() {
	now := time.Now()
	form.RegisteredOn = NewOfferOfferDate(now)

	form.PartnerName = strings.TrimSpace(form.PartnerName)
	form.PartnerEmail = strings.TrimSpace(form.PartnerEmail)
	form.PartnerCompany = strings.TrimSpace(form.PartnerCompany)
	form.CustomerName = strings.TrimSpace(form.CustomerName)
	form.CustomerEmail = strings.TrimSpace(form.CustomerEmail)
	form.CustomerCompany = strings.TrimSpace(form.CustomerCompany)
}

func (form DealRegistrationInfo) Validate() error {
	if !strings.Contains(form.PartnerEmail, "@") {
		return fmt.Errorf("invalid partner email: %s", form.PartnerEmail)
	}
	if !strings.Contains(form.CustomerEmail, "@") {
		return fmt.Errorf("invalid customer email: %s", form.CustomerEmail)
	}
	return nil
}

func (s *Server) HandleDealRegistration(info *DealRegistrationInfo) error {
	// record in spreadsheet
	go func() {
		clients := []*DealRegistrationInfo{info}
		writer := gdrive.NewWriter(s.srvSheets, DealSpreadsheetId, "Deal Registration")
		err := gocsv.MarshalCSV(clients, writer)
		if err != nil {
			klog.Warningln(err)
			return
		}

		// mail incoming deals
		mailer := NewDealRegistrationMailer(info)
		fmt.Println("sending email for deal registration", info.CustomerCompany)
		err = mailer.SendMail(s.mg, MailIncomingDeals, "", nil)
		if err != nil {
			klog.Warningln(err)
			return
		}
	}()

	return nil
}
