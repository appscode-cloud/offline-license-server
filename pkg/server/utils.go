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
	"net"
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
	gdrive "gomodules.xyz/gdrive-utils"
	"gomodules.xyz/sets"
	"google.golang.org/api/sheets/v4"
	"k8s.io/klog/v2"
)

func DomainWithMXRecord(domain string) error {
	records, err := net.LookupMX(domain)
	if err != nil {
		return fmt.Errorf("no MX records for domain %s: %w", domain, err)
	}
	if len(records) == 0 {
		return fmt.Errorf("no MX records for domain %s", domain)
	}
	return nil
}

func IsEnterpriseProduct(product string) bool {
	return strings.HasSuffix(strings.ToLower(product), "-enterprise")
}

func IsPAYGProduct(product string) bool {
	if _, ok := templateIds[strings.ToLower(product)]; !ok {
		return false
	}
	return strings.HasSuffix(strings.ToLower(product), "-payg")
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func DecorateGeoData(db *geoip2.Reader, entry *GeoLocation) {
	if db == nil {
		return
	}
	ips := strings.Split(entry.IP, ",")
	if len(ips) == 0 {
		return
	}
	var ip net.IP
	if host, _, err := net.SplitHostPort(strings.TrimSpace(ips[0])); err == nil {
		ip = net.ParseIP(host)
	} else {
		ip = net.ParseIP(strings.TrimSpace(ips[0]))
	}
	if ip == nil {
		return
	}
	record, err := db.City(ip)
	if err != nil {
		klog.Warningf("failed to detect geo data for ip %s. reason: %v", ip, err)
		return
	}

	entry.IP = ip.String()
	entry.City = record.City.Names["en"]
	entry.Country = record.Country.IsoCode
	entry.Timezone = record.Location.TimeZone
	entry.Coordinates = fmt.Sprintf("%v,%v", record.Location.Latitude, record.Location.Longitude)
}

func ListExistingLicensees(srv *sheets.Service, spreadsheetId, sheetName, header string) sets.String {
	//const (
	//	sheetName = "License Issue Log"
	//	header    = "Email"
	//)
	reader, err := gdrive.NewColumnReader(srv, spreadsheetId, sheetName, header)
	if err != nil {
		panic(err)
	}
	cols, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	emails := sets.NewString()
	for _, row := range cols {
		emails.Insert(row...)
	}
	return emails
}
