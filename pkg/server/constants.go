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
	MailSender         = "license-issuer@appscode.ninja"
	MailLicenseTracker = "issued-license-tracker@appscode.com"
	MailReplyTo        = "support@appscode.com"
)

const LicenseIssuerName = "AppsCode Inc."

const DefaultTTLForEnterpriseProduct = 14 * 24 * time.Hour
const DefaultTTLForCommunityProduct = 365 * 24 * time.Hour

const LicenseBucket = "appscode-licenses"
const LicenseBucketURL = "gs://" + LicenseBucket
const GoogleApplicationCredentials = "/home/tamal/AppsCode/credentials/license-issuer@appscode-domains.json"
