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
	"strings"
)

// FormatRFC822Email generates an RFC 822 compliant email address string from a name and email address.
// It properly handles special characters in the name by quoting if necessary.
func FormatRFC822Email(name, email string) string {
	// Trim whitespace from inputs
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(email)

	// If no name is provided, return just the email address
	if name == "" {
		return email
	}

	// Check if the name needs quoting (contains special characters or spaces)
	needsQuoting := false
	for _, r := range name {
		if r <= 32 || r >= 127 || strings.ContainsRune("()<>@,;:\\\"[]", r) {
			needsQuoting = true
			break
		}
	}

	// If name needs quoting or contains spaces, enclose it in double quotes
	if needsQuoting || strings.Contains(name, " ") {
		// Escape any existing double quotes and backslashes
		escapedName := strings.ReplaceAll(name, "\\", "\\\\")
		escapedName = strings.ReplaceAll(escapedName, "\"", "\\\"")
		return fmt.Sprintf("\"%s\" <%s>", escapedName, email)
	}

	// If name is simple (no special chars or spaces), return unquoted format
	return fmt.Sprintf("%s <%s>", name, email)
}
