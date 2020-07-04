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
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/keighl/metabolize"
)

type GoImport struct {
	RepoRoot string
	VCSRoot  string
}

func (g GoImport) String() string {
	if g.VCSRoot == "" {
		return g.RepoRoot
	}
	return g.RepoRoot + " " + g.VCSRoot
}

type MetaData struct {
	GoImport string `meta:"go-import"`
}

func DetectVCSRoot(repoURL string) (string, error) {
	if !strings.Contains(repoURL, "://") {
		repoURL = "https://" + repoURL
	}

	uRepo, err := url.Parse(repoURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse repo url: %v", err)
	}
	qRepo := uRepo.Query()
	qRepo.Set("go-get", "1")
	uRepo.RawQuery = qRepo.Encode()

	res, err := http.Get(uRepo.String())
	if err != nil {
		return "", err
	}
	data := new(MetaData)

	err = metabolize.Metabolize(res.Body, data)
	if err != nil {
		return "", err
	}

	// GoImport: stash.appscode.dev/cli git https://github.com/stashed/cli
	if data.GoImport == "" {
		return "", fmt.Errorf("%s is missing go-import meta tag", uRepo.String())
	}
	fmt.Printf("GoImport: %s\n", data.GoImport)

	parts := strings.Fields(data.GoImport)
	if len(parts) != 3 {
		return "", fmt.Errorf("%s contains badly formatted go-import meta tag %s", uRepo.String(), data.GoImport)
	}

	uVCS, err := url.Parse(parts[2])
	if err != nil {
		return "", fmt.Errorf("failed to parse VCS root %s: %v", parts[2], err)
	}
	//uVCS.Scheme = ""
	vcsURL := path.Join(uVCS.Hostname(), uVCS.Path)
	return strings.TrimSuffix(vcsURL, path.Ext(vcsURL)), nil
}
