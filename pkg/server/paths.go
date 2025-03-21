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

import "fmt"

func CACertificatesPath() string {
	return "certificates"
}

func EmailVerifiedPath(domain, email string) string {
	return fmt.Sprintf("domains/%s/emails/%s/verified", domain, email)
}

func EmailBannedPath(domain, email string) string {
	return fmt.Sprintf("domains/%s/emails/%s/banned", domain, email)
}

func EmailTokenPath(domain, email, token string) string {
	return fmt.Sprintf("domains/%s/emails/%s/tokens/%s", domain, email, token)
}

func AgreementPath(domain, product string) string {
	return fmt.Sprintf("domains/%s/products/%s/agreement.json", domain, product)
}

func (l ProductLicense) LicenseCertPath(cluster string) string {
	if l.ID > 0 {
		return fmt.Sprintf("id/%d/products/%s/clusters/%s/tls.crt", l.ID, l.Product, cluster)
	}
	return fmt.Sprintf("domains/%s/products/%s/clusters/%s/tls.crt", l.Domain, l.Product, cluster)
}

func (l ProductLicense) LicenseKeyPath(cluster string) string {
	if l.ID > 0 {
		return fmt.Sprintf("id/%d/products/%s/clusters/%s/tls.key", l.ID, l.Product, cluster)
	}
	return fmt.Sprintf("domains/%s/products/%s/clusters/%s/tls.key", l.Domain, l.Product, cluster)
}

func ProductAccessLogPath(domain, product, cluster, timestamp string) string {
	return fmt.Sprintf("domains/%s/products/%s/clusters/%s/accesslog/%s", domain, product, cluster, timestamp)
}

func FullLicenseIssueLogPath(domain, product, cluster, timestamp string) string {
	return fmt.Sprintf("domains/%s/products/%s/clusters/%s/full-license-issued/%s", domain, product, cluster, timestamp)
}

func EmailAccessLogPath(domain, email, product, timestamp string) string {
	return fmt.Sprintf("domains/%s/emails/%s/products/%s/accesslog/%s", domain, email, product, timestamp)
}
