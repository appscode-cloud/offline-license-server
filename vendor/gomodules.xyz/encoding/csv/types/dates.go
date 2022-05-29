/*
Copyright AppsCode Inc. and Contributors

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

package types

import (
	"sort"
	"strings"
	"time"
)

type Dates []time.Time

// Convert the internal date as CSV string
func (date *Dates) MarshalCSV() (string, error) {
	if date == nil {
		return "", nil
	}

	dates := make([]time.Time, 0, len(*date))
	for _, d := range *date {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})
	parts := make([]string, 0, len(*date))
	for _, d := range dates {
		parts = append(parts, d.UTC().Format(TimestampFormat))
	}
	return strings.Join(parts, ","), nil
}

// Convert the CSV string as internal date
func (date *Dates) UnmarshalCSV(csv string) (err error) {
	parts := strings.Split(csv, ",")

	dates := make([]time.Time, 0, len(parts))
	for _, part := range parts {
		d, err := time.Parse(TimestampFormat, part)
		if err != nil {
			return err
		}
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	*date = dates
	return nil
}
