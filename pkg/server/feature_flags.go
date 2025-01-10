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

	"gomodules.xyz/errors"
	"gomodules.xyz/sets"
)

var knownFlags = sets.NewString("DisableAnalytics", "Constraints")

type FeatureFlags map[string]string

func (f FeatureFlags) IsValid() error {
	var errs []error
	for k := range f {
		if !knownFlags.Has(k) {
			errs = append(errs, fmt.Errorf("unknown feature flag %q", k))
		}
	}
	return errors.NewAggregate(errs)
}

func (f FeatureFlags) ToSlice() []string {
	if len(f) == 0 {
		return nil
	}
	result := make([]string, 0, len(f))
	for k, v := range f {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}
