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

func NewKubeDBQuotationMailer(params interface{}) Mailer {
	src := `Hello {{ .Name }},

Thanks for your interest in licensing KubeDB. We have prepared a quotation of the KubeDB PAYG license for you.

1. We offer usage based pricing for KubeDB PAYG edition similar to AWS or Google cloud etc. and it includes the Stash backup support for no additional fees. We have attached the quotation for KubeDB PAYG edition for your reference. We have also included a kubedb_pricing_table.pdf file that shows the management fees for various common database sizes.

2. KubeDB PAYG comes with a 14 day free trial. So, you don't need to purchase a license for ephemeral Kubernetes clusters (typically found in Dev or CI/CD environments).

3. The various support options are detailed in the kubedb-support-plans.pdf. We offer a Basic support level for free with our PAYG license. For SLA bound tickets, we charge extra and offer Gold / Platinum plans.

If you have any questions or concerns, please do not hesitate to contact us. If you have any technical questions, someone from our product team will get back to you.

Regards,
AppsCode Team

[![Website](https://cdn.appscode.com/images/website.png)](https://appscode.com) [![Linkedin](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/ln.png)](https://www.linkedin.com/company/appscode/) [![Twitter](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/tt.png)](https://twitter.com/AppsCodeHQ) [![Youtube](https://codetwocdn.azureedge.net/images/mail-signatures/generator-dm/pad-box/yt.png)](https://www.youtube.com/c/AppsCodeInc)
`
	return Mailer{
		Sender:          MailSales,
		BCC:             MailSales,
		ReplyTo:         MailSales,
		Subject:         "Re: KubeDB PAYG price quotation",
		Body:            src,
		params:          params,
		AttachmentBytes: nil,
		GDriveFiles: map[string]string{
			"kubedb_pricing_table.pdf": "1-RRVPczOoPQZ21-BICtabrkpcfhLfJrQ",
			"kubedb-support-plans.pdf": "1zDvN0KUcvKFrgY_0PZCj4lfxri6xnGgr",
		},
		EnableTracking: true,
	}
}
