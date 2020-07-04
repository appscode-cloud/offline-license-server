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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/appscodelabs/release-automaton/api"
	"github.com/appscodelabs/release-automaton/templates"

	"github.com/Masterminds/sprig"
)

func UpdateChangelog(dir string, release api.Release, repoURL, tag string, commits []api.Commit) {
	var status api.ChangelogStatus
	for _, projects := range release.Projects {
		for u, project := range projects {
			if u == repoURL {
				status = project.Changelog
				if status == api.SkipChangelog {
					return
				}
				break
			}
		}
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	filenameChlog := filepath.Join(dir, "CHANGELOG.json")
	chlog := LoadChangelog(dir, release)

	var repoFound bool
	for repoIdx := range chlog.Projects {
		if chlog.Projects[repoIdx].URL == repoURL {
			repoFound = true

			var tagFound bool
			for tagIdx := range chlog.Projects[repoIdx].Releases {
				if chlog.Projects[repoIdx].Releases[tagIdx].Tag == tag {
					chlog.Projects[repoIdx].Releases[tagIdx].Commits = commits
					tagFound = true
					break
				}
			}
			if !tagFound {
				chlog.Projects[repoIdx].Releases = append(chlog.Projects[repoIdx].Releases, api.ReleaseChangelog{
					Tag:     tag,
					Commits: commits,
				})
			}
			break
		}
	}
	if !repoFound {
		chlog.Projects = append(chlog.Projects, api.ProjectChangelog{
			URL: repoURL,
			Releases: []api.ReleaseChangelog{
				{
					Tag:     tag,
					Commits: commits,
				},
			},
		})
	}
	chlog.Sort()

	data, err := MarshalJson(chlog)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filenameChlog, data, 0644)
	if err != nil {
		panic(err)
	}

	WriteChangelogMarkdown(filepath.Join(dir, "README.md"), "changelog.tpl", chlog)
}

func LoadChangelog(dir string, release api.Release) api.Changelog {
	var chlog api.Changelog

	filename := filepath.Join(dir, "CHANGELOG.json")
	data, err := ioutil.ReadFile(filename)
	if err == nil {
		err = json.Unmarshal(data, &chlog)
		if err != nil {
			panic(err)
		}
	}
	chlog.ProductLine = release.ProductLine
	chlog.Release = release.Release
	chlog.ReleaseProjectURL = fmt.Sprintf("https://github.com/%s", os.Getenv("GITHUB_REPOSITORY"))
	chlog.DocsURL = fmt.Sprintf(release.DocsURLTemplate, release.Release)
	chlog.ReleaseDate = time.Now().UTC()
	chlog.KubernetesVersion = release.KubernetesVersion

	return chlog
}

func WriteChangelogMarkdown(filename string, tplname string, data interface{}) {
	var err error

	err = os.MkdirAll(filepath.Dir(filename), 0755)
	if err != nil {
		panic(err)
	}

	tpl := template.New("").Funcs(sprig.TxtFuncMap())
	for _, name := range templates.AssetNames() {
		tpl, err = tpl.New(name).Parse(string(templates.MustAsset(name)))
		if err != nil {
			panic(err)
		}
	}

	var buf bytes.Buffer
	err = tpl.ExecuteTemplate(&buf, tplname, data)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}
