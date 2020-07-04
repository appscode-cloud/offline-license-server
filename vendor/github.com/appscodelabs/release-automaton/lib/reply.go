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
	"strings"

	"github.com/appscodelabs/release-automaton/api"
)

func ParseReply(s string) *api.Reply {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil
	}

	rt := api.ReplyType(fields[0])
	params := fields[1:]

	switch rt {
	case api.OkToRelease:
		fallthrough
	case api.Done:
		if len(params) > 0 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt}
	case api.Tagged:
		if len(params) != 1 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt, Tagged: &api.TaggedReplyData{
			Repo: params[0],
		}}
	case api.Go:
		if len(params) != 2 && len(params) != 3 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		data := &api.GoReplyData{
			Repo:       params[0],
			ModulePath: params[1],
		}
		if len(params) == 3 {
			data.VCSRoot = params[2]
		}
		return &api.Reply{Type: rt, Go: data}
	case api.PR:
		if len(params) != 1 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		owner, repo, prNumber := ParsePullRequestURL(params[0])
		return &api.Reply{Type: rt, PR: &api.PullRequestReplyData{
			Repo:   fmt.Sprintf("github.com/%s/%s", owner, repo),
			Number: prNumber,
		}}
	case api.ReadyToTag:
		if len(params) != 2 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt, ReadyToTag: &api.ReadyToTagReplyData{
			Repo:           params[0],
			MergeCommitSHA: params[1],
		}}
	case api.CherryPicked:
		if len(params) != 3 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt, CherryPicked: &api.CherryPickedReplyData{
			Repo:           params[0],
			Branch:         params[1],
			MergeCommitSHA: params[2],
		}}
	case api.Chart:
		if len(params) != 2 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt, Chart: &api.ChartReplyData{
			Repo: params[0],
			Tag:  params[1],
		}}
	case api.ChartPublished:
		if len(params) != 1 {
			panic(fmt.Errorf("unsupported parameters with reply %s", s))
		}
		return &api.Reply{Type: rt, ChartPublished: &api.ChartPublishedReplyData{
			Repo: params[0],
		}}
	default:
		fmt.Printf("unknown reply type found in %s\n", s)
		return nil
	}
}

func ParseComment(s string) []api.Reply {
	var out []api.Reply
	for _, line := range strings.Split(s, "\n") {
		if reply := ParseReply(line); reply != nil {
			out = append(out, *reply)
		}
	}
	return out
}
