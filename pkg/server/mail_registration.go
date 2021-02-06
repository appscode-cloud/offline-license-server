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

func NewRegistrationMailer(params interface{}) Mailer {
	src := `Hi,
Please use the token below to issue licenses using this email address.

{{.Token}}

Please let us know if you have any questions.

Regards,
AppsCode Team

[![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`
	return Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSupport,
		Subject:         "AppsCode license server token",
		Body:            src,
		params:          params,
		AttachmentBytes: nil,
		EnableTracking:  true,
	}
}
