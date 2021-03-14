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
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/avct/uasurfer"
	"github.com/go-macaron/auth"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/gocarina/gocsv"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/sets"
	"gopkg.in/macaron.v1"
)

type WebinarSchedule struct {
	Title          string   `json:"title" csv:"Title" form:"title"`
	Schedule       DateTime `json:"schedule" csv:"Schedule" form:"schedule"`
	Summary        string   `json:"summary" csv:"Summary" form:"summary"`
	Speaker        string   `json:"speaker" csv:"Speaker" form:"speaker"`
	SpeakerBio     string   `json:"speaker_bio" csv:"Speaker Bio" form:"speaker_bio"`
	SpeakerPicture string   `json:"speaker_picture" csv:"Speaker Picture" form:"speaker_picture"`
}

type WebinarMeetingID struct {
	GoogleCalendarEventID string `json:"google_calendar_event_id" csv:"Google Calendar Event ID"`
	ZoomMeetingID         int    `json:"zoom_meeting_id" csv:"Zoom Meeting ID"`
	ZoomMeetingPassword   string `json:"zoom_meeting_password" csv:"Zoom Meeting Password"`
}

type WebinarInfo struct {
	WebinarSchedule
	WebinarMeetingID
}

type WebinarRegistrationForm struct {
	FirstName string `json:"first_name" csv:"First Name" form:"first_name"`
	LastName  string `json:"last_name" csv:"Last Name" form:"last_name"`
	Phone     string `json:"phone" csv:"Phone" form:"phone"`
	JobTitle  string `json:"job_title" csv:"Job Title" form:"job_title"`
	Company   string `json:"company" csv:"Company" form:"company"`
	WorkEmail string `json:"work_email" csv:"Work Email" form:"work_email"`

	ClusterProvider string `json:"cluster_provider" csv:"Cluster Provider" form:"cluster_provider"`
	ExperienceLevel string `json:"experience_level" csv:"Experience Level" form:"experience_level"`
	MarketingReach  string `json:"marketing_reach" csv:"Marketing Reach" form:"marketing_reach"`
}

type WebinarRegistrationEmail struct {
	WorkEmail string `json:"work_email" csv:"Work Email" form:"work_email"`
}

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format(WebinarScheduleFormat), nil
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(WebinarScheduleFormat, csv)
	return err
}

func (s *Server) RegisterWebinarAPI(m *macaron.Macaron) {
	m.Get("/_/webinars", func(ctx *macaron.Context, c cache.Cache, log *log.Logger) {
		key := ctx.Req.URL.Path
		out := c.Get(key)
		if out == nil {
			schedule, err := s.NextWebinarSchedule()
			if err != nil {
				ctx.Error(http.StatusInternalServerError, err.Error())
				return
			}
			out = schedule
			_ = c.Put(key, out, 60*60)
		} else {
			log.Println(key, "found")
		}
		ctx.JSON(http.StatusOK, out)
	})

	m.Post("/_/webinars/:date/register", binding.Bind(WebinarRegistrationForm{}), func(ctx *macaron.Context, form WebinarRegistrationForm) {
		date := ctx.Params("date")
		err := s.RegisterForWebinar(ctx, date, form)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		// ctx.JSON(http.StatusOK, form)
		ctx.Redirect("https://appscode.com", http.StatusSeeOther)
	})

	m.Get("/_/webinars/:date/emails", auth.Basic(os.Getenv("APPSCODE_PRICING_USERNAME"), os.Getenv("APPSCODE_PRICING_PASSWORD")), func(ctx *macaron.Context) {
		date := ctx.Params("date")
		attendees, err := s.ListWebinarAttendees(date)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, attendees)
	})
}

func (s *Server) ListWebinarAttendees(date string) ([]string, error) {
	reader, err := gdrive.NewColumnReader(s.sheetsService, WebinarSpreadsheetId, date, "Work Email")
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	attendees := []*WebinarRegistrationEmail{}
	if err := gocsv.UnmarshalCSV(reader, &attendees); err != nil { // Load attendees from file
		return nil, err
	}

	result := sets.NewString()
	for _, v := range attendees {
		result.Insert(v.WorkEmail)
	}
	return result.List(), nil
}

func (s *Server) NextWebinarSchedule() (*WebinarSchedule, error) {
	reader, err := gdrive.NewRowReader(s.sheetsService, WebinarSpreadsheetId, WebinarScheduleSheet, &gdrive.Filter{
		Header: "Schedule",
		By: func(column []interface{}) (int, error) {
			type TP struct {
				Schedule time.Time
				Pos      int
			}
			var upcoming []TP
			now := time.Now()
			for i, v := range column {
				// 3/11/2021 3:00:00
				t, err := time.Parse(WebinarScheduleFormat, v.(string))
				if err != nil {
					return -1, err
				}
				if t.After(now) {
					upcoming = append(upcoming, TP{
						Schedule: t,
						Pos:      i,
					})
				}
			}
			if len(upcoming) == 0 {
				return -1, io.EOF
			}
			sort.Slice(upcoming, func(i, j int) bool {
				return upcoming[i].Schedule.Before(upcoming[j].Schedule)
			})
			return upcoming[0].Pos, nil
		},
	})
	if err == io.EOF {
		return &WebinarSchedule{}, nil
	} else if err != nil {
		return nil, err
	}

	schedules := []*WebinarSchedule{}
	if err := gocsv.UnmarshalCSV(reader, &schedules); err != nil { // Load clients from file
		return nil, err
	}

	if len(schedules) > 0 {
		return schedules[0], nil
	}
	return &WebinarSchedule{}, nil
}

