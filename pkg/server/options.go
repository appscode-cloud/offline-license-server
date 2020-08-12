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
	CertDir   string
	CertEmail string
	Hosts     []string
	Port      int
	EnableSSL bool

	LicenseBucket string

	// Your available domain names can be found here:
	// (https://app.mailgun.com/app/domains)
	MailgunDomain string

	// You can find the Private API Key in your Account Menu, under "Settings":
	// (https://app.mailgun.com/app/account/security)
	MailgunPrivateAPIKey string

	MailSender         string
	MailLicenseTracker string
	MailReplyTo        string
}

func NewOptions() *Options {
	return &Options{
		CertDir:              "certs",
		CertEmail:            "tamal@appscode.com",
		Hosts:                []string{"license-issuer.appscode.com"},
		Port:                 4000,
		LicenseBucket:        LicenseBucket,
		MailgunDomain:        os.Getenv("MAILGUN_DOMAIN"),
		MailgunPrivateAPIKey: os.Getenv("MAILGUN_KEY"),
		MailSender:           MailSender,
		MailLicenseTracker:   MailLicenseTracker,
		MailReplyTo:          MailReplyTo,
	}
}

func (s *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.CertDir, "ssl.cert-dir", s.CertDir, "Directory where certs are stored")
	fs.StringVar(&s.CertEmail, "ssl.email", s.CertEmail, "Email used by Let's Encrypt to notify about problems with issued certificates")
	fs.StringSliceVar(&s.Hosts, "ssl.hosts", s.Hosts, "Hosts for which certificate will be issued")
	fs.IntVar(&s.Port, "port", s.Port, "Port used when SSL is not enabled")
	fs.BoolVar(&s.EnableSSL, "ssl", s.EnableSSL, "Set true to enable SSL via Let's Encrypt")

	fs.StringVar(&s.LicenseBucket, "bucket", s.LicenseBucket, "Name of GCS bucket used to store licenses")

	fs.StringVar(&s.MailgunDomain, "mailgun.domain", s.MailgunDomain, "Mailgun domain")
	fs.StringVar(&s.MailgunPrivateAPIKey, "mailgun.api-key", s.MailgunPrivateAPIKey, "Mailgun private api key")

	fs.StringVar(&s.MailSender, "mail.sender", s.MailgunDomain, "License sender mail")
	fs.StringVar(&s.MailLicenseTracker, "mail.license-tracker", s.MailLicenseTracker, "License tracker email")
	fs.StringVar(&s.MailReplyTo, "mail.reply-to", s.MailReplyTo, "Reply email for license emails")
}
