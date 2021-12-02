/*
Copyright AppsCode Inc. and Contributors

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

package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/mailgun/mailgun-go/v4"
	"gomodules.xyz/encoding/json"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/sets"
	"google.golang.org/api/sheets/v4"
	"k8s.io/klog/v2"
)

type Contact struct {
	Email                   string    `csv:"email"`
	Data                    string    `csv:"data"` // json format
	Stop                    bool      `csv:"stop"`
	Step_0_Timestamp        Timestamp `csv:"s0_timestamp"`
	Step_0_WaitForCondition bool      `csv:"s0_wait_for_cond"`
	Step_0_Done             bool      `csv:"s0_done"`
	Step_1_Timestamp        Timestamp `csv:"s1_timestamp"`
	Step_1_WaitForCondition bool      `csv:"s1_wait_for_cond"`
	Step_1_Done             bool      `csv:"s1_done"`
	Step_2_Timestamp        Timestamp `csv:"s2_timestamp"`
	Step_2_WaitForCondition bool      `csv:"s2_wait_for_cond"`
	Step_2_Done             bool      `csv:"s2_done"`
	Step_3_Timestamp        Timestamp `csv:"s3_timestamp"`
	Step_3_WaitForCondition bool      `csv:"s3_wait_for_cond"`
	Step_3_Done             bool      `csv:"s3_done"`
	Step_4_Timestamp        Timestamp `csv:"s4_timestamp"`
	Step_4_WaitForCondition bool      `csv:"s4_wait_for_cond"`
	Step_4_Done             bool      `csv:"s4_done"`
}

type Timestamp struct {
	time.Time
}

func (date *Timestamp) MarshalCSV() (string, error) {
	if date.IsZero() {
		return "", nil
	}
	return date.Time.UTC().Format("01/02/2006 15:04:05"), nil
}

func (date *Timestamp) String() string {
	return date.Time.UTC().Format(time.RFC3339) // Redundant, just for example
}

func (date *Timestamp) UnmarshalCSV(csv string) (err error) {
	if csv != "" {
		date.Time, err = time.Parse("01/02/2006 15:04:05", csv)
	}
	return err
}

type DripCampaign struct {
	Name  string
	Steps []CampaignStep

	M             mailgun.Mailgun
	SheetService  *sheets.Service
	SpreadsheetId string
	SheetName     string
}

type CampaignStep struct {
	WaitTime time.Duration
	Mailer   Mailer
}

func (dc *DripCampaign) Prepare(c *Contact, t time.Time) {
	for idx, step := range dc.Steps {
		switch idx {
		case 0:
			c.Step_0_Timestamp = Timestamp{t.Add(step.WaitTime)}
		case 1:
			c.Step_1_Timestamp = Timestamp{t.Add(step.WaitTime)}
		case 2:
			c.Step_2_Timestamp = Timestamp{t.Add(step.WaitTime)}
		case 3:
			c.Step_3_Timestamp = Timestamp{t.Add(step.WaitTime)}
		case 4:
			c.Step_4_Timestamp = Timestamp{t.Add(step.WaitTime)}
		}
	}
}

func (dc *DripCampaign) AddContact(c Contact) error {
	dc.Prepare(&c, time.Now())

	si, err := gdrive.NewSpreadsheet(dc.SheetService, dc.SpreadsheetId)
	if err != nil {
		return err
	}
	_, err = si.EnsureSheet(dc.SheetName, nil)
	if err != nil {
		return err
	}

	w := gdrive.NewWriter(dc.SheetService, dc.SpreadsheetId, dc.SheetName)
	return gocsv.MarshalCSV([]*Contact{&c}, w)
}

func (dc *DripCampaign) Run(ctx context.Context) error {
	si, err := gdrive.NewSpreadsheet(dc.SheetService, dc.SpreadsheetId)
	if err != nil {
		return err
	}
	_, err = si.EnsureSheet(dc.SheetName, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := func() (e2 error) {
			defer func() {
				if r := recover(); r != nil {
					e2 = fmt.Errorf("panic: %v [recovered]", r)
				}
			}()
			e2 = dc.ProcessCampaign()
			return
		}()
		if err != nil {
			klog.ErrorS(err, "failed processing drip campaign", "name", dc.Name)
		} else {
			klog.InfoS("completed processing drip campaign", "name", dc.Name)
		}
		time.Sleep(1 * time.Hour)
	}
}

func (dc *DripCampaign) ProcessCampaign() error {
	now := time.Now()
	reader, err := gdrive.NewReader(dc.SheetService, dc.SpreadsheetId, dc.SheetName, 1)
	if err != nil {
		return err
	}
	var contacts []*Contact
	err = gocsv.UnmarshalCSV(reader, &contacts)
	if err != nil {
		return err
	}
	for _, c := range contacts {
		if c.Stop {
			continue
		}
		if !c.Step_0_Timestamp.IsZero() &&
			!c.Step_0_WaitForCondition &&
			now.After(c.Step_0_Timestamp.Time) &&
			!c.Step_0_Done {
			if err := dc.processStep(0, dc.Steps[0], *c); err != nil {
				klog.ErrorS(err, "failed to process campaign step", "email", c.Email, "step", 0)
			}
			continue
		}
		if !c.Step_1_Timestamp.IsZero() &&
			!c.Step_1_WaitForCondition &&
			now.After(c.Step_1_Timestamp.Time) &&
			!c.Step_1_Done {
			if err := dc.processStep(1, dc.Steps[1], *c); err != nil {
				klog.ErrorS(err, "failed to process campaign step", "email", c.Email, "step", 1)
			}
			continue
		}
		if !c.Step_2_Timestamp.IsZero() &&
			!c.Step_2_WaitForCondition &&
			now.After(c.Step_2_Timestamp.Time) &&
			!c.Step_2_Done {
			if err := dc.processStep(2, dc.Steps[2], *c); err != nil {
				klog.ErrorS(err, "failed to process campaign step", "email", c.Email, "step", 2)
			}
			continue
		}
		if !c.Step_3_Timestamp.IsZero() &&
			!c.Step_3_WaitForCondition &&
			now.After(c.Step_3_Timestamp.Time) &&
			!c.Step_3_Done {
			if err := dc.processStep(3, dc.Steps[3], *c); err != nil {
				klog.ErrorS(err, "failed to process campaign step", "email", c.Email, "step", 3)
			}
			continue
		}
		if !c.Step_4_Timestamp.IsZero() &&
			!c.Step_4_WaitForCondition &&
			now.After(c.Step_4_Timestamp.Time) &&
			!c.Step_4_Done {
			if err := dc.processStep(4, dc.Steps[4], *c); err != nil {
				klog.ErrorS(err, "failed to process campaign step", "email", c.Email, "step", 4)
			}
			continue
		}
	}
	return nil
}

func (dc *DripCampaign) processStep(stepIndex int, step CampaignStep, c Contact) error {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(c.Data), &params); err != nil {
		return err
	}

	m := step.Mailer
	m.Params = params
	err := m.SendMail(dc.M, c.Email, "", nil)
	if err != nil {
		return err
	}

	switch stepIndex {
	case 0:
		c.Step_0_Done = true
	case 1:
		c.Step_1_Done = true
	case 2:
		c.Step_2_Done = true
	case 3:
		c.Step_3_Done = true
	case 4:
		c.Step_4_Done = true
	}

	w := gdrive.NewRowWriter(dc.SheetService, dc.SpreadsheetId, dc.SheetName, &gdrive.Predicate{
		Header: "email",
		By: func(v []interface{}) (int, error) {
			for idx, entry := range v {
				if entry.(string) == c.Email {
					return idx, nil
				}
			}
			return -1, fmt.Errorf("missing email %s", c.Email)
		},
	})
	return gocsv.MarshalCSV([]*Contact{&c}, w)
}

func (dc *DripCampaign) ListAudiences() (sets.String, error) {
	reader, err := gdrive.NewColumnReader(dc.SheetService, dc.SpreadsheetId, dc.SheetName, "email")
	if err != nil {
		return nil, err
	}
	cols, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	emails := sets.NewString()
	for _, row := range cols {
		emails.Insert(row...)
	}
	return emails, nil
}
