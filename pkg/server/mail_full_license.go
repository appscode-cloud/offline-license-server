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

import "fmt"

func NewEnterpriseLicenseMailer(info LicenseMailData) Mailer {
	src := fmt.Sprintf(`Hi {{.Name}},
Thanks for purchasing license for {{.Product}}. The full license for Kubernetes cluster {{.Cluster}} is attached with this email. 

%s
{{ .License | trim }}
%s

Please let us know if you have any questions.

Regards,
AppsCode Team

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`, "```", "```")

	return Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSupport,
		Subject:         fmt.Sprintf("%s License for cluster %s", info.Product, info.Cluster),
		Body:            src,
		params:          info,
		AttachmentBytes: nil,
		EnableTracking:  true,
	}
}
