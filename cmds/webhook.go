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

//nolint:goconst
package cmds

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/appscodelabs/release-automaton/api"
	"github.com/appscodelabs/release-automaton/lib"

	"github.com/alessio/shellescape"
	shell "github.com/codeskyblue/go-sh"
	"github.com/google/go-github/v32/github"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/acme/autocert"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	msgSingleJSON = "Request body must only contain a single JSON object"
)

var (
	secretKey   = ""
	certDir     = "certs"
	email       = "tamal@appscode.com"
	hosts       = []string{"gh-ci-webhook.appscode.ninja"}
	port        = 8989
	enableSSL   bool
	gh          = lib.NewGitHubClient()
	queueLength = 100
	prs         chan PREvent
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Run webhook server",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer()
		},
	}

	cmd.Flags().StringVar(&secretKey, "secret-key", secretKey, "Secret key to verify webhook payloads")
	cmd.Flags().StringVar(&certDir, "cert-dir", certDir, "Directory where certs are stored")
	cmd.Flags().StringVar(&email, "email", email, "Email used by Let's Encrypt to notify about problems with issued certificates")
	cmd.Flags().StringSliceVar(&hosts, "hosts", hosts, "Hosts for which certificate will be issued")
	cmd.Flags().IntVar(&port, "port", port, "Port used when SSL is not enabled")
	cmd.Flags().BoolVar(&enableSSL, "ssl", enableSSL, "Set true to enable SSL via Let's Encrypt")
	cmd.Flags().IntVar(&queueLength, "queue-length", queueLength, "Length of queue used to hold pr events")
	return cmd
}

type PREvent struct {
	PRRepoURL   string // [https://]github.com/{owner}/{name}
	TestRepoURL string // [https://]github.com/{owner}/{name}

	PRTitle  string
	PRNumber int
	PRState  string
	PRMerged bool

	HeadRef string
	HeadSHA string // Also used as branch name in test repo so we can track in the check-runs events
}

func (e PREvent) Branch() string {
	return e.HeadSHA + "@" + e.HeadRef
}

func (e PREvent) EnvFile() []byte {
	var buf bytes.Buffer
	buf.WriteString("PR_REPO_URL=" + shellescape.Quote(e.PRRepoURL))
	buf.WriteRune('\n')
	buf.WriteString("TEST_REPO_URL=" + shellescape.Quote(e.TestRepoURL))
	buf.WriteRune('\n')
	buf.WriteString("PR_TITLE=" + shellescape.Quote(e.PRTitle))
	buf.WriteRune('\n')
	buf.WriteString(fmt.Sprintf("PR_NUMBER=%d", e.PRNumber))
	buf.WriteRune('\n')
	buf.WriteString("PR_STATE=" + shellescape.Quote(e.PRState))
	buf.WriteRune('\n')
	buf.WriteString("PR_HEAD_REF=" + shellescape.Quote(e.HeadRef))
	buf.WriteRune('\n')
	buf.WriteString("PR_HEAD_SHA=" + shellescape.Quote(e.HeadSHA))
	buf.WriteRune('\n')
	return buf.Bytes()
}

func runServer() error {
	sh := shell.NewSession()
	sh.ShowCMD = true
	sh.PipeFail = true
	sh.PipeStdErrors = true

	err := os.RemoveAll(api.Workspace)
	if err != nil {
		panic(err)
	}

	prs = make(chan PREvent, queueLength)

	go processPREvent(gh, sh)

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "Hello, TLS user! Your config: %+v", r.TLS)
	}).Methods(http.MethodGet)
	r.HandleFunc("/check-ci-runs", serveHTTP).Methods(http.MethodPost)
	r.HandleFunc("/check-pr-runs", serveHTTP).Methods(http.MethodPost)
	r.HandleFunc("/pr", serveHTTP).Methods(http.MethodPost)
	r.Use()

	if !enableSSL {
		addr := fmt.Sprintf(":%d", port)
		fmt.Println("Listening to addr", addr)
		return http.ListenAndServe(addr, r)
	}

	// ref:
	// - https://goenning.net/2017/11/08/free-and-automated-ssl-certificates-with-go/
	// - https://stackoverflow.com/a/40494806/244009
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(certDir),
		HostPolicy: autocert.HostWhitelist(hosts...),
		Email:      email,
	}
	server := &http.Server{
		Addr:         ":https",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go func() {
		// does automatic http to https redirects
		err := http.ListenAndServe(":http", certManager.HTTPHandler(nil))
		if err != nil {
			panic(err)
		}
	}()
	return server.ListenAndServeTLS("", "") //Key and cert are coming from Let's Encrypt
}

