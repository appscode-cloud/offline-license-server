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
	"time"
)

func (s *Server) IssueEnterpriseLicense(info LicenseForm, extendBy time.Duration, ff FeatureFlags) error {
	crtLicense, accesslog, err := IssueEnterpriseLicense(s.fs, s.certs, info, extendBy, ff)
	if err != nil {
		return err
	}

	{
		err = LogLicense(s.sheet, accesslog)
		if err != nil {
			return err
		}
	}

	// avoid sending emails for know test emails
	if !knowTestEmails.Has(info.Email) {
		mailer := NewEnterpriseLicenseMailer(LicenseMailData{
			LicenseForm: info,
			License:     string(crtLicense),
		})
		mailer.AttachmentBytes = map[string][]byte{
			fmt.Sprintf("%s-license-%s.txt", info.Product, info.Cluster): crtLicense,
		}
		err = mailer.SendMail(s.mg, info.Email, info.CC, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
