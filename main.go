package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v42/github"
	githubactions "github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

type ghAction struct {
	gh     *github.Client
	action *githubactions.Action
	inputs *inputs
	pr     *github.PullRequest
}

type inputs struct {
	pr_number int
	gh_user   string
	gh_repo   string
	depth     int
}

func (s *ghAction) setup() {
	a := githubactions.New()
	event := os.Getenv("GITHUB_EVENT_NAME")
	if event != "pull_request" {
		a.Warningf("This action is designed to wor with 'pull_request' event. Current event: %s", event)
	}

	// collect GitHub action inputs
	// TODO:
	// - Validate inputs
	// - Set inputs defaults
	i := &inputs{}
	i.pr_number, _ = strconv.Atoi(a.GetInput("pr_number"))
	r := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	i.gh_user = r[0]
	i.gh_repo = r[1]
	i.depth, _ = strconv.Atoi(a.GetInput("depth"))

	// GitHub authenticated client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.GetInput("token")})
	s.gh = github.NewClient(oauth2.NewClient(context.Background(), ts))

	a.Infof("GitHub repository: %s\n", r)
	a.Infof("Pull request Number: %d\n", i.pr_number)

	s.action = a
	s.inputs = i
}

func (s *ghAction) getPrLabels() {
	pr, _, err := s.gh.PullRequests.Get(
		context.Background(),
		s.inputs.gh_user,
		s.inputs.gh_repo,
		s.inputs.pr_number,
	)
	if err != nil {
		s.action.Fatalf("Failed getting PR object: %+v\n", err)
	}
	s.pr = pr
	s.action.Debugf("PR lables: %+v\n", pr.Labels)
}

func main() {
	a := &ghAction{}
	a.setup()
	a.getPrLabels()

	files := getChangedFiles(a)
	a.getPrLabels()
	path := identifyPath(files, a)
	// Set the path as the action output
	// ${{ steps.STEP_ID.outputs.tf_path }}
	a.action.SetOutput("tf_path", path)
}

func getChangedFiles(a *ghAction) []*github.CommitFile {
	files, _, err := a.gh.PullRequests.ListFiles(
		context.Background(),
		a.inputs.gh_user,
		a.inputs.gh_repo,
		a.inputs.pr_number,
		&github.ListOptions{},
	)
	if err != nil {
		a.action.Fatalf("Failed getting PR files: %+v\n", err)
	}

	return files
}

func identifyPath(f []*github.CommitFile, a *ghAction) string {
	var paths []string

	for _, f := range f {
		dir := filepath.Dir(*f.Filename)
		a.action.Infof("Validating change for %s\n", dir)
		// Ignore files in the root of repository and files not matching desired depth
		if dir != "." && len(strings.Split(dir, string(os.PathSeparator))) == a.inputs.depth {
			paths = append(paths, dir)
		}
	}

	ps := removeDuplicateValues(paths)
	n := len(ps)
	a.action.Infof("Valid paths under witch changes are made in this PR: %+q\n", ps)
	switch {
	case n == 0:
		a.action.Warningf("NO valid paths were found.\n")
		// Not erroring, decision can be made in the next steps of the action
		return "undefined"
	case n == 1:
		// TODO: Needs to be validated against include/exclude
		return ps[0]
	default:
		// TODO:
		// - validate against include || exclude patherns provided as inputs, https://pkg.go.dev/path/filepath#Match
		// - Check existing label?
		// - How to determine which one to return?
		a.action.Warningf("More then one potential project paths found.\n")
		a.action.Warningf("Returning the first match.\n")
		return ps[0]
	}
}

func removeDuplicateValues(ps []string) []string {
	keys := make(map[string]bool)
	res := []string{}

	for _, p := range ps {
		if _, ok := keys[p]; !ok {
			keys[p] = true
			res = append(res, p)
		}
	}
	return res
}
