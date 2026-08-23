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
	"strings"
	"testing"

	"go.bytebuilders.dev/offline-license-server/pkg/server"
)

// Tos is intentionally left unset ("false") in every case below so that
// Validate() returns from the terms-of-service check before it ever reaches
// the network-dependent MX record lookup on the email domain. This keeps
// the test focused on, and isolated to, the cluster ID validation branch.
func TestLicenseFormValidate_ClusterRequirement(t *testing.T) {
	tests := []struct {
		name          string
		productAlias  string
		cluster       string
		wantClusterOK bool // true if the cluster check should be skipped/pass
	}{
		{
			name:          "kubedb-enterprise requires a valid cluster uuid",
			productAlias:  "kubedb-enterprise",
			cluster:       "",
			wantClusterOK: false,
		},
		{
			name:          "kubedb-enterprise rejects a malformed cluster id",
			productAlias:  "kubedb-enterprise",
			cluster:       "not-a-uuid",
			wantClusterOK: false,
		},
		{
			name:          "kubedb-enterprise accepts a valid cluster uuid",
			productAlias:  "kubedb-enterprise",
			cluster:       "5f2e2f62-d8b4-43d0-8d6d-d1592a9a76d4",
			wantClusterOK: true,
		},
		{
			name:          "postgres-enterprise does not require a cluster id",
			productAlias:  "postgres-enterprise",
			cluster:       "",
			wantClusterOK: true,
		},
		{
			name:          "postgres-enterprise ignores a malformed cluster id",
			productAlias:  "postgres-enterprise",
			cluster:       "not-a-uuid",
			wantClusterOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := server.LicenseForm{
				Name:         "Jane Doe",
				Email:        "jane.doe@example.com",
				ProductAlias: tt.productAlias,
				Cluster:      tt.cluster,
				Tos:          "false",
			}

			err := form.Validate()
			if err == nil {
				t.Fatalf("Validate() = nil; want an error since tos is not accepted")
			}

			isClusterErr := strings.Contains(err.Error(), "UUID")
			if isClusterErr == tt.wantClusterOK {
				t.Errorf("Validate() error = %q; wantClusterOK=%v", err, tt.wantClusterOK)
			}
		})
	}
}
