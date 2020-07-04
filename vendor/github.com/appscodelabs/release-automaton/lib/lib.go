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

package lib

import (
	"os"
	"sort"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Pair struct {
	Key   string
	Value string
}

func ToOrderedPair(in map[string]string) []Pair {
	out := make([]Pair, 0, len(in))
	for k, v := range in {
		out = append(out, Pair{
			Key:   k,
			Value: v,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

func MergeMaps(dst, src map[string]string) map[string]string {
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
	return dst
}

func Keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func Values(m map[string]string) []string {
	values := sets.NewString()
	for _, v := range m {
		values.Insert(v)
	}
	return values.UnsortedList()
}

// Exists reports whether the named file or directory Exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return false
	}
	return true
}

func UniqComments(comments []string) []string {
	out := make([]string, 0, len(comments))
	s := sets.NewString()
	for i := len(comments) - 1; i >= 0; i-- {
		if !s.Has(comments[i]) {
			out = append([]string{comments[i]}, out...)
			s.Insert(comments[i])
		}
	}
	return out
}
