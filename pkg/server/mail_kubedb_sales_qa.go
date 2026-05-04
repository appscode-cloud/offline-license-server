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
	"encoding/json"
	"fmt"
	"strings"

	"gomodules.xyz/mailer"
)

type salesQANote struct {
	Section  int    `json:"section"`
	Question int    `json:"question"`
	Prompt   string `json:"prompt"`
	Signal   string `json:"signal"`
	Note     string `json:"note"`
}

func NewKubeDBSalesQAMailer(info *KubeDBSalesQAInfo) mailer.Mailer {
	notesMarkdown := "- No notes were submitted."
	if info.NotesJSON != "" {
		var notes []salesQANote
		if err := json.Unmarshal([]byte(info.NotesJSON), &notes); err != nil {
			notesMarkdown = fmt.Sprintf("- Failed to parse notes JSON: %v", err)
		} else if len(notes) > 0 {
			var b strings.Builder
			for _, n := range notes {
				line := fmt.Sprintf("- S%dQ%d", n.Section, n.Question)
				if n.Signal != "" {
					line += fmt.Sprintf(" (%s)", strings.ToUpper(n.Signal))
				}
				if n.Prompt != "" {
					line += fmt.Sprintf(": %s", n.Prompt)
				}
				b.WriteString(line)
				b.WriteString("\n")
				if n.Note != "" {
					b.WriteString(fmt.Sprintf("  - Note: %s\n", n.Note))
				}
			}
			notesMarkdown = strings.TrimSpace(b.String())
		}
	}

	src := fmt.Sprintf(`Hi,
A new KubeDB sales QA result has been submitted with the following details:

## Prospect
- Name: %s
- Email: %s
- Company: %s
- Title: %s
- Phone: %s
- Country: %s
- Address: %s
- Notes: %s

## Qualification Summary
- Distro: %s (%s)
- Verdict: %s
- Verdict Text: %s
- Hot: %d
- Warm: %d
- Cold: %d

## Notes
%s

Regards,
KubeDB Sales QA System
`,
		info.ContactName,
		info.ContactEmail,
		info.ContactCompany,
		info.ContactTitle,
		info.ContactPhone,
		info.ContactCountry,
		info.ContactAddress,
		info.ContactNotes,
		info.DistroLabel,
		info.Distro,
		info.Verdict,
		info.VerdictText,
		info.HotCount,
		info.WarmCount,
		info.ColdCount,
		notesMarkdown,
	)
	return mailer.Mailer{
		Sender:          MailSales,
		BCC:             "",
		ReplyTo:         info.ContactEmail,
		Subject:         fmt.Sprintf("KubeDB Sales QA: %s", info.ContactCompany),
		Body:            src,
		Params:          nil,
		AttachmentBytes: nil,
		GDriveFiles:     nil,
	}
}
