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

	MailLicenseSender  = "license-issuer@mail.appscode.com"
	MailLicenseTracker = "issued-license-tracker@appscode.com"
	MailSupport        = "support@appscode.com"
	MailSales          = "sales@appscode.com"

	DefaultTTLForEnterpriseProduct     = 30 * 24 * time.Hour
	DefaultFullTTLForEnterpriseProduct = 365 * 24 * time.Hour
	DefaultTTLForCommunityProduct      = 365 * 24 * time.Hour

	LicenseIssuerName = "AppsCode Inc."
	LicenseBucket     = "licenses.appscode.com"

	WebinarSpreadsheetId  = "1VW9K1yRLw6IFnr4o9ZJqaEamBahfqnjfl79EHeAZBzg"
	WebinarScheduleFormat = "1/2/2006 15:04:05"
	WebinarScheduleSheet  = "Schedule"
	WebinarCalendarId     = "c_gccijq3fpvbsgg68le9tq37pqs@group.calendar.google.com"
)

var knowTestEmails = sets.NewString("1gtm@appscode.com")
var skipEmailDomains = sets.NewString("appscode.com")

const (
	MailingListSubscriptionURL = "https://listmonk-testing.appscode.com/subscription/form"

	MailingList_KubeDB    = "a5f00cb2-f398-4408-a13a-28b6db8a32ba"
	MailingList_Stash     = "3ab3161e-d02c-42cf-ad96-bb406620d693"
	MailingList_Kubeform  = "cd797afa-04d4-45c8-86e0-642a59b2d7f4"
	MailingList_KubeVault = "b0a46c28-43c3-4048-8059-c3897474b577"
)

type PlanInfo struct {
	Features     []string
	MailingLists []string
}

// plan name => features
var supportedProducts = map[string]PlanInfo{
	"kubedb-community": {
		Features:     []string{"kubedb-community"},
		MailingLists: []string{MailingList_KubeDB, MailingList_Stash},
	},
	"kubedb-enterprise": {
		Features:     []string{"kubedb-enterprise", "kubedb-community", "kubedb-autoscaler", "kubedb-ext-stash"},
		MailingLists: []string{MailingList_KubeDB, MailingList_Stash},
	},
	"stash-community": {
		Features:     []string{"stash-community"},
		MailingLists: []string{MailingList_Stash},
	},
	"stash-enterprise": {
		Features:     []string{"stash-enterprise", "stash-community", "kubedb-ext-stash"},
		MailingLists: []string{MailingList_Stash},
	},
	"kubevault-community": {
		Features:     []string{"kubevault-community"},
		MailingLists: []string{MailingList_KubeVault},
	},
	"kubevault-enterprise": {
		Features:     []string{"kubevault-enterprise", "kubevault-community"},
		MailingLists: []string{MailingList_KubeVault},
	},
	"kubeform-community": {
		Features:     []string{"kubeform-community"},
		MailingLists: []string{MailingList_Kubeform},
	},
	"kubeform-enterprise": {
		Features:     []string{"kubeform-enterprise", "kubeform-community"},
		MailingLists: []string{MailingList_Kubeform},
	},
}
