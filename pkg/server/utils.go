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
	"gomodules.xyz/x/log"
)

func Domain(email string) string {
	idx := strings.LastIndexByte(email, '@')
	if idx == -1 {
		panic(fmt.Errorf("email %s is missing domain", email))
	}
	return email[idx+1:]
}

func IsEnterpriseProduct(product string) bool {
	return strings.HasSuffix(strings.ToLower(product), "-enterprise")
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

func DecorateGeoData(db *geoip2.Reader, entry *LogEntry) {
	if db == nil {
		return
	}
	ips := strings.Split(entry.IP, ",")
	if len(ips) == 0 {
		return
	}
	ip := net.ParseIP(strings.TrimSpace(ips[0]))
	if ip == nil {
		return
	}
	record, err := db.City(ip)
	if err != nil {
		log.Warningf("failed to detect geo data for ip %s. reason: %v", ip, err)
		return
	}

	entry.City = record.City.Names["en"]
	entry.Country = record.Country.IsoCode
	entry.Timezone = record.Location.TimeZone
	entry.Coordinates = fmt.Sprintf("%v,%v", record.Location.Latitude, record.Location.Longitude)
}
