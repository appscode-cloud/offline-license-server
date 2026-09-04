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

package server_test

import (
	"testing"

	"go.bytebuilders.dev/offline-license-server/pkg/server"
)

func TestKubeDBInquiryInfo_Validate_Spam(t *testing.T) {
	baseForm := func() server.KubeDBInquiryInfo {
		return server.KubeDBInquiryInfo{
			CustomerName:    "Jane Doe",
			CustomerEmail:   "jane.doe@example.com",
			CustomerCompany: "Example Inc",
			CustomerAddress: "123 Main St",
			CustomerCountry: "USA",
			Notes:           "Interested in a KubeDB Enterprise quote",
		}
	}

	tests := []struct {
		name     string
		mutate   func(*server.KubeDBInquiryInfo)
		wantSpam bool
	}{
		{
			name:     "legitimate inquiry",
			mutate:   func(f *server.KubeDBInquiryInfo) {},
			wantSpam: false,
		},
		{
			name: "bitcoin in notes",
			mutate: func(f *server.KubeDBInquiryInfo) {
				f.Notes = "Please pay in bitcoin"
			},
			wantSpam: true,
		},
		{
			name: "coinbase in notes, mixed case",
			mutate: func(f *server.KubeDBInquiryInfo) {
				f.Notes = "Send funds via CoinBase"
			},
			wantSpam: true,
		},
		{
			name: "bitcoin in company name",
			mutate: func(f *server.KubeDBInquiryInfo) {
				f.CustomerCompany = "Bitcoin Ventures LLC"
			},
			wantSpam: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := baseForm()
			tt.mutate(&form)

			if got := form.IsSpam(); got != tt.wantSpam {
				t.Errorf("IsSpam() = %v; want %v", got, tt.wantSpam)
			}

			err := form.Validate()
			if tt.wantSpam && err == nil {
				t.Errorf("Validate() = nil; want spam error")
			}
			if !tt.wantSpam && err != nil {
				t.Errorf("Validate() = %v; want nil", err)
			}
		})
	}
}
