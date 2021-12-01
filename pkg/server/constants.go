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

	"gomodules.xyz/sets"
)

const (
	AccountFolderId      = "1RBXgSR0jud5cpCqeC90fAdyb0Oaz7EIc"
	LicenseSpreadsheetId = "1evwv2ON94R38M-Lkrw8b6dpVSkRYHUWsNOuI7X0_-zA"
	DripSpreadsheetId    = "10Jx3-1Ww2UQ7xNjs9-CRvJX4iIA22EDu-EsLKoHp1hc"

	MailLicenseSender  = "license-issuer@mail.appscode.com"
	MailLicenseTracker = "issued-license-tracker@appscode.com"
	MailSupport        = "support@appscode.com"
	MailSales          = "sales@appscode.com"
	MailHello          = "hello@appscode.com"

	DefaultTTLForEnterpriseProduct     = 30 * 24 * time.Hour
	DefaultFullTTLForEnterpriseProduct = 365 * 24 * time.Hour
	DefaultTTLForCommunityProduct      = 365 * 24 * time.Hour

	LicenseIssuerName = "AppsCode Inc."
	LicenseBucket     = "licenses.appscode.com"

	WebinarSpreadsheetId  = "1VW9K1yRLw6IFnr4o9ZJqaEamBahfqnjfl79EHeAZBzg"
	WebinarScheduleFormat = "1/2/2006 15:04:05"
	WebinarScheduleSheet  = "Schedule"
	WebinarCalendarId     = "c_gccijq3fpvbsgg68le9tq37pqs@group.calendar.google.com"

	MailingList_Console    = "06a84456-bfdf-4edf-97c1-7e7d4ad48f67"
	MailingList_KubeDB     = "a5f00cb2-f398-4408-a13a-28b6db8a32ba"
	MailingList_Kubeform   = "cd797afa-04d4-45c8-86e0-642a59b2d7f4"
	MailingList_KubeVault  = "b0a46c28-43c3-4048-8059-c3897474b577"
	MailingList_Stash      = "3ab3161e-d02c-42cf-ad96-bb406620d693"
	MailingList_Voyager    = "6c6d1338-bb38-40f6-bab4-ff09c2f6e184"
	MailingList_Auditor    = "65e961d0-d60e-4bf5-8a22-3a5ec28a32c5"
	MailingList_Panopticon = "47ae2f13-5034-483e-be9a-682b32b39315"
)

var knowTestEmails = sets.NewString("1gtm@appscode.com")
var skipEmailDomains = sets.NewString("appscode.com")

type PlanInfo struct {
	DisplayName    string
	ProductLine    string
	TierName       string
	TwitterHandle  string
	QuickstartLink string
	Features       []string
	MailingLists   []string
}

