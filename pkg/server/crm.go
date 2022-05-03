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
	"strings"

	"github.com/gobuffalo/flect"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	"sigs.k8s.io/yaml"
)

func (s *Server) ensureCRMEntity(lead *freshsalesclient.Lead) (freshsalesclient.EntityType, int64, error) {
	result, err := s.freshsales.LookupByEmail(lead.Email, freshsalesclient.EntityLead, freshsalesclient.EntityContact)
	if err != nil {
		return "", 0, err
	}

	if len(result.Contacts.Contacts) > 0 {
		// contact found
		return freshsalesclient.EntityContact, result.Contacts.Contacts[0].ID, nil
	} else if len(result.Leads.Leads) > 0 {
		// lead found
		return freshsalesclient.EntityLead, result.Leads.Leads[0].ID, nil
	}

	// create lead
	lead, err = s.freshsales.CreateLead(lead)
	if err != nil {
		return "", 0, err
	}
	return freshsalesclient.EntityLead, lead.ID, nil
}

func (s *Server) createLead(email string, name string) *freshsalesclient.Lead {
	fields := strings.Fields(name)
	var firstname, lastname string
	if len(fields) > 0 {
		firstname = strings.Join(fields[0:len(fields)-1], " ")
		lastname = fields[len(fields)-1]
	}
	return &freshsalesclient.Lead{
		Email:       email,
		DisplayName: name,
		FirstName:   firstname,
		LastName:    lastname,
	}
}

type LicenseEventType string

const (
	EventTypeLicenseIssued  = "license_issued"
	EventTypeLicenseBlocked = "license_blocked"
)

func (s *Server) noteEventLicenseIssued(info LogEntry, event LicenseEventType) error {
	et, id, err := s.ensureCRMEntity(s.createLead(info.Email, info.Name))
	if err != nil {
		return err
	}

	// add note
	e := EventLicenseIssued{
		BaseNoteDescription: freshsalesclient.BaseNoteDescription{
			Event: string(event),
			Client: freshsalesclient.ClientInfo{
				OS:     info.UA.OS.Name.StringTrimPrefix(),
				Device: info.UA.DeviceType.StringTrimPrefix(),
				Location: freshsalesclient.GeoLocation{
					IP:          info.IP,
					Timezone:    info.Timezone,
					City:        info.City,
					Country:     info.Country,
					Coordinates: info.Coordinates,
				},
			},
		},
		License: LicenseRef{
			Product: info.Product,
			Cluster: info.Cluster,
		},
	}
	desc, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.freshsales.AddNote(id, et, string(desc))
	return err
}

func (s *Server) noteEventQuotation(form ProductQuotation, e EventQuotationGenerated) error {
	result, err := s.freshsales.LookupByEmail(form.Email, freshsalesclient.EntityLead, freshsalesclient.EntityContact)
	if err != nil {
		return err
	}

	var et freshsalesclient.EntityType
	var id int64
	if len(result.Contacts.Contacts) > 0 {
		// contact found
		et = freshsalesclient.EntityContact
		contact := result.Contacts.Contacts[0]
		id = contact.ID

		var changed bool
		if contact.DisplayName != form.Name {
			contact.DisplayName = form.Name
			changed = true
		}
		if contact.JobTitle != form.Title {
			contact.JobTitle = form.Title
			changed = true
		}
		if contact.WorkNumber != form.Telephone {
			contact.WorkNumber = form.Telephone
			changed = true
		}
		//if contact.com.Company.Name == "" {
		//	contact.Company.Name = form.Company
		//	changed = true
		//}

		if changed {
			_, err = s.freshsales.UpdateContact(&contact)
			if err != nil {
				return err
			}
		}
	} else if len(result.Leads.Leads) > 0 {
		// lead found
		et = freshsalesclient.EntityLead
		lead := result.Leads.Leads[0]
		id = lead.ID

		var changed bool
		if lead.DisplayName != form.Name {
			lead.DisplayName = form.Name
			changed = true
		}
		if lead.JobTitle != form.Title {
			lead.JobTitle = form.Title
			changed = true
		}
		if lead.WorkNumber != form.Telephone {
			lead.WorkNumber = form.Telephone
			changed = true
		}
		if lead.Company.Name != form.Company {
			lead.Company.Name = form.Company
			changed = true
		}

		if changed {
			_, err = s.freshsales.UpdateLead(&lead)
			if err != nil {
				return err
			}
		}
	} else {
		// create lead
		fields := strings.Fields(form.Name)
		lead := &freshsalesclient.Lead{
			Email:       form.Email,
			DisplayName: form.Name,
			FirstName:   strings.Join(fields[0:len(fields)-1], " "),
			LastName:    fields[len(fields)-1],
			JobTitle:    form.Title,
			WorkNumber:  form.Telephone,
			Company: freshsalesclient.Company{
				Name: form.Company,
			},
		}
		lead, err := s.freshsales.CreateLead(lead)
		if err != nil {
			return err
		}
		et = freshsalesclient.EntityLead
		id = lead.ID
	}

	// add note
	desc, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.freshsales.AddNote(id, et, string(desc))
	return err
}

