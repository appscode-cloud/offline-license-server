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

func NewBlockedLicenseMailer(info LicenseMailData) Mailer {
	src := `Hello,
FYI, an attempt was issue license for {{.Product}} from blocked email/domain {{.Email}}. Please review for further action.

Regards,
License server
`

	return Mailer{
		Sender:          MailLicenseSender,
		BCC:             MailLicenseTracker,
		ReplyTo:         MailSales,
		Subject:         fmt.Sprintf("[LICENSE_BLOCKED] plan:%s email:%s", info.Product, info.Email),
		Body:            src,
		params:          info,
		AttachmentBytes: nil,
		EnableTracking:  false,
	}
}
