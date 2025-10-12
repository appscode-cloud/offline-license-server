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

package server_test

import (
	"testing"

	"go.bytebuilders.dev/offline-license-server/pkg/server"
)

func TestFormatRFC822Email(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		// Basic case: simple name and email
		{
			name:     "John Doe",
			email:    "john.doe@example.com",
			expected: "\"John Doe\" <john.doe@example.com>",
		},
		// No name provided
		{
			name:     "",
			email:    "john.doe@example.com",
			expected: "john.doe@example.com",
		},
		// Name with special characters requiring quoting
		{
			name:     "John \"The Boss\" Doe",
			email:    "john.doe@example.com",
			expected: "\"John \\\"The Boss\\\" Doe\" <john.doe@example.com>",
		},
		// Name with special characters (comma, semicolon)
		{
			name:     "Doe, John",
			email:    "john.doe@example.com",
			expected: "\"Doe, John\" <john.doe@example.com>",
		},
		// Name with non-ASCII characters
		{
			name:     "José Müller",
			email:    "jose.muller@example.com",
			expected: "\"José Müller\" <jose.muller@example.com>",
		},
		// Simple name without spaces (no quotes needed)
		{
			name:     "John",
			email:    "john@example.com",
			expected: "John <john@example.com>",
		},
		// Name with leading/trailing spaces
		{
			name:     "  Alice Bob  ",
			email:    "alice.bob@example.com",
			expected: "\"Alice Bob\" <alice.bob@example.com>",
		},
		// Email with mixed case
		{
			name:     "Alice",
			email:    "Alice.Bob@Example.Com",
			expected: "Alice <Alice.Bob@Example.Com>",
		},
		// Name with backslash
		{
			name:     "John\\Doe",
			email:    "john.doe@example.com",
			expected: "\"John\\\\Doe\" <john.doe@example.com>",
		},
		// Empty name and email
		{
			name:     "",
			email:    "",
			expected: "",
		},
		// Name with parentheses
		{
			name:     "John (JD) Doe",
			email:    "john.doe@example.com",
			expected: "\"John (JD) Doe\" <john.doe@example.com>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.email, func(t *testing.T) {
			result := server.FormatRFC822Email(tt.name, tt.email)
			if result != tt.expected {
				t.Errorf("FormatRFC822Email(%q, %q) = %q; want %q", tt.name, tt.email, result, tt.expected)
			}
		})
	}
}
