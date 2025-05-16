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
	"crypto/x509"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	licenseapi "go.bytebuilders.dev/license-verifier/apis/licenses/v1alpha1"

	"github.com/pkg/errors"
	"gomodules.xyz/blobfs"
	"gomodules.xyz/cert"
	"gomodules.xyz/cert/certstore"
	. "gomodules.xyz/email-providers"
	godiacritics "gopkg.in/Regis24GmbH/go-diacritics.v2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// fs := blobfs.New("gs://licenses.appscode.com")
func GetCertStore(fs blobfs.Interface, issuer string) (*certstore.CertStore, error) {
	caCertPath := CACertificatesPath()
	issuerName := LicenseIssuerName
	if issuer != "" {
		caCertPath = path.Join(CACertificatesPath(), issuer)
		issuerName = issuer
	}
	certs := certstore.New(fs, caCertPath, 0, issuerName)
	err := certs.InitCA()
	if err != nil {
		return nil, err
	}
	return certs, nil
}

func IssueEnterpriseLicense(fs blobfs.Interface, certs *certstore.CertStore, info LicenseForm, extendBy time.Duration, ff licenseapi.FeatureFlags) ([]byte, *LogEntry, error) {
	if !IsEnterpriseProduct(info.Product()) {
		return nil, nil, fmt.Errorf("%s is not an Enterprise product", info.Product())
	}

	domain := Domain(info.Email)

	if IsDisposableEmail(domain) {
		return nil, nil, fmt.Errorf("disposable email %s is not supported", info.Email)
	}

	if exists, err := fs.Exists(context.TODO(), EmailBannedPath(domain, info.Email)); err == nil && exists {
		return nil, nil, fmt.Errorf("email %s is banned", info.Email)
	}

	// 1 yr domain license
	license := &ProductLicense{
		ID:      info.ID,
		Domain:  domain,
		Product: info.Product(),
		Agreement: &LicenseAgreement{
			NumClusters: 1, // is not used currently
			ExpiryDate:  metav1.NewTime(time.Now().Add(extendBy).UTC().Truncate(time.Second)),
		},
	}

	var crtLicense []byte
	exists, err := fs.Exists(context.TODO(), license.LicenseCertPath(info.Cluster))
	if err != nil {
		return nil, nil, err
	}
	if exists {
		data, err := fs.ReadFile(context.TODO(), license.LicenseCertPath(info.Cluster))
		if err != nil {
			return nil, nil, err
		}
		certs, err := cert.ParseCertsPEM(data)
		if err != nil {
			return nil, nil, err
		}
		if len(certs) > 1 {
			return nil, nil, fmt.Errorf("multiple certificates found in %s", license.LicenseCertPath(info.Cluster))
		}

		existingFeatureFlags := licenseapi.FeatureFlags{}
		for _, ff := range certs[0].Subject.Locality {
			parts := strings.SplitN(ff, "=", 2)
			if len(parts) == 2 {
				existingFeatureFlags[licenseapi.FeatureFlag(parts[0])] = parts[1]
			}
		}

		if !certs[0].NotAfter.Before(license.Agreement.ExpiryDate.Time) &&
			reflect.DeepEqual(existingFeatureFlags, ff) {

			// Original license is sufficiently valid. Keep using that.
			crtLicense = cert.EncodeCertPEM(certs[0])
			license.Agreement.ExpiryDate = metav1.NewTime(certs[0].NotAfter.UTC())
		}
	}
	if len(crtLicense) == 0 {
		crtLicense, err = CreateLicense(fs, certs, info, *license, info.Cluster, ff)
		if err != nil {
			return nil, nil, err
		}
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	accesslog := LogEntry{
		LicenseForm: info,
		Timestamp:   timestamp,
	}
	// only log for https://appscode.com/issue-license/
	if license.ID <= 0 {
		{
			// record request
			data, err := json.MarshalIndent(accesslog, "", "  ")
			if err != nil {
				return nil, nil, err
			}
			err = fs.WriteFile(context.TODO(), FullLicenseIssueLogPath(domain, info.Product(), info.Cluster, timestamp), data)
			if err != nil {
				return nil, nil, err
			}
		}

		{
			// mark email as verified
			if exists, err := fs.Exists(context.TODO(), EmailVerifiedPath(domain, info.Email)); err == nil && !exists {
				err = fs.WriteFile(context.TODO(), EmailVerifiedPath(domain, info.Email), []byte(timestamp))
				if err != nil {
					return nil, nil, err
				}
			}
		}
	}

	return crtLicense, &accesslog, nil
}

func CreateLicense(fs blobfs.Interface, certs *certstore.CertStore, info LicenseForm, license ProductLicense, cluster string, ff licenseapi.FeatureFlags) ([]byte, error) {
	// agreement, TTL
	sans := AltNames{
		DNSNames: []string{cluster},
		EmailAddresses: []string{
			fmt.Sprintf("%s <%s>", godiacritics.Normalize(info.Name), info.Email),
			info.Email,
		},
	}
	cfg := Config{
		CommonName:         getCN(sans),
		Country:            SupportedProducts[license.Product].ProductLine,
		Province:           SupportedProducts[license.Product].TierName,
		Organization:       SupportedProducts[license.Product].Features,
		OrganizationalUnit: license.Product, // plan
		Locality:           ff.ToSlice(),
		AltNames:           sans,
		Usages:             []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	now := time.Now()
	cfg.NotBefore = now
	if license.Agreement != nil {
		cfg.NotAfter = license.Agreement.ExpiryDate.UTC()
	} else if license.TTL != nil {
		cfg.NotAfter = now.Add(license.TTL.Duration).UTC()
	} else {
		return nil, apierrors.NewInternalError(fmt.Errorf("Missing license TTL")) // this should never happen
	}

	key, err := cert.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate private key")
	}
	crt, err := NewSignedCert(cfg, key, certs.CACert(), certs.CAKey())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate client certificate")
	}

	err = fs.WriteFile(context.TODO(), license.LicenseCertPath(cluster), cert.EncodeCertPEM(crt))
	if err != nil {
		return nil, err
	}
	err = fs.WriteFile(context.TODO(), license.LicenseKeyPath(cluster), cert.EncodePrivateKeyPEM(key))
	if err != nil {
		return nil, err
	}

	return cert.EncodeCertPEM(crt), nil
}
