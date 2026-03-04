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

import "gomodules.xyz/mailer"

func NewRegistrationMailer(params any) mailer.Mailer {
	src := `Hi,
Please use the token below to issue licenses using this email address.

{{.Token}}

Please let us know if you have any questions.

Regards,
Team AppsCode

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://cdn.appscode.com/images/ln.png)](https://www.linkedin.com/company/appscode/) [![X](https://cdn.appscode.com/images/tt.png)](https://x.com/AppsCodeHQ) [![Youtube](https://cdn.appscode.com/images/yt.png)](https://www.youtube.com/@appscode)
`
	return mailer.Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSupport,
		Subject:         "AppsCode license server token",
		Body:            src,
		Params:          params,
		AttachmentBytes: nil,
	}
}
