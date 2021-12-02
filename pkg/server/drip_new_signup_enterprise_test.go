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
	"testing"
)

func TestNewEnterpriseSignupCampaign(t *testing.T) {
	dc := NewEnterpriseSignupCampaign(nil, nil)
	for idx, step := range dc.Steps {
		step.Mailer.Params = &SignupCampaignData{
			Name:                "Tamal Saha",
			Product:             "kubedb-enterprise",
			ProductDisplayName:  "KubeDB",
			IsEnterpriseProduct: true,
			TwitterHandle:       "KubeDB",
			QuickstartLink:      "https://kubedb.com/docs/latest/",
		}
		_, _, _, err := step.Mailer.Render()
		if err != nil {
			t.Errorf("NewEnterpriseSignupCampaign() STEP_%d failed, reason: %v", idx, err)
		}
	}
}
