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
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	gdrive "gomodules.xyz/gdrive-utils"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type KubeDBInquiryInfo struct {
	// Customer Information
	CustomerName    string `form:"customer-name" binding:"Required" json:"customer-name" csv:"customer-name"`
	CustomerEmail   string `form:"customer-email" binding:"Required;Email" json:"customer-email" csv:"customer-email"`
	CustomerCompany string `form:"customer-company" binding:"Required" json:"customer-company" csv:"customer-company"`
	CustomerPhone   string `form:"customer-phone" json:"customer-phone" csv:"customer-phone"`
	CustomerAddress string `form:"customer-address" binding:"Required" json:"customer-address" csv:"customer-address"`
	CustomerCountry string `form:"customer-country" binding:"Required" json:"customer-country" csv:"customer-country"`

	// Deal Information
	EstimatedDatabaseMemory string `form:"estimated-database-memory" json:"estimated-database-memory" csv:"estimated-database-memory"`
	KubernetesSetup         string `form:"kubernetes-setup" json:"kubernetes-setup" csv:"kubernetes-setup"`
	SupportPlan             string `form:"support-plan" json:"support-plan" csv:"support-plan"`
	ProjectTimeline         string `form:"project-timeline" json:"project-timeline" csv:"project-timeline"`
	ProfessionalServices    string `form:"professional-services" json:"professional-services" csv:"professional-services"`
	Notes                   string `form:"notes" json:"notes" csv:"notes"`

	// Internal fields
	RegisteredOn OfferDate `form:"-" json:"-" csv:"registered-on"`
}

func (form *KubeDBInquiryInfo) Complete() {
	now := time.Now()
	form.RegisteredOn = NewOfferOfferDate(now)

	form.CustomerName = strings.TrimSpace(form.CustomerName)
	form.CustomerEmail = strings.TrimSpace(form.CustomerEmail)
	form.CustomerCompany = strings.TrimSpace(form.CustomerCompany)
}

func (form KubeDBInquiryInfo) Validate() error {
	if !strings.Contains(form.CustomerEmail, "@") {
		return fmt.Errorf("invalid customer email: %s", form.CustomerEmail)
	}
	return nil
}

func (s *Server) HandleKubeDBInquiry(info *KubeDBInquiryInfo) error {
	go func() {
		clients := []*KubeDBInquiryInfo{info}
		writer := gdrive.NewWriter(s.srvSheets, DealSpreadsheetId, "KubeDB Inquiry")
		err := gocsv.MarshalCSV(clients, writer)
		if err != nil {
			klog.Warningln(err)
			return
		}

		err = s.noteEventKubeDBInquiry(info)
		if err != nil {
			klog.Warningln(err)
			return
		}

		mailer := NewKubeDBInquiryMailer(info)
		fmt.Println("sending email for kubedb inquiry", info.CustomerCompany)
		err = mailer.SendMail(s.mg, MailIncomingDeals, "", nil)
		if err != nil {
			klog.Warningln(err)
			return
		}
	}()

	return nil
}

type EventKubeDBInquiry struct {
	freshsalesclient.BaseNoteDescription `json:",inline"`
	KubeDBInquiryInfo                    `json:",inline"`
}

func (s *Server) noteEventKubeDBInquiry(info *KubeDBInquiryInfo) error {
	fields := strings.Fields(info.CustomerName)
	var firstName, lastName string
	if len(fields) > 0 {
		firstName = strings.Join(fields[0:len(fields)-1], " ")
		lastName = fields[len(fields)-1]
	}

	var companyID int64
	companyResults, err := s.freshsales.Search(info.CustomerCompany, freshsalesclient.EntitySalesAccount)
	if err == nil && len(companyResults) > 0 {
		id, _ := strconv.ParseInt(companyResults[0].ID, 10, 64)
		companyID = id
	} else {
		account := &freshsalesclient.SalesAccount{
			Name:    info.CustomerCompany,
			Address: info.CustomerAddress,
			Country: info.CustomerCountry,
			Phone:   info.CustomerPhone,
		}
		newAccount, err := s.freshsales.CreateAccount(account)
		if err != nil {
			klog.Warningln(err)
		} else if newAccount != nil {
			companyID = newAccount.ID
		}
	}

	var contactID int64
	contactResult, err := s.freshsales.LookupByEmail(info.CustomerEmail, freshsalesclient.EntityContact)
	if err == nil && len(contactResult.Contacts.Contacts) > 0 {
		contact := contactResult.Contacts.Contacts[0]
		contactID = contact.ID

		var changed bool
		if contact.DisplayName != info.CustomerName {
			contact.DisplayName = info.CustomerName
			changed = true
		}
		if contact.WorkNumber != info.CustomerPhone {
			contact.WorkNumber = info.CustomerPhone
			changed = true
		}
		if contact.Address != info.CustomerAddress {
			contact.Address = info.CustomerAddress
			changed = true
		}
		if contact.Country != info.CustomerCountry {
			contact.Country = info.CustomerCountry
			changed = true
		}

		if changed {
			_, err = s.freshsales.UpdateContact(&contact)
			if err != nil {
				klog.Warningln(err)
			}
		}
	} else {
		contact := &freshsalesclient.Contact{
			Email:          info.CustomerEmail,
			DisplayName:    info.CustomerName,
			FirstName:      firstName,
			LastName:       lastName,
			WorkNumber:     info.CustomerPhone,
			Address:        info.CustomerAddress,
			Country:        info.CustomerCountry,
			SalesAccountID: companyID,
		}
		newContact, err := s.freshsales.CreateContact(contact)
		if err != nil {
			klog.Warningln(err)
		} else if newContact != nil {
			contactID = newContact.ID
		}
	}

	e := EventKubeDBInquiry{
		BaseNoteDescription: freshsalesclient.BaseNoteDescription{
			Event: "kubedb_inquiry",
		},
		KubeDBInquiryInfo: *info,
	}
	desc, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.freshsales.AddNote(contactID, freshsalesclient.EntityContact, string(desc))
	return err
}