func (s *Server) RegisterForWebinar(ctx *macaron.Context, date string, form WebinarRegistrationForm) error {
	sheetName := date
	clients := []*WebinarRegistrationForm{
		&form,
	}
	writer := gdrive.NewWriter(s.sheetsService, WebinarSpreadsheetId, sheetName)
	err := gocsv.MarshalCSV(clients, writer)
	if err != nil {
		return err
	}

	// create zoom, google calendar event if not exists,
	// add attendant if google calendar meeting exists

	tdate, err := time.Parse("2006-1-2", date)
	if err != nil {
		return err
	}
	yw, mw, dw := tdate.Date()

	reader, err := gdrive.NewRowReader(s.sheetsService, WebinarSpreadsheetId, "Schedule", &gdrive.Filter{
		Header: "Schedule",
		By: func(values []interface{}) (int, error) {
			for i, v := range values {
				t2, err := time.Parse(WebinarScheduleFormat, v.(string))
				if err != nil {
					return -1, err
				}
				y2, m2, d2 := t2.Date()

				if yw == y2 && mw == m2 && dw == d2 {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})
	if err != nil {
		return err
	}

	meetings := []*WebinarInfo{}
	if err := gocsv.UnmarshalCSV(reader, &meetings); err != nil { // Load clients from file
		return err
	}

	var result *WebinarInfo
	if len(meetings) > 0 {
		result = meetings[0]
	}
	if result == nil {
		return fmt.Errorf("can't find webinar schedule")
	}

	{
		// record in CRM
		ua := uasurfer.Parse(ctx.Req.UserAgent())
		location := GeoLocation{
			IP: GetIP(ctx.Req.Request),
		}
		DecorateGeoData(s.geodb, &location)

		_ = s.noteEventWebinarRegistration(form, EventWebinarRegistration{
			BaseNoteDescription: freshsalesclient.BaseNoteDescription{
				Event: "webinar_registration",
				Client: freshsalesclient.ClientInfo{
					OS:     ua.OS.Name.StringTrimPrefix(),
					Device: ua.DeviceType.StringTrimPrefix(),
					Location: freshsalesclient.GeoLocation{
						IP:          location.IP,
						Timezone:    location.Timezone,
						City:        location.City,
						Country:     location.Country,
						Coordinates: location.Coordinates,
					},
				},
			},
			Webinar: WebinarRecord{
				Title:           result.Title,
				Schedule:        result.Schedule,
				Speaker:         result.Speaker,
				ClusterProvider: form.ClusterProvider,
				ExperienceLevel: form.ExperienceLevel,
				MarketingReach:  form.MarketingReach,
			},
		})
	}

	if result.GoogleCalendarEventID != "" {
		wats, err := gdrive.NewColumnReader(s.sheetsService, WebinarSpreadsheetId, sheetName, "Work Email")
		if err != nil {
			return err
		}
		atts := []*WebinarRegistrationEmail{}
		if err := gocsv.UnmarshalCSV(wats, &atts); err != nil { // Load clients from file
			return err
		}

		emails := make([]string, len(atts))
		for i, a := range atts {
			emails[i] = a.WorkEmail
		}
		return AddEventAttendants(s.calendarService, WebinarCalendarId, result.GoogleCalendarEventID, emails)
	}

	ww := gdrive.NewRowWriter(s.sheetsService, WebinarSpreadsheetId, "Schedule", &gdrive.Filter{
		Header: "Schedule",
		By: func(values []interface{}) (int, error) {
			for i, v := range values {
				t2, err := time.Parse(WebinarScheduleFormat, v.(string))
				if err != nil {
					return -1, err
				}
				y2, m2, d2 := t2.Date()

				if yw == y2 && mw == m2 && dw == d2 {
					return i, nil
				}
			}
			return -1, io.EOF
		},
	})

	meetinginfo, err := CreateZoomMeeting(s.calendarService, s.zc, WebinarCalendarId, s.zoomAccountEmail, &result.WebinarSchedule, 60*time.Minute, []string{
		form.WorkEmail,
	})
	if err != nil {
		return err
	}

	meetings2 := []*WebinarMeetingID{
		meetinginfo,
	}
	return gocsv.MarshalCSV(meetings2, ww)
}
