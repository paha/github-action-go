package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v42/github"
	githubactions "github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

type ghAction struct {
	gh        *github.Client
	action    *githubactions.Action
	inputs    *inputs
	pr_labels []*github.Label
	path      string
}

type inputs struct {
	pr_number int
	gh_user   string
	gh_repo   string
	depth     int
	include   string
	exclude   string
}

func (s *ghAction) setup() {
	a := githubactions.New()
	event := os.Getenv("GITHUB_EVENT_NAME")
	if event != "pull_request" {
		a.Warningf("This action is designed to work with 'pull_request' event. Current event: %s", event)
	}

	// collect GitHub action inputs
	i := &inputs{}
	var err error
	i.pr_number, _ = strconv.Atoi(strings.Split(os.Getenv("GITHUB_REF_NAME"), "/")[0])
	// Not all GitHub events would contain ref with PR id. Fall back on input if PR number can't be identified.
	if i.pr_number == 0 {
		a.Warningf("Failed to identify PR number from GITHUB_REF_NAME. Trying to check inputs.")
		i.pr_number, _ = strconv.Atoi(a.GetInput("pr_number"))
	}
	r := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	i.gh_user = r[0]
	i.gh_repo = r[1]
	// depth is optional, if not provided set to 1
	d := a.GetInput("depth")
	if len(d) == 0 {
		d = "1"
	}
	i.depth, err = strconv.Atoi(d)
	if err != nil {
		a.Errorf("Failed to get path depth: %v", err)
	}
	if i.depth > 4 && i.depth < 1 {
		a.Warningf("Path depth has an unreasonable value: %d", i.depth)
	}
	i.include = a.GetInput("include")
	i.exclude = a.GetInput("exclude")

	// GitHub authenticated client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.GetInput("token")})
	s.gh = github.NewClient(oauth2.NewClient(context.Background(), ts))

	a.Infof("GitHub repository: %s", r)
	a.Infof("Pull request Number: %d", i.pr_number)
	// DEBUG
	// a.Infof("ENV: %+v", os.Environ())

	s.action = a
	s.inputs = i
}

func (s *ghAction) getChangedFiles() []*github.CommitFile {
	files, _, err := s.gh.PullRequests.ListFiles(
		context.Background(),
		s.inputs.gh_user,
		s.inputs.gh_repo,
		s.inputs.pr_number,
		&github.ListOptions{},
	)
	if err != nil {
		s.action.Fatalf("Failed getting PR files: %+v\n", err)
	}

	return files
}

func (s *ghAction) getIssueLabels() {
	issue, _, err := s.gh.Issues.Get(
		context.Background(),
		s.inputs.gh_user,
		s.inputs.gh_repo,
		s.inputs.pr_number,
	)
	if err != nil {
		s.action.Fatalf("Failed getting PR object: %+v", err)
	}
	s.pr_labels = issue.Labels
	fmt.Printf("Pull Requiest labels: %+v\n", issue.Labels)
}

func (s *ghAction) getCurrentPathLabel() *github.Label {
	for _, l := range s.pr_labels {
		if strings.HasPrefix(*l.Name, identifier) {
			return l
		}
	}

	return nil
}

func (s *ghAction) addLabel(n string) {
	_, _, err := s.gh.Issues.AddLabelsToIssue(
		context.Background(),
		s.inputs.gh_user,
		s.inputs.gh_repo,
		s.inputs.pr_number,
		[]string{n},
	)
	if err != nil {
		s.action.Warningf("Failed adding label: %+v", err)
	}
	fmt.Printf("Added PR label %s\n", n)
}

func (s *ghAction) rmLabel(n string) {
	_, err := s.gh.Issues.RemoveLabelForIssue(
		context.Background(),
		s.inputs.gh_user,
		s.inputs.gh_repo,
		s.inputs.pr_number,
		n,
	)
	if err != nil {
		s.action.Warningf("Failed removing label: %+v", err)
	}
	fmt.Printf("Deleted PR label %s\n", n)
}
