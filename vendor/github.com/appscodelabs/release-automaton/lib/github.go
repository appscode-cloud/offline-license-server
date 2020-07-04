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
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/appscodelabs/release-automaton/api"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/util/sets"
)

func NewGitHubClient() *github.Client {
	token, found := os.LookupEnv(api.GitHubTokenKey)
	if !found {
		panic(api.GitHubTokenKey + " env var is not set")
	}

	// Create the http client.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.TODO(), ts)

	return github.NewClient(tc)
}

func ListLabelsByIssue(ctx context.Context, gh *github.Client, owner, repo string, number int) (sets.String, error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	result := sets.NewString()
	for {
		labels, resp, err := gh.Issues.ListLabelsByIssue(ctx, owner, repo, number, opt)
		if err != nil {
			break
		}
		for _, entry := range labels {
			result.Insert(entry.GetName())
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return result, nil
}

func ListReviews(ctx context.Context, gh *github.Client, owner, repo string, number int) ([]*github.PullRequestReview, error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	var result []*github.PullRequestReview
	for {
		reviews, resp, err := gh.PullRequests.ListReviews(ctx, owner, repo, number, opt)
		if err != nil {
			break
		}
		result = append(result, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return result, nil
}

func ListPullRequestComment(ctx context.Context, gh *github.Client, owner, repo string, number int) ([]*github.PullRequestComment, error) {
	opt := &github.PullRequestListCommentsOptions{
		Sort:      "created",
		Direction: "asc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var result []*github.PullRequestComment
	for {
		comments, resp, err := gh.PullRequests.ListComments(ctx, owner, repo, number, opt)
		if err != nil {
			break
		}
		result = append(result, comments...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return result, nil
}

func ListComments(ctx context.Context, gh *github.Client, owner, repo string, number int) ([]*github.IssueComment, error) {
	opt := &github.IssueListCommentsOptions{
		Sort:      github.String("created"),
		Direction: github.String("asc"),
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var result []*github.IssueComment
	for {
		comments, resp, err := gh.Issues.ListComments(ctx, owner, repo, number, opt)
		if err != nil {
			break
		}
		result = append(result, comments...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return result, nil
}

// https://developer.github.com/v3/pulls/reviews/#create-a-review-for-a-pull-request
func PRApproved(gh *github.Client, owner string, repo string, prNumber int) bool {
	reviews, err := ListReviews(context.TODO(), gh, owner, repo, prNumber)
	if err != nil {
		panic(err)
	}
	for _, review := range reviews {
		if review.GetState() == "REQUEST_CHANGES" {
			return false
		}
	}
	for _, review := range reviews {
		if review.GetState() == "APPROVED" {
			return true
		}
	}
	return false
}

func CreatePR(gh *github.Client, owner string, repo string, req *github.NewPullRequest, labels ...string) (*github.PullRequest, error) {
	labelSet := sets.NewString(labels...)
	var result *github.PullRequest

	head := req.GetHead()
	if !strings.ContainsRune(head, ':') {
		head = owner + ":" + req.GetHead()
	}
	prs, _, err := gh.PullRequests.List(context.TODO(), owner, repo, &github.PullRequestListOptions{
		State: "open",
		Head:  head,
		Base:  req.GetBase(),
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(prs) == 0 {
		result, _, err = gh.PullRequests.Create(context.TODO(), owner, repo, req)
		// "A pull request already Exists" error should NEVER happen since we already checked for existence
		if err != nil {
			return nil, err
		}
		//if e2, ok := err.(*github.ErrorResponse); ok {
		//	var matched bool
		//	for _, entry := range e2.Errors {
		//		if strings.HasPrefix(entry.Message, "A pull request already Exists") {
		//			matched = true
		//			break
		//		}
		//	}
		//	if !matched {
		//		return nil, err
		//	}
		//	// else ignore error because pr already Exists
		//	// else should NEVER happen since we already checked for existence
		//} else if err != nil {
		//	return nil, err
		//}
	} else {
		result = prs[0]
		for _, label := range result.Labels {
			labelSet.Delete(label.GetName())
		}
	}

	if labelSet.Len() > 0 {
		_, _, err := gh.Issues.AddLabelsToIssue(context.TODO(), owner, repo, result.GetNumber(), labelSet.UnsortedList())
		if err != nil {
			return nil, err
		}
	}

	return result, err
}

func ClosePR(gh *github.Client, owner string, repo string, head, base string) (*github.PullRequest, error) {
	if !strings.ContainsRune(head, ':') {
		head = owner + ":" + head
	}
	prs, _, err := gh.PullRequests.List(context.TODO(), owner, repo, &github.PullRequestListOptions{
		State: "open",
		Head:  head,
		Base:  base,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(prs) == 1 {
		pr, _, err := gh.PullRequests.Edit(context.TODO(), owner, repo, prs[0].GetNumber(), &github.PullRequest{
			State: github.String("closed"),
		})
		return pr, err
	}

	return nil, fmt.Errorf("pr not found")
}

func LabelPR(gh *github.Client, owner string, repo, head, base string, labels ...string) error {
	labelSet := sets.NewString(labels...)
	var result *github.PullRequest

	if !strings.ContainsRune(head, ':') {
		head = owner + ":" + head
	}
	prs, _, err := gh.PullRequests.List(context.TODO(), owner, repo, &github.PullRequestListOptions{
		State: "open",
		Head:  head,
		Base:  base,
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return err
	}
	if len(prs) == 0 {
		return fmt.Errorf("no open pr found")
	}

	result = prs[0]
	for _, label := range result.Labels {
		labelSet.Delete(label.GetName())
	}
	if labelSet.Len() > 0 {
		_, _, err := gh.Issues.AddLabelsToIssue(context.TODO(), owner, repo, result.GetNumber(), labelSet.UnsortedList())
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveLabel(gh *github.Client, owner string, repo string, number int, label string) error {
	_, err := gh.Issues.RemoveLabelForIssue(context.TODO(), owner, repo, number, label)
	if ge, ok := err.(*github.ErrorResponse); ok {
		if ge.Response.StatusCode == http.StatusNotFound {
			return nil
		}
	}
	return err
}

func ListTags2(ctx context.Context, gh *github.Client, owner, repo string) ([]*github.RepositoryTag, error) {
	opt := &github.ListOptions{
		PerPage: 100,
	}

	var result []*github.RepositoryTag
	for {
		reviews, resp, err := gh.Repositories.ListTags(ctx, owner, repo, opt)
		if err != nil {
			break
		}
		result = append(result, reviews...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return result, nil
}
