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

	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProductLicense struct {
	Domain    string            `json:"domain"`
	Product   string            `json:"product"`
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
	Product string `form:"product" binding:"Required" json:"product"`
	Cluster string `form:"cluster" binding:"Required" json:"cluster"`
	Tos     string `form:"tos" binding:"Required" json:"tos"`
	Sales   string `form:"sales" json:"sales"`
	Token   string `form:"token" json:"token"`
}

func (form LicenseForm) Describe() string {
	return fmt.Sprintf("%s license for cluster %s", form.Product, form.Cluster)
}

func (form LicenseForm) Validate() error {
	_, err := uuid.Parse(form.Cluster)
	if err != nil {
		return err
	}
	if _, found := supportedProducts[form.Product]; !found {
		return fmt.Errorf("unknown product: %s", form.Product)
	}
	if agree, _ := strconv.ParseBool(form.Tos); !agree {
		return fmt.Errorf("user must agree to terms and services")
	}
	return nil
}

type LogEntry struct {
	LicenseForm `json:",inline,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
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
		"Contact For Sales",
	}
}

func (info LogEntry) Data() []string {
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
		info.Sales,
	}
}
