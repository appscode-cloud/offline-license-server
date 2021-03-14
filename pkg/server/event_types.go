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

import freshsalesclient "gomodules.xyz/freshsales-client-go"

type EventQuotationGenerated struct {
	freshsalesclient.BaseNoteDescription `json:",inline,omitempty"`

	Quotation     string `json:"quotation"`
	TemplateDoc   string `json:"template_doc"`
	TemplateDocId string `json:"template_doc_id"`
}

type EventLicenseIssued struct {
	freshsalesclient.BaseNoteDescription `json:",inline,omitempty"`

	License LicenseRef `json:"license"`
}

type LicenseRef struct {
	Product string `form:"product"`
	Cluster string `form:"cluster"`
}

type EventMailgun struct {
	freshsalesclient.BaseNoteDescription `json:",inline,omitempty"`

	Message Message `json:"message,omitempty"`
}

type Message struct {
	MessageID string `json:"message-id,omitempty"`
	Subject   string `json:"subject,omitempty"`
	Url       string `json:"url,omitempty"`
}

type EventWebinarRegistration struct {
	freshsalesclient.BaseNoteDescription `json:",inline,omitempty"`

	Webinar WebinarRecord `json:"webinar,omitempty"`
}

type WebinarRecord struct {
	Title    string   `json:"title" csv:"Title" form:"title"`
	Schedule DateTime `json:"schedule" csv:"Schedule" form:"schedule"`
	Speaker  string   `json:"speaker" csv:"Speaker" form:"speaker"`

	ClusterProvider string `json:"cluster_provider" csv:"Cluster Provider" form:"cluster_provider"`
	ExperienceLevel string `json:"experience_level" csv:"Experience Level" form:"experience_level"`
	MarketingReach  string `json:"marketing_reach" csv:"Marketing Reach" form:"marketing_reach"`
}