// nolint:unused
func (s *Server) noteEventMailgun(email string, e EventMailgun) error {
	name := email
	idx := strings.LastIndex(email, "@")
	if idx >= 0 {
		name = email[:idx]
	}
	name = flect.Humanize(name)
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = flect.Titleize(name)

	et, id, err := s.ensureCRMEntity(s.createLead(email, name))
	if err != nil {
		return err
	}

	// add note
	desc, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.freshsales.AddNote(id, et, string(desc))
	return err
}

func (s *Server) noteEventWebinarRegistration(form WebinarRegistrationForm, e EventWebinarRegistration) error {
	result, err := s.freshsales.LookupByEmail(form.WorkEmail, freshsalesclient.EntityLead, freshsalesclient.EntityContact)
	if err != nil {
		return err
	}

	var et freshsalesclient.EntityType
	var id int64
	if len(result.Contacts.Contacts) > 0 {
		// contact found
		et = freshsalesclient.EntityContact
		contact := result.Contacts.Contacts[0]
		id = contact.ID

		var changed bool
		name := form.FirstName + " " + form.LastName
		if contact.DisplayName != name {
			contact.DisplayName = name
			changed = true
		}
		if contact.JobTitle != form.JobTitle {
			contact.JobTitle = form.JobTitle
			changed = true
		}
		if contact.WorkNumber != form.Phone {
			contact.WorkNumber = form.Phone
			changed = true
		}
		//if contact.Company.Name == "" {
		//	contact.Company.Name = form.Company
		//	changed = true
		//}

		if changed {
			_, err = s.freshsales.UpdateContact(&contact)
			if err != nil {
				return err
			}
		}
	} else if len(result.Leads.Leads) > 0 {
		// lead found
		et = freshsalesclient.EntityLead
		lead := result.Leads.Leads[0]
		id = lead.ID

		var changed bool
		name := form.FirstName + " " + form.LastName
		if lead.DisplayName != name {
			lead.DisplayName = name
			changed = true
		}
		if lead.JobTitle != form.JobTitle {
			lead.JobTitle = form.JobTitle
			changed = true
		}
		if lead.WorkNumber != form.Phone {
			lead.WorkNumber = form.Phone
			changed = true
		}
		if lead.Company.Name != form.Company {
			lead.Company.Name = form.Company
			changed = true
		}

		if changed {
			_, err = s.freshsales.UpdateLead(&lead)
			if err != nil {
				return err
			}
		}
	} else {
		// create lead
		lead := &freshsalesclient.Lead{
			Email:       form.WorkEmail,
			DisplayName: form.FirstName + " " + form.LastName,
			FirstName:   form.FirstName,
			LastName:    form.LastName,
			JobTitle:    form.JobTitle,
			WorkNumber:  form.Phone,
			Company: freshsalesclient.Company{
				Name: form.Company,
			},
		}
		lead, err := s.freshsales.CreateLead(lead)
		if err != nil {
			return err
		}
		et = freshsalesclient.EntityLead
		id = lead.ID
	}

	// add note
	desc, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	_, err = s.freshsales.AddNote(id, et, string(desc))
	return err
}
