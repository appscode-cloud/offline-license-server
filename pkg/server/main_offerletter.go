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

	"sigs.k8s.io/yaml"
)

func NewOfferLetterMailer(info *CandidateInfo, folderId string) Mailer {
	data, err := yaml.Marshal(info)
	if err != nil {
		data = []byte("Error: " + err.Error())
	}

	src := fmt.Sprintf(`Hi,
An offer letter has been generated for the following candidate:

%s

The generated offer letter can be found here: https://drive.google.com/drive/folders/%s

Regards,
Offer Letter Generator
`, string(data), folderId)
	return Mailer{
		Sender:          MailHR,
		BCC:             "",
		ReplyTo:         MailHR,
		Subject:         fmt.Sprintf("Offer letter generated for %s", info.Email),
		Body:            src,
		params:          nil,
		AttachmentBytes: nil,
		GDriveFiles:     nil,
		EnableTracking:  false,
	}
}
