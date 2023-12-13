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

	"gomodules.xyz/cert"
	"gomodules.xyz/mailer"
)

func NewEnterpriseLicenseMailer(info LicenseMailData) mailer.Mailer {
	var fromTimestamp, toTimeStamp string
	crts, err := cert.ParseCertsPEM([]byte(info.License))
	if err != nil {
		fromTimestamp = err.Error()
		toTimeStamp = err.Error()
	} else {
		for _, crt := range crts {
			fromTimestamp = crt.NotBefore.UTC().Format("02 Jan, 2006")
			toTimeStamp = crt.NotAfter.UTC().Format("02 Jan, 2006")
			break
		}
	}
	displayName := SupportedProducts[info.Product()].DisplayName

	src := fmt.Sprintf(`Hi {{.Name}},
Thanks for purchasing license for %s. The full license for Kubernetes cluster {{.Cluster}} is attached with this email.

Valid From: %s
Valid To: %s

%s
{{ .License | trim }}
%s

Please let us know if you have any questions.

Regards,
Team AppsCode

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`, displayName, fromTimestamp, toTimeStamp, "```", "```")

	return mailer.Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSupport,
		Subject:         fmt.Sprintf("%s License for cluster %s", displayName, info.Cluster),
		Body:            src,
		Params:          info,
		AttachmentBytes: nil,
	}
}
