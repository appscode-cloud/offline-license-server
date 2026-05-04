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

func NewDealRegistrationMailer(info *DealRegistrationInfo) mailer.Mailer {
	src := fmt.Sprintf(`Hi,
A new deal has been registered with the following details:

## Partner
- Name: %s
- Email: %s
- Company: %s
- Region: %s

## Customer
- Name: %s
- Email: %s
- Company: %s
- Phone: %s
- Address: %s
- Country: %s

## Deal
- Product: %s
- Kubernetes Setup: %s
- Estimated Deal Size: %s
- Estimated Database Memory: %s
- Estimated Kubernetes Nodes: %s
- Estimated Kubernetes Clusters: %s
- Project Timeline: %s
- Competitor Product: %s
- Notes: %s

Regards,
Deal Registration System
`,
		info.PartnerName,
		info.PartnerEmail,
		info.PartnerCompany,
		info.Region,
		info.CustomerName,
		info.CustomerEmail,
		info.CustomerCompany,
		info.CustomerPhone,
		info.CustomerAddress,
		info.CustomerCountry,
		info.Product,
		info.KubernetesSetup,
		info.EstimatedDealSize,
		info.EstimatedDatabaseMemory,
		info.EstimatedKubernetesNodes,
		info.EstimatedKubernetesClusters,
		info.ProjectTimeline,
		info.CompetitorProduct,
		info.Notes,
	)
	return mailer.Mailer{
		Sender:          MailSales,
		BCC:             "",
		ReplyTo:         info.PartnerEmail,
		Subject:         fmt.Sprintf("Deal Registration: %s - %s", info.CustomerCompany, info.Product),
		Body:            src,
		Params:          nil,
		AttachmentBytes: nil,
		GDriveFiles:     nil,
	}
}
