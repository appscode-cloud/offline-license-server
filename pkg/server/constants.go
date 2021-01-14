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
	"time"
)

const (
	MailSender         = "license-issuer@mail.appscode.com"
	MailLicenseTracker = "issued-license-tracker@appscode.com"
	MailReplyTo        = "support@appscode.com"

	DefaultTTLForEnterpriseProduct     = 14 * 24 * time.Hour
	DefaultFullTTLForEnterpriseProduct = 365 * 24 * time.Hour
	DefaultTTLForCommunityProduct      = 365 * 24 * time.Hour

	LicenseIssuerName = "AppsCode Inc."
	LicenseBucket     = "licenses.appscode.com"
)

var supportedProducts = map[string][]string{
	"kubedb-community":     {"kubedb-community"},
	"kubedb-enterprise":    {"kubedb-enterprise", "kubedb-community", "kubedb-autoscaler"},
	"stash-community":      {"stash-community"},
	"stash-enterprise":     {"stash-enterprise", "stash-community"},
	"kubevault-community":  {"kubevault-community"},
	"kubevault-enterprise": {"kubevault-enterprise", "kubevault-community"},
	"kubeform-community":   {"kubeform-community"},
	"kubeform-enterprise":  {"kubeform-enterprise", "kubeform-community"},
}
