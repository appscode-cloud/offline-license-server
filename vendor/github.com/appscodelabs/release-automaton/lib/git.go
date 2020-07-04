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
	"fmt"
	"strings"

	"github.com/appscodelabs/release-automaton/api"

	shell "github.com/codeskyblue/go-sh"
)

func ListTags(sh *shell.Session) ([]string, error) {
	data, err := sh.Command("git", "tag").Output()
	if err != nil {
		return nil, err
	}
	return strings.Fields(string(data)), nil
}

func RemoteBranchExists(sh *shell.Session, branch string) bool {
	data, err := sh.Command("git", "ls-remote", "--heads", "origin", branch).Output()
	if err != nil {
		panic(err)
	}
	return len(bytes.TrimSpace(data)) > 0
}

func RemoteTagExists(sh *shell.Session, tag string) bool {
	// git ls-remote --exit-code --tags origin <tag>
	err := sh.Command("git", "ls-remote", "--exit-code", "--tags", "origin", tag).Run()
	return err == nil
}

func GetRemoteTag(sh *shell.Session, tag string) string {
	// git ls-remote --exit-code --tags origin <tag>
	data, err := sh.Command("git", "ls-remote", "--exit-code", "--tags", "origin", tag).Output()
	if err != nil {
		return ""
	}
	return strings.Fields(string(data))[0]
}

type ConditionFunc func(*shell.Session, string) bool

func MeetsCondition(fn ConditionFunc, sh *shell.Session, items ...string) bool {
	for _, item := range items {
		if !fn(sh, item) {
			return false
		}
	}
	return true
}

func FirstCommit(sh *shell.Session) string {
	// git rev-list --max-parents=0 HEAD
	// ref: https://stackoverflow.com/a/5189296
	data, err := sh.Command("git", "rev-list", "--max-parents=0", "HEAD").Output()
	if err != nil {
		panic(err)
	}
	commits := strings.Fields(string(data))
	return commits[len(commits)-1]
}

func LastCommitSHA(sh *shell.Session) string {
	// git show -s --format=%H
	data, err := sh.Command("git", "show", "-s", "--format=%H").Output()
	if err != nil {
		panic(err)
	}
	commits := strings.Fields(string(data))
	return commits[0]
}

func LastCommitSubject(sh *shell.Session) string {
	// git show -s --format=%s
	data, err := sh.Command("git", "show", "-s", "--format=%s").Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(data))
}

func LastCommitBody(sh *shell.Session, trimBlankLines bool) string {
	// git show -s --format=%b
	data, err := sh.Command("git", "show", "-s", "--format=%b").Output()
	if err != nil {
		panic(err)
	}

	if !trimBlankLines {
		return strings.TrimSpace(string(data))
	}

	var out []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func AnyRepoModified(wd string, sh *shell.Session) bool {
	wdorig := sh.Getwd()
	defer sh.SetDir(wdorig)

	sh.SetDir(wd)
	return RepoModified(sh)
}

func RepoModified(sh *shell.Session) bool {
	err := sh.Command("git", "add", "--all").Run()
	if err != nil {
		panic(err)
	}
	// https://stackoverflow.com/questions/10385551/get-exit-code-go
	err = sh.Command("git", "diff", "--exit-code", "-s", "HEAD").Run()
	return err != nil
}

func CommitAnyRepo(wd string, sh *shell.Session, tag string, messages ...string) error {
	wdorig := sh.Getwd()
	defer sh.SetDir(wdorig)

	sh.SetDir(wd)
	return CommitRepo(sh, tag, messages...)
}

func CommitRepo(sh *shell.Session, tag string, messages ...string) error {
	err := sh.Command("git", "add", "--all").Run()
	if err != nil {
		return err
	}
	//  git commit -a -s -m "Prepare for release %tag"
	args := []interface{}{
		"commit", "-a", "-s",
	}
	if tag != "" {
		args = append(args, "-m", "Prepare for release "+tag)
	}
	for _, msg := range messages {
		args = append(args, "-m", msg)
	}
	return sh.Command("git", args...).Run()
}

func PushAnyRepo(wd string, sh *shell.Session, pushTag bool) error {
	wdorig := sh.Getwd()
	defer sh.SetDir(wdorig)

	sh.SetDir(wd)
	return PushRepo(sh, pushTag)
}

func PushRepo(sh *shell.Session, pushTag bool) error {
	args := []interface{}{"push", "-u", "origin", "HEAD"}
	if pushTag {
		args = append(args, "--tags")
	}
	return sh.Command("git", args...).Run()
}

func TagRepo(sh *shell.Session, tag string, messages ...string) error {
	args := []interface{}{
		"tag", "-a", tag, "-m", tag,
	}
	for _, msg := range messages {
		args = append(args, "-m", msg)
	}
	return sh.Command("git", args...).Run()
}

func ListCommits(sh *shell.Session, start, end string) []api.Commit {
	// git log --oneline --ancestry-path start..end | cat
	// ref: https://stackoverflow.com/a/44344164/244009
	data, err := sh.Command("git", "log", "--oneline", "--ancestry-path", fmt.Sprintf("%s..%s", start, end)).Output()
	if err != nil {
		panic(err)
	}
	var commits []api.Commit
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		idx := strings.IndexRune(line, ' ')
		if idx != -1 {
			commits = append(commits, api.Commit{
				SHA:     line[:idx],
				Subject: line[idx+1:],
			})
		}
	}
	return commits
}

func ResetRepo(sh *shell.Session) error {
	// git add --all; git stash; git stash drop
	err := sh.Command("git", "add", "--all").Run()
	if err != nil {
		return err
	}
	_ = sh.Command("git", "stash").Run()
	return nil
}
