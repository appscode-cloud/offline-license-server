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

/*
import (
	"encoding/json"
	"fmt"
	"net/http"


	"github.com/mailgun/mailgun-go/v4"
	"github.com/mailgun/mailgun-go/v4/events"
	freshsalesclient "gomodules.xyz/freshsales-client-go"
)

func (s *Server) HandleMailgunWebhook(w http.ResponseWriter, r *http.Request) {
	var payload mailgun.WebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		fmt.Printf("decode JSON error: %s", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	verified, err := s.mg.VerifyWebhookSignature(payload.Signature)
	if err != nil {
		fmt.Printf("verify error: %s\n", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	if !verified {
		w.WriteHeader(http.StatusNotAcceptable)
		fmt.Printf("failed verification %+v\n", payload.Signature)
		return
	}

	fmt.Printf("Verified Signature\n")

	// Parse the raw event to extract the

	e, err := mailgun.ParseEvent(payload.EventData)
	if err != nil {
		fmt.Printf("parse event error: %s\n", err)
		return
	}

	switch event := e.(type) {
	case *events.Opened:
		_ = s.noteEventMailgun(event.Recipient, EventMailgun{
			BaseNoteDescription: freshsalesclient.BaseNoteDescription{
				Event: event.Name,
				Client: freshsalesclient.ClientInfo{
					OS:     event.ClientInfo.ClientOS,
					Device: event.ClientInfo.DeviceType,
					Location: freshsalesclient.GeoLocation{
						City:    event.GeoLocation.City,
						Country: event.GeoLocation.Country,
					},
				},
			},
			Message: Message{
				MessageID: event.Message.Headers.MessageID,
				Subject:   event.Message.Headers.Subject,
			},
		})
		fmt.Printf("Email opened: %s\n", event.Message.Headers.MessageID)
	case *events.Clicked:
		_ = s.noteEventMailgun(event.Recipient, EventMailgun{
			BaseNoteDescription: freshsalesclient.BaseNoteDescription{
				Event: event.Name,
				Client: freshsalesclient.ClientInfo{
					OS:     event.ClientInfo.ClientOS,
					Device: event.ClientInfo.DeviceType,
					Location: freshsalesclient.GeoLocation{
						City:    event.GeoLocation.City,
						Country: event.GeoLocation.Country,
					},
				},
			},
			Message: Message{
				MessageID: event.Message.Headers.MessageID,
				Subject:   event.Message.Headers.Subject,
				Url:       event.Url,
			},
		})
		fmt.Printf("Link clicked: %s\n", event.Url)
	}
}
*/
