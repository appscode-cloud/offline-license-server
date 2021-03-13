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
	"strings"
	"time"

	"github.com/himalayan-institute/zoom-lib-golang"
	"github.com/k3a/html2text"
	passgen "gomodules.xyz/password-generator"
	"gomodules.xyz/pointer"
	"gomodules.xyz/sets"
	"google.golang.org/api/calendar/v3"
)

func CreateZoomMeeting(srv *calendar.Service, zc *zoom.Client, calendarId, zoomEmail string, schedule *WebinarSchedule, duration time.Duration, attendees []string) (*WebinarMeetingID, error) {
	user, err := zc.GetUser(zoom.GetUserOpts{EmailOrID: zoomEmail})
	if err != nil {
		return nil, fmt.Errorf("failed to get zoom user: %v", err)
	}

	meeting, err := zc.CreateMeeting(zoom.CreateMeetingOptions{
		HostID: user.ID,
		Topic:  schedule.Title,
		Type:   zoom.MeetingTypeScheduled,
		StartTime: &zoom.Time{
			Time: schedule.Schedule.Time,
		},
		Duration:       25,
		Timezone:       schedule.Schedule.Location().String(),
		Password:       passgen.GenerateForCharset(10, passgen.AlphaNum),
		Agenda:         html2text.HTML2Text(schedule.Summary),
		TrackingFields: nil,
		Settings: zoom.MeetingSettings{
			HostVideo:         false,
			ParticipantVideo:  false,
			ChinaMeeting:      false,
			IndiaMeeting:      false,
			JoinBeforeHost:    true,
			MuteUponEntry:     true,
			Watermark:         false,
			UsePMI:            false,
			ApprovalType:      zoom.ApprovalTypeManuallyApprove,
			RegistrationType:  zoom.RegistrationTypeRegisterEachTime,
			Audio:             zoom.AudioBoth,
			AutoRecording:     zoom.AutoRecordingLocal,
			CloseRegistration: false,
			WaitingRoom:       false,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create meeting: %v", err)
	}

	//var phones []string
	//for _, num := range meeting.Settings.GobalDialInNumbers {
	//	if num.Country == "US" && num.Type == "toll" {
	//		phones = append(phones, num.Number)
	//	}
	//}

	atts := make([]*calendar.EventAttendee, len(attendees))
	for _, email := range attendees {
		atts = append(atts, &calendar.EventAttendee{
			Email: email,
		})
	}

	event := &calendar.Event{
		Summary:     "AppsCode Webinar: " + schedule.Title,
		Description: html2text.HTML2Text(schedule.Summary),
		Start: &calendar.EventDateTime{
			DateTime: schedule.Schedule.UTC().Format(time.RFC3339),
			TimeZone: schedule.Schedule.Location().String(),
		},
		End: &calendar.EventDateTime{
			DateTime: schedule.Schedule.Add(duration).Format(time.RFC3339),
			TimeZone: schedule.Schedule.Location().String(),
		},
		GuestsCanInviteOthers:   pointer.TrueP(),
		GuestsCanModify:         false,
		GuestsCanSeeOtherGuests: pointer.FalseP(),
		Attendees:               atts,
		ConferenceData: &calendar.ConferenceData{
			ConferenceId: fmt.Sprintf("%d", meeting.ID),
			ConferenceSolution: &calendar.ConferenceSolution{
				IconUri: "https://lh3.googleusercontent.com/ugWKRyPiOCwjn5jfaoVgC-O80F3nhKH1dKMGsibXvGV1tc6pGXLOJk9WO7dwhw8-Xl9IwkKZEFDbeMDgnx-kf8YGJZ9uhrJMK9KP8-ISybmbgg1LK121obq2o5ML0YugbWh-JevWMu4FxxTKzM2r68bfDG_NY-BNnHSG7NcOKxo-RE7dfObk3VkycbRZk_GUK_1UUI0KitNg7HBfyqFyxIPOmds0l-h2Q1atWtDWLi29n_2-s5uw_kV4l2KeeaSGws_x8h0zsYWLDP5wdKWwYMYiQD2AFM32SHJ4wLAcAKnwoZxUSyT_lWFTP0PHJ6PwETDGNZjmwh3hD6Drn7Z3mnX662S6tkoPD92LtMDA0eNLr6lg-ypI2fhaSGKOeWFwA5eFjds7YcH-axp6cleuiEp05iyPO8uqtRDRMEqQhPaiRTcw7pY5UAVbz2yXbMLVofojrGTOhdvlYvIdDgBOSUkqCshBDV4A2TJyDXxFjpSYaRvwwWIT0JgrIxLpBhnyd3_w6m5My5QtgBJEt_S2Dq4bXwCAA7VcRiD61WmDyHfU3dBiWQUNjcH39IKT9V1fbUcUkfDPM_AGjp7pwgG3w9yDemGi1OGlRXS4pU7UwF24c2dozrmaK17iPaExn0cmIgtBnFUXReY48NI8h2MNd_QysNMWYNYbufoPD7trSu6nS39wlUDQer2V",
				Key: &calendar.ConferenceSolutionKey{
					Type: "addOn",
				},
				Name: "Zoom Meeting",
			},
			EntryPoints: []*calendar.EntryPoint{
				{
					EntryPointType: "video",
					Label:          strings.TrimPrefix(meeting.JoinURL, "https://"),
					MeetingCode:    fmt.Sprintf("%d", meeting.ID),
					Passcode:       meeting.Password,
					Uri:            meeting.JoinURL,
				},
				//{
				//	EntryPointType: "phone",
				//	Label:          phones[0],
				//	RegionCode:     "US",
				//	Passcode:       fmt.Sprintf("%d", meeting.Password),
				//	Uri:            fmt.Sprintf("tel:%s", strings.Join(phones, ",")),
				//},
				//{
				//	EntryPointType: "more",
				//	Uri:            "https://us02web.zoom.us/u/kp0VS4U41",
				//},
			},
			// Notes:              "",
			Parameters: &calendar.ConferenceParameters{
				AddOnParameters: &calendar.ConferenceParametersAddOnParameters{
					Parameters: map[string]string{
						"meetingCreatedBy": user.Email,
						"meetingType":      fmt.Sprintf("%d", meeting.Type),
						"meetingUuid":      meeting.UUID,
						"realMeetingId":    fmt.Sprintf("%d", meeting.ID),
					},
				},
			},
		},
	}

	event, err = srv.Events.Insert(calendarId, event).ConferenceDataVersion(1).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to create event. %v\n", err)
	}

	return &WebinarMeetingID{
		GoogleCalendarEventID: event.Id,
		ZoomMeetingID:         meeting.ID,
		ZoomMeetingPassword:   meeting.Password,
	}, nil
}

func AddEventAttendants(srv *calendar.Service, calendarId, eventId string, emails []string) error {
	sortEmails := sets.NewString(emails...)

	e2, err := srv.Events.Get(calendarId, eventId).Do()
	if err != nil {
		return err
	}
	for _, a := range e2.Attendees {
		sortEmails.Insert(a.Email)
	}

	attendees := make([]*calendar.EventAttendee, len(emails))
	for i, email := range sortEmails.List() {
		attendees[i] = &calendar.EventAttendee{
			Email: email,
		}
	}
	event := &calendar.Event{
		Id:        eventId,
		Attendees: attendees,
	}
	_, err = srv.Events.Patch(calendarId, event.Id, event).ConferenceDataVersion(1).Do()
	return err
}