func processPREvent(gh *github.Client, sh *shell.Session) {
	for pr := range prs {
		fmt.Printf("%#v \n", pr)
		if err := openPR(gh, sh, pr); err != nil {
			fmt.Println(err)
		}
	}
}

func openPR(gh *github.Client, sh *shell.Session, event PREvent) error {
	owner, repo := lib.ParseRepoURL(event.TestRepoURL)

	// pushd, popd
	wdOrig := sh.Getwd()
	defer sh.SetDir(wdOrig)

	// TODO: cache git repo
	wdCur := filepath.Join(api.Workspace, owner)
	err := os.MkdirAll(wdCur, 0755)
	if err != nil {
		return err
	}

	if !lib.Exists(filepath.Join(wdCur, repo)) {
		sh.SetDir(wdCur)

		err = sh.Command("git",
			"clone",
			// "--no-tags", //TODO: ok?
			"--no-recurse-submodules",
			//"--depth=1",
			//"--no-single-branch",
			fmt.Sprintf("https://%s:%s@%s.git", os.Getenv(api.GitHubUserKey), os.Getenv(api.GitHubTokenKey), event.TestRepoURL),
		).Run()
		if err != nil {
			return err
		}
	}
	wdCur = filepath.Join(wdCur, repo)
	sh.SetDir(wdCur)

	err = sh.Command("git", "checkout", "master").Run()
	if err != nil {
		return err
	}
	err = sh.Command("git", "fetch", "origin", "--prune").Run()
	if err != nil {
		return err
	}
	err = sh.Command("git", "reset", "--hard", "origin/master").Run()
	if err != nil {
		return err
	}

	// ignore error in case branch does not exist
	_ = sh.Command("git", "branch", "-D", event.Branch()).Run()

	if event.PRState == "closed" {
		pr, err := lib.ClosePR(gh, owner, repo, event.Branch(), "master")
		if err != nil {
			return err
		}
		err = sh.Command("git", "push", "--delete", "origin", event.Branch()).Run()
		if err != nil {
			return err
		}
		// git push origin --delete feature/login
		fmt.Println("Closed pr:", pr.GetHTMLURL())
		return nil
	}

	if !lib.RemoteBranchExists(sh, event.Branch()) {
		err = sh.Command("git", "checkout", "-b", event.Branch()).Run()
		if err != nil {
			return err
		}
	} else {
		err = sh.Command("git", "checkout", event.Branch()).Run()
		if err != nil {
			return err
		}
	}

	err = ioutil.WriteFile(filepath.Join(wdCur, "event.env"), event.EnvFile(), 0644)
	if err != nil {
		return err
	}
	if lib.RepoModified(sh) {
		err = lib.CommitRepo(sh, "", fmt.Sprintf("%s@%s", event.PRRepoURL, event.HeadSHA))
		if err != nil {
			return err
		}
		err = lib.PushRepo(sh, true)
		if err != nil {
			return err
		}
	}

	// open pr against project repo
	_, err = lib.CreatePR(gh, owner, repo, &github.NewPullRequest{
		Title:               github.String(event.PRTitle),
		Head:                github.String(event.Branch()),
		Base:                github.String("master"),
		Body:                github.String(""),
		MaintainerCanModify: github.Bool(true),
		Draft:               github.Bool(false),
	})
	return err
}

func serveHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(secretKey))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := r.URL.Query()
	switch event := event.(type) {
	case *github.CheckRunEvent:
		if _, ok := query["pr-repo"]; ok {
			handleCIRepoEvent(event, query)
			return
		}
		if _, ok := query["ci-repo"]; ok {
			handlePRRepoEvent(event, query)
			return
		}
		http.Error(w, "unsupported event", http.StatusOK)
		return
	case *github.PullRequestEvent:
		handlePREvent(event, query)
		return
	default:
		http.Error(w, "unsupported event", http.StatusOK)
		return
	}
}

func handlePRRepoEvent(event *github.CheckRunEvent, query url.Values) {
	if event.GetCheckRun().GetApp().GetSlug() == "github-actions" &&
		event.GetCheckRun().GetName() == "Build" &&
		event.GetCheckRun().GetStatus() == "completed" &&
		event.GetCheckRun().GetConclusion() == "success" {

		pr, _, err := gh.PullRequests.Get(
			context.TODO(),
			event.GetRepo().GetOwner().GetLogin(),
			event.GetRepo().GetName(),
			event.GetCheckRun().PullRequests[0].GetNumber(),
		)
		if err != nil {
			panic(err)
		}

		prs <- PREvent{
			PRRepoURL:   strings.TrimPrefix(event.GetRepo().GetHTMLURL(), "https://"),
			TestRepoURL: strings.TrimPrefix(query.Get("ci-repo"), "https://"),
			PRNumber:    event.GetCheckRun().PullRequests[0].GetNumber(),
			PRTitle:     pr.GetTitle(),
			PRState:     pr.GetState(),
			PRMerged:    pr.GetMerged(),
			HeadRef:     event.GetCheckRun().PullRequests[0].GetHead().GetRef(),
			HeadSHA:     event.GetCheckRun().PullRequests[0].GetHead().GetSHA(),
		}
	}
}

func handleCIRepoEvent(event *github.CheckRunEvent, query url.Values) {
	if event.GetCheckRun().GetApp().GetSlug() == "github-actions" {
		owner, repo := lib.ParseRepoURL(query.Get("pr-repo"))
		ref := strings.Split(event.GetCheckRun().PullRequests[0].GetHead().GetRef(), "@")[0] // branch name matches pr repo's sha

		var state string
		if event.GetCheckRun().GetStatus() == "queued" || event.GetCheckRun().GetStatus() == "in_progress" {
			state = "pending"
		} else if event.GetCheckRun().GetStatus() == "completed" {
			switch event.GetCheckRun().GetConclusion() {
			case "success":
				state = "success"
			case "failure":
				state = "failure"
			case "neutral":
				state = "pending"
			case "cancelled", "timed_out", "action_required", "stale":
				state = "error"
			}
		}
		sr, _, err := gh.Repositories.CreateStatus(context.TODO(), owner, repo, ref, &github.RepoStatus{
			// pending, success, error, or failure.
			State:       github.String(state),
			TargetURL:   event.GetCheckRun().HTMLURL,
			Description: github.String(event.GetRepo().GetFullName()),
			Context:     event.GetCheckRun().Name,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(sr)
	}
}

func handlePREvent(event *github.PullRequestEvent, query url.Values) {
	actions := lib.GetQueryParameter(query, "actions")
	if actions.Len() == 0 {
		actions = sets.NewString("opened", "synchronize", "closed", "reopened")
	}
	if actions.Has(event.GetAction()) {
		prs <- PREvent{
			PRRepoURL:   strings.TrimPrefix(event.GetRepo().GetHTMLURL(), "https://"),
			TestRepoURL: strings.TrimPrefix(query.Get("ci-repo"), "https://"),
			PRNumber:    event.GetPullRequest().GetNumber(),
			PRTitle:     event.GetPullRequest().GetTitle(),
			PRState:     event.GetPullRequest().GetState(),
			PRMerged:    event.GetPullRequest().GetMerged(),
			HeadRef:     event.GetPullRequest().GetHead().GetRef(),
			HeadSHA:     event.GetPullRequest().GetHead().GetSHA(),
		}
	}
}
