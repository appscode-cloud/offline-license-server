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
	"fmt"
)

type QuotationEmailData struct {
	QuotationForm
	Offer    string // KubeDB, Stash
	FullPlan string // Pay-As-You-Go (PAYG), Enterprise
	Plan     string // PAYG, Enterprise
}

func NewQuotationMailer(info QuotationEmailData) Mailer {
	src := `Hello {{ .Name }},

Thanks for your interest in licensing {{.Offer}}.

1. We have attached a quotation for {{.Offer}} {{.FullPlan}} edition for your reference.

2. {{.Offer}} {{.Plan}} comes with a 30 day free trial. So, you don't need to purchase a license for ephemeral Kubernetes clusters (typically found in Dev or CI/CD environments).

3. The various support options are detailed in the attached quotation. We offer a Basic support level for free with our {{.Plan}} license. For SLA bound tickets, we charge extra and offer Gold / Platinum plans.

If you have any questions or concerns, please do not hesitate to contact us. If you have any technical questions, someone from our product team will get back to you.

Regards,
Team AppsCode

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`

	return Mailer{
		Sender:          MailSales,
		BCC:             MailSales,
		ReplyTo:         MailSales,
		Subject:         fmt.Sprintf("%s %s Quotation - %s", info.Offer, info.Plan, info.Company),
		Body:            src,
		params:          info,
		AttachmentBytes: nil,
		GDriveFiles:     nil,
		EnableTracking:  true,
	}
}
