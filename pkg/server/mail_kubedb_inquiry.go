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

func NewKubeDBInquiryMailer(info *KubeDBInquiryInfo) mailer.Mailer {
	src := fmt.Sprintf(`Hi,
A new KubeDB inquiry has been submitted with the following details:

## Customer
- Name: %s
- Email: %s
- Company: %s
- Phone: %s
- Address: %s
- Country: %s

## Inquiry
- Estimated Database Memory: %s
- Kubernetes Setup: %s
- Support Plan: %s
- Project Timeline: %s
- Professional Services: %s
- Notes: %s

Regards,
KubeDB Inquiry System
`,
		info.CustomerName,
		info.CustomerEmail,
		info.CustomerCompany,
		info.CustomerPhone,
		info.CustomerAddress,
		info.CustomerCountry,
		info.EstimatedDatabaseMemory,
		info.KubernetesSetup,
		info.SupportPlan,
		info.ProjectTimeline,
		info.ProfessionalServices,
		info.Notes,
	)
	return mailer.Mailer{
		Sender:          MailSales,
		BCC:             "",
		ReplyTo:         info.CustomerEmail,
		Subject:         fmt.Sprintf("KubeDB Inquiry: %s", info.CustomerCompany),
		Body:            src,
		Params:          nil,
		AttachmentBytes: nil,
		GDriveFiles:     nil,
	}
}
