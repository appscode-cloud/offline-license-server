/*
Copyright The Kubepack Authors.

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

package cmds

import (
	"fmt"
	"os"
	"strings"

	"gomodules.xyz/cert"
)

func Fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func merge(cn string, sans []string) []string {
	var found bool
	for _, name := range sans {
		if strings.EqualFold(name, cn) {
			found = true
			break
		}
	}
	if found {
		return sans
	}
	return append([]string{cn}, sans...)
}

func Filename(cfg cert.Config) string {
	if len(cfg.Organization) == 0 {
		return cfg.AltNames.DNSNames[0]
	}
	return cfg.AltNames.DNSNames[0] + "@" + cfg.Organization[0]
}
