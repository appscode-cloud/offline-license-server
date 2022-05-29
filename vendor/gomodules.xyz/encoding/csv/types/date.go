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
	DateFormat = "1/2/2006"
)

type Date struct {
	time.Time
}

// Convert the internal date as CSV string
func (d *Date) MarshalCSV() (string, error) {
	return d.Time.UTC().Format(DateFormat), nil
}

// Convert the CSV string as internal date
func (d *Date) UnmarshalCSV(csv string) (err error) {
	d.Time, err = time.Parse(DateFormat, csv)
	return err
}
