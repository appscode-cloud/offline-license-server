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

import "strings"

type StringSlice []string

// Convert the internal date as CSV string
func (slice *StringSlice) MarshalCSV() (string, error) {
	if slice == nil {
		return "", nil
	}
	return strings.Join(*slice, ","), nil
}

// You could also use the standard Stringer interface
func (slice *StringSlice) String() string {
	if slice == nil {
		return ""
	}
	return strings.Join(*slice, ",")
}

// Convert the CSV string as internal date
func (slice *StringSlice) UnmarshalCSV(csv string) error {
	*slice = strings.Split(csv, ",")
	return nil
}
