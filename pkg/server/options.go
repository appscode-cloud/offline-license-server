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
	listmonkclient "gomodules.xyz/listmonk-client-go"
)

type Options struct {
	Issuer string

	CertDir   string
	CertEmail string
	Hosts     []string
	Port      int
	EnableSSL bool

	GeoCityDatabase string

	TaskDir string

	LicenseBucket        string
	LicenseSpreadsheetId string

	SMTPAddress  string
	SMTPUsername string
	SMTPPassword string

	listmonkHost     string
	listmonkUsername string
	listmonkPassword string

	GoogleCredentialDir string

	BlockedDomains []string
	BlockedEmails  []string

	EnableDripCampaign bool
}

func NewOptions() *Options {
	cwd, _ := os.Getwd()
	return &Options{
		Issuer:               "",
		CertDir:              "certs",
		CertEmail:            "tamal@appscode.com",
		Hosts:                []string{"license-issuer.appscode.com", "x.appscode.com"},
		Port:                 4000,
		GeoCityDatabase:      "",
		TaskDir:              "tasks",
		LicenseBucket:        LicenseBucket,
		LicenseSpreadsheetId: LicenseSpreadsheetId,
		SMTPAddress:          os.Getenv("SMTP_ADDRESS"),
		SMTPUsername:         os.Getenv("SMTP_USERNAME"),
		SMTPPassword:         os.Getenv("SMTP_PASSWORD"),
		listmonkHost:         listmonkclient.ListmonkProd,
		listmonkUsername:     os.Getenv("LISTMONK_USERNAME"),
		listmonkPassword:     os.Getenv("LISTMONK_PASSWORD"),
		GoogleCredentialDir:  cwd,
		EnableDripCampaign:   true,
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

	fs.StringVar(&s.TaskDir, "scheduler.db-dir", s.TaskDir, "Directory where task db files are stored")

	fs.StringVar(&s.LicenseBucket, "bucket", s.LicenseBucket, "Name of GCS bucket used to store licenses")
	fs.StringVar(&s.LicenseSpreadsheetId, "spreadsheet-id", s.LicenseSpreadsheetId, "Google Spreadsheet Id used to store license issue log")

	fs.StringVar(&s.SMTPAddress, "smtp.address", s.SMTPAddress, "SMTP server host:port")
	fs.StringVar(&s.SMTPUsername, "smtp.username", s.SMTPUsername, "SMTP username")
	fs.StringVar(&s.SMTPPassword, "smtp.password", s.SMTPPassword, "SMTP password")

	fs.StringVar(&s.listmonkHost, "listmonk.host", s.listmonkHost, "Listmonk host url")
	fs.StringVar(&s.listmonkUsername, "listmonk.username", s.listmonkUsername, "Listmonk username")
	fs.StringVar(&s.listmonkPassword, "listmonk.password", s.listmonkPassword, "Listmonk password")

	fs.StringVar(&s.GoogleCredentialDir, "google.credential-dir", s.GoogleCredentialDir, "Directory used to store Google credential")

	fs.StringSliceVar(&s.BlockedDomains, "blocked-domains", s.BlockedDomains, "Domains blocked from downloading license automatically")
	fs.StringSliceVar(&s.BlockedEmails, "blocked-emails", s.BlockedEmails, "Emails blocked from downloading license automatically")

	fs.BoolVar(&s.EnableDripCampaign, "drip-campaign", s.EnableDripCampaign, "Set true to enable drip campaign runner")
}
