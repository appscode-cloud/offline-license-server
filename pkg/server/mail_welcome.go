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

	"gomodules.xyz/mailer"
)

func NewWelcomeMailer(info LicenseForm) mailer.Mailer {
	params := struct {
		LicenseForm
		ProductDisplayName  string
		IsEnterpriseProduct bool
	}{
		LicenseForm:         info,
		ProductDisplayName:  SupportedProducts[info.Product].DisplayName,
		IsEnterpriseProduct: IsEnterpriseProduct(info.Product),
	}

	src := fmt.Sprintf(`Hi {{.Name}},

Thanks for trying {{.ProductDisplayName}}. Our engineers can help you with any issues during the evaluation process. Please email %s with any questions regarding {{.ProductDisplayName}}.

{{ if not .IsEnterpriseProduct }}
We noticed that you are trying the Community Edition. We offer a 30 day FREE evaluation license for our {{.ProductDisplayName}} Enterprise product. The Enterprise version offers important features for Day-2 operations. If you need more time for evaluation, we should be able to extend the trial period.
{{ end }}

We look forward to hearing from you.

Regards,
Team AppsCode

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`, MailSupport)

	return mailer.Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSupport,
		Subject:         fmt.Sprintf("Welcome to %s", SupportedProducts[info.Product].DisplayName),
		Body:            src,
		Params:          params,
		AttachmentBytes: nil,
	}
}
