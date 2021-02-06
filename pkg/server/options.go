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
	"os"

	"github.com/spf13/pflag"
)

type Options struct {
	Issuer string

	CertDir   string
	CertEmail string
	Hosts     []string
	Port      int
	EnableSSL bool

	GeoCityDatabase string

	LicenseBucket        string
	LicenseSpreadsheetId string

	// Your available domain names can be found here:
	// (https://app.mailgun.com/app/domains)
	MailgunDomain string

	// You can find the Private API Key in your Account Menu, under "Settings":
	// (https://app.mailgun.com/app/account/security)
	MailgunPrivateAPIKey string

	freshsalesHost     string
	freshsalesAPIToken string

	GoogleCredentialDir string
}

func NewOptions() *Options {
	cwd, _ := os.Getwd()
	return &Options{
		Issuer:               "",
		CertDir:              "certs",
		CertEmail:            "tamal@appscode.com",
		Hosts:                []string{"license-issuer.appscode.com"},
		Port:                 4000,
		GeoCityDatabase:      "",
		LicenseBucket:        LicenseBucket,
		LicenseSpreadsheetId: LicenseSpreadsheetId,
		MailgunDomain:        os.Getenv("MG_DOMAIN"),
		MailgunPrivateAPIKey: os.Getenv("MG_API_KEY"),
		freshsalesHost:       "https://appscode.freshsales.io",
		freshsalesAPIToken:   os.Getenv("CRM_API_TOKEN"),
		GoogleCredentialDir:  cwd,
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Issuer, "ssl.issuer", s.Issuer, "Name of License issuer")

	fs.StringVar(&s.CertDir, "ssl.cert-dir", s.CertDir, "Directory where certs are stored")
	fs.StringVar(&s.CertEmail, "ssl.email", s.CertEmail, "Email used by Let's Encrypt to notify about problems with issued certificates")
	fs.StringSliceVar(&s.Hosts, "ssl.hosts", s.Hosts, "Hosts for which certificate will be issued")
	fs.IntVar(&s.Port, "port", s.Port, "Port used when SSL is not enabled")
	fs.BoolVar(&s.EnableSSL, "ssl", s.EnableSSL, "Set true to enable SSL via Let's Encrypt")

	fs.StringVar(&s.GeoCityDatabase, "geo-city-database-file", s.GeoCityDatabase, "Path to GeoLite2-City.mmdb")

	fs.StringVar(&s.LicenseBucket, "bucket", s.LicenseBucket, "Name of GCS bucket used to store licenses")
	fs.StringVar(&s.LicenseSpreadsheetId, "spreadsheet-id", s.LicenseSpreadsheetId, "Google Spreadsheet Id used to store license issue log")

	fs.StringVar(&s.MailgunDomain, "mailgun.domain", s.MailgunDomain, "Mailgun domain")
	fs.StringVar(&s.MailgunPrivateAPIKey, "mailgun.api-key", s.MailgunPrivateAPIKey, "Mailgun private api key")

	fs.StringVar(&s.freshsalesHost, "freshsales.host", s.freshsalesHost, "Freshsales host url")
	fs.StringVar(&s.freshsalesAPIToken, "freshsales.token", s.freshsalesAPIToken, "Freshsales api token")

	fs.StringVar(&s.GoogleCredentialDir, "google.credential-dir", s.GoogleCredentialDir, "Directory used to store Google credential")
}
