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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gomodules.xyz/cert"
	. "gomodules.xyz/email-providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Server) IssueEnterpriseLicense(info LicenseForm, extendBy time.Duration, ff FeatureFlags, sendMail bool) ([]byte, error) {
	if !IsEnterpriseProduct(info.Product) {
		return nil, fmt.Errorf("%s is not an Enterprise product", info.Product)
	}

	domain := Domain(info.Email)

	if IsDisposableEmail(domain) {
		return nil, fmt.Errorf("disposable email %s is not supported", info.Email)
	}

	if exists, err := s.fs.Exists(context.TODO(), EmailBannedPath(domain, info.Email)); err == nil && exists {
		return nil, fmt.Errorf("email %s is banned", info.Email)
	}

	// 1 yr domain license
	license := &ProductLicense{
		Domain:  domain,
		Product: info.Product,
		Agreement: &LicenseAgreement{
			NumClusters: 1, // is not used currently
			ExpiryDate:  metav1.NewTime(time.Now().Add(extendBy).UTC()),
		},
	}

	var crtLicense []byte
	exists, err := s.fs.Exists(context.TODO(), LicenseCertPath(license.Domain, license.Product, info.Cluster))
	if err != nil {
		return nil, err
	}
	if exists {
		data, err := s.fs.ReadFile(context.TODO(), LicenseCertPath(license.Domain, license.Product, info.Cluster))
		if err != nil {
			return nil, err
		}
		certs, err := cert.ParseCertsPEM(data)
		if err != nil {
			return nil, err
		}
		if len(certs) > 1 {
			return nil, fmt.Errorf("multiple certificates found in %s", LicenseCertPath(license.Domain, license.Product, info.Cluster))
		}

		if !certs[0].NotAfter.Before(license.Agreement.ExpiryDate.Time) {
			// Original license is sufficiently valid. Keep using that.
			crtLicense = cert.EncodeCertPEM(certs[0])
			license.Agreement.ExpiryDate = metav1.NewTime(certs[0].NotAfter.UTC())
		}
	}
	if len(crtLicense) == 0 {
		crtLicense, err = s.CreateLicense(info, *license, info.Cluster, ff)
		if err != nil {
			return nil, err
		}
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	{
		// record request
		accesslog := LogEntry{
			LicenseForm: info,
			Timestamp:   timestamp,
		}

		data, err := json.MarshalIndent(accesslog, "", "  ")
		if err != nil {
			return nil, err
		}
		err = s.fs.WriteFile(context.TODO(), FullLicenseIssueLogPath(domain, info.Product, info.Cluster, timestamp), data)
		if err != nil {
			return nil, err
		}

		//err = LogLicense(s.sheet, accesslog)
		//if err != nil {
		//	return nil, err
		//}
	}

	if sendMail {
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
				return nil, err
			}
		}
	}

	{
		// mark email as verified
		if exists, err := s.fs.Exists(context.TODO(), EmailVerifiedPath(domain, info.Email)); err == nil && !exists {
			err = s.fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, info.Email), []byte(timestamp))
			if err != nil {
				return nil, err
			}
		}
	}

	return crtLicense, nil
}