// plan name => features
var supportedProducts = map[string]PlanInfo{
	"kubedb-community": {
		DisplayName:    "KubeDB",
		ProductLine:    "kubedb",
		TierName:       "community",
		TwitterHandle:  "KubeDB",
		QuickstartLink: "https://kubedb.com/docs/latest/",
		Features:       []string{"kubedb-community", "panopticon-community", "kubedb-monitoring-agent"},
		MailingLists:   []string{MailingList_KubeDB, MailingList_Stash, MailingList_Panopticon},
	},
	"kubedb-enterprise": {
		DisplayName:    "KubeDB",
		ProductLine:    "kubedb",
		TierName:       "enterprise",
		TwitterHandle:  "KubeDB",
		QuickstartLink: "https://kubedb.com/docs/latest/",
		Features:       []string{"kubedb-enterprise", "kubedb-community", "kubedb-autoscaler", "kubedb-ext-stash", "panopticon-enterprise", "kubedb-monitoring-agent"},
		MailingLists:   []string{MailingList_KubeDB, MailingList_Stash, MailingList_Panopticon},
	},
	"stash-community": {
		DisplayName:    "Stash",
		ProductLine:    "stash",
		TierName:       "community",
		TwitterHandle:  "KubeStash",
		QuickstartLink: "https://stash.run/docs/latest/",
		Features:       []string{"stash-community", "panopticon-community"},
		MailingLists:   []string{MailingList_Stash, MailingList_Panopticon},
	},
	"stash-enterprise": {
		DisplayName:    "Stash",
		ProductLine:    "stash",
		TierName:       "enterprise",
		TwitterHandle:  "KubeStash",
		QuickstartLink: "https://stash.run/docs/latest/",
		Features:       []string{"stash-enterprise", "stash-community", "kubedb-ext-stash", "panopticon-enterprise"},
		MailingLists:   []string{MailingList_Stash, MailingList_Panopticon},
	},
	"kubevault-community": {
		DisplayName:    "KubeVault",
		ProductLine:    "kubevault",
		TierName:       "community",
		TwitterHandle:  "KubeVault",
		QuickstartLink: "https://kubevault.com/docs/latest/",
		Features:       []string{"kubevault-community", "panopticon-community"},
		MailingLists:   []string{MailingList_KubeVault, MailingList_Panopticon},
	},
	"kubevault-enterprise": {
		DisplayName:    "KubeVault",
		ProductLine:    "kubevault",
		TierName:       "enterprise",
		TwitterHandle:  "KubeVault",
		QuickstartLink: "https://kubevault.com/docs/latest/",
		Features:       []string{"kubevault-enterprise", "kubevault-community", "panopticon-enterprise"},
		MailingLists:   []string{MailingList_KubeVault, MailingList_Panopticon},
	},
	"kubeform-community": {
		DisplayName:    "Kubeform",
		ProductLine:    "kubeform",
		TierName:       "community",
		TwitterHandle:  "Kubeform",
		QuickstartLink: "https://kubeform.com/docs/latest/",
		Features:       []string{"kubeform-community", "panopticon-community"},
		MailingLists:   []string{MailingList_Kubeform, MailingList_Panopticon},
	},
	"kubeform-enterprise": {
		DisplayName:    "Kubeform",
		ProductLine:    "kubeform",
		TierName:       "enterprise",
		TwitterHandle:  "Kubeform",
		QuickstartLink: "https://kubeform.com/docs/latest/",
		Features:       []string{"kubeform-enterprise", "kubeform-community", "panopticon-enterprise"},
		MailingLists:   []string{MailingList_Kubeform, MailingList_Panopticon},
	},
	"voyager-community": {
		DisplayName:    "Voyager",
		ProductLine:    "voyager",
		TierName:       "community",
		TwitterHandle:  "voyagermesh",
		QuickstartLink: "https://voyagermesh.com/docs/latest/",
		Features:       []string{"voyager-community", "panopticon-community"},
		MailingLists:   []string{MailingList_Voyager, MailingList_Panopticon},
	},
	"voyager-enterprise": {
		DisplayName:    "Voyager",
		ProductLine:    "voyager",
		TierName:       "enterprise",
		TwitterHandle:  "voyagermesh",
		QuickstartLink: "https://voyagermesh.com/docs/latest/",
		Features:       []string{"voyager-enterprise", "voyager-community", "panopticon-enterprise"},
		MailingLists:   []string{MailingList_Voyager, MailingList_Panopticon},
	},
	"console-enterprise": {
		ProductLine:  "console",
		TierName:     "enterprise",
		Features:     []string{"console-enterprise", "auditor-enterprise", "panopticon-enterprise", "cluster-connector"},
		MailingLists: []string{MailingList_Console, MailingList_Panopticon},
	},
	"auditor-enterprise": {
		ProductLine:  "auditor",
		TierName:     "enterprise",
		Features:     []string{"auditor-enterprise"},
		MailingLists: []string{MailingList_Auditor},
	},
	"panopticon-community": {
		DisplayName:    "Panopticon",
		ProductLine:    "panopticon",
		TierName:       "community",
		TwitterHandle:  "Kubeops",
		QuickstartLink: "https://blog.byte.builders/post/introducing-panopticon/",
		Features:       []string{"panopticon-community"},
		MailingLists:   []string{MailingList_Panopticon},
	},
	"panopticon-enterprise": {
		DisplayName:    "Panopticon",
		ProductLine:    "panopticon",
		TierName:       "enterprise",
		TwitterHandle:  "Kubeops",
		QuickstartLink: "https://blog.byte.builders/post/introducing-panopticon/",
		Features:       []string{"panopticon-enterprise", "panopticon-community"},
		MailingLists:   []string{MailingList_Panopticon},
	},
}
