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

import "time"

const (
	TimestampFormat = "1/2/2006 15:04:05"
)

type Timestamp struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *Timestamp) MarshalCSV() (string, error) {
	return date.Time.UTC().Format(TimestampFormat), nil
}

// Convert the CSV string as internal date
func (date *Timestamp) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse(TimestampFormat, csv)
	return err
}
