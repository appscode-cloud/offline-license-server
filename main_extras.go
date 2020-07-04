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

//nolint
package main

import (
	"context"
	"fmt"

	"github.com/appscodelabs/release-automaton/lib"

	"github.com/google/go-github/v32/github"
)

func main_CreateStatus() {
	gh := lib.NewGitHubClient()
	owner := "appscode-cloud"
	repo := "hugo-actions-demo"
	ref := "8f9913d940c4c75be8d0050c6fa79907a2c0c1e4"
	sr, _, err := gh.Repositories.CreateStatus(context.TODO(), owner, repo, ref, &github.RepoStatus{
		State:       github.String("pending"),
		TargetURL:   github.String("https://appscode.com"),
		Description: github.String("P_E_N_D"),
		Context:     github.String("GH_CI"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(sr)
}
