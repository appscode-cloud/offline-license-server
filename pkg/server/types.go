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
	"strconv"

	"github.com/avct/uasurfer"
	"github.com/google/uuid"
	. "gomodules.xyz/email-providers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProductLicense struct {
	Domain    string            `json:"domain"`
	Product   string            `json:"product"` // This is now called plan in a parsed LicenseInfo
	TTL       *metav1.Duration  `json:"ttl,omitempty"`
	Agreement *LicenseAgreement `json:"agreement,omitempty"`
}

type LicenseAgreement struct {
	NumClusters int         `json:"num_clusters"`
	ExpiryDate  metav1.Time `json:"expiry_date"`
}

type RegisterRequest struct {
	Email string `form:"email" binding:"Required;Email" json:"email"`
}

type LicenseForm struct {
	Name    string `form:"name" binding:"Required" json:"name"`
	Email   string `form:"email" binding:"Required;Email" json:"email"`
	CC      string `form:"cc" json:"cc"`
	Product string `form:"product" binding:"Required" json:"product"` // This is now called plan in a parsed LicenseInfo
	Cluster string `form:"cluster" binding:"Required" json:"cluster"`
	Tos     string `form:"tos" binding:"Required" json:"tos"`
	Token   string `form:"token" json:"token"`
}

type LicenseMailData struct {
	LicenseForm `json:",inline,omitempty"`
	License     string
}

func (form LicenseForm) Validate() error {
	_, err := uuid.Parse(form.Cluster)
	if err != nil {
		return err
	}
	if _, found := SupportedProducts[form.Product]; !found {
		return fmt.Errorf("unknown product: %s", form.Product)
	}
	if agree, _ := strconv.ParseBool(form.Tos); !agree {
		return fmt.Errorf("user must agree to terms and services")
	}
	return nil
}

type LogEntry struct {
	LicenseForm `json:",inline,omitempty"`
	GeoLocation `json:",inline,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	UA          *uasurfer.UserAgent `json:"-"`
}

type GeoLocation struct {
	IP          string `json:"ip,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	Coordinates string `json:"coordinates,omitempty"`
}

func (_ LogEntry) Headers() []string {
	return []string{
		"Domain",
		"Name",
		"Email",
		"Product",
		"Cluster",
		"Timestamp",
		"IP",
		"Timezone",
		"City",
		"Country",
		"Coordinates",
		"Client OS",
		"Client Device",
	}
}

func (info LogEntry) Data() []string {
	var clientOS, clientDevice string
	if info.UA != nil {
		clientOS = info.UA.OS.Name.StringTrimPrefix()
		clientDevice = info.UA.DeviceType.StringTrimPrefix()
	}
	return []string{
		Domain(info.Email),
		info.Name,
		info.Email,
		info.Product,
		info.Cluster,
		info.Timestamp,
		info.IP,
		info.Timezone,
		info.City,
		info.Country,
		info.Coordinates,
		clientOS,
		clientDevice,
	}
}
