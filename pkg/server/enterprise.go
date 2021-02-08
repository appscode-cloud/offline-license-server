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

func (s *Server) IssueEnterpriseLicense(info LicenseForm, extendBy time.Duration) error {
	if !IsEnterpriseProduct(info.Product) {
		return fmt.Errorf("%s is not an Enterprise product", info.Product)
	}

	domain := Domain(info.Email)

	if IsDisposableEmail(domain) {
		return fmt.Errorf("disposable email %s is not supported", info.Email)
	}

	if exists, err := s.fs.Exists(context.TODO(), EmailBannedPath(domain, info.Email)); err == nil && exists {
		return fmt.Errorf("email %s is banned", info.Email)
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
		return err
	}
	if exists {
		data, err := s.fs.ReadFile(context.TODO(), LicenseCertPath(license.Domain, license.Product, info.Cluster))
		if err != nil {
			return err
		}
		certs, err := cert.ParseCertsPEM(data)
		if err != nil {
			return err
		}
		if len(certs) > 1 {
			return fmt.Errorf("multiple certificates found in %s", LicenseCertPath(license.Domain, license.Product, info.Cluster))
		}
		ttl := certs[0].NotAfter.Sub(certs[0].NotBefore)
		if ttl > DefaultTTLForEnterpriseProduct {
			// if expires in next 14 days issue new license
			if time.Until(certs[0].NotAfter) < DefaultTTLForEnterpriseProduct {
				license.Agreement.ExpiryDate = metav1.NewTime(certs[0].NotAfter.Add(extendBy).UTC())
			} else {
				// Original license is > 14 days valid. Keep using that.
				crtLicense = cert.EncodeCertPEM(certs[0])
			}
		}
	}
	if len(crtLicense) == 0 {
		crtLicense, err = s.CreateLicense(*license, info.Cluster)
		if err != nil {
			return err
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
			return err
		}
		err = s.fs.WriteFile(context.TODO(), FullLicenseIssueLogPath(domain, info.Product, info.Cluster, timestamp), data)
		if err != nil {
			return err
		}

		err = LogLicense(s.sheet, accesslog)
		if err != nil {
			return err
		}
	}

	{
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
	}

	{
		// mark email as verified
		if exists, err := s.fs.Exists(context.TODO(), EmailVerifiedPath(domain, info.Email)); err == nil && !exists {
			err = s.fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, info.Email), []byte(timestamp))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
