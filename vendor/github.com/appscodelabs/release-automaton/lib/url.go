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
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

func ParsePullRequestURL(prURL string) (string, string, int) {
	if !strings.Contains(prURL, "://") {
		prURL = "https://" + prURL
	}

	u, err := url.Parse(prURL)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(u.Path, "/")
	if u.Hostname() != "github.com" || len(parts) != 5 || parts[3] != "pull" {
		panic(fmt.Errorf("invalid or unsupported release tracker url: %s", prURL))
	}

	owner := parts[1]
	repo := parts[2]
	prNumber, err := strconv.Atoi(parts[4])
	if err != nil {
		panic(err)
	}
	return owner, repo, prNumber
}

func ParseRepoURL(repoURL string) (string, string) {
	if !strings.Contains(repoURL, "://") {
		repoURL = "https://" + repoURL
	}

	u, err := url.Parse(repoURL)
	if err != nil {
		panic(err)
	}
	parts := strings.Split(u.Path, "/")
	if u.Hostname() != "github.com" || len(parts) != 3 {
		panic(fmt.Errorf("invalid or unsupported repo url: %s", repoURL))
	}

	owner := parts[1]
	repo := parts[2]
	return owner, repo
}

func GetQueryParameter(v url.Values, key string) sets.String {
	out := sets.NewString()

	if v == nil {
		return out
	}
	vs := v[key]
	if len(vs) == 0 {
		return out
	}

	for _, s := range vs {
		entries := strings.Split(s, ",")
		for _, e := range entries {
			e = strings.TrimSpace(e)
			if e != "" {
				out.Insert(e)
			}
		}
	}
	return out
}
