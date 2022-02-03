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
	i := &inputs{}
	var err error
	i.pr_number, err = strconv.Atoi(a.GetInput("pr_number"))
	if err != nil {
		a.Errorf("Failed to get PR number: %v", err)
	}
	r := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	i.gh_user = r[0]
	i.gh_repo = r[1]
	i.depth, err = strconv.Atoi(a.GetInput("depth"))
	if err != nil {
		a.Errorf("Failed to get path depth: %v", err)
	}
	if i.depth > 4 && i.depth < 1 {
		a.Warningf("Path depth has unreasonable value: %d", i.depth)
	}

	// GitHub authenticated client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: a.GetInput("token")})
	s.gh = github.NewClient(oauth2.NewClient(context.Background(), ts))

	a.Infof("GitHub repository: %s", r)
	a.Infof("Pull request Number: %d", i.pr_number)
	// a.Infof("ENV: %+v", os.Environ())

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
		s.action.Fatalf("Failed getting PR object: %+v", err)
	}
	s.pr = pr
	s.action.Infof("PR lables: %+v", pr.Labels)
}

func main() {
	a := &ghAction{}
	a.setup()

	files := getChangedFiles(a)
	a.getPrLabels()
	path := identifyPath(files, a)
	// Set the path as the action output
	// Example: ${{ steps.STEP_ID.outputs.tf_path }}
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
	ps := cleanDirPath(f, a)
	a.action.Infof("Valid paths under witch changes are made in this PR: %+q", ps)
	
	switch {
	case len(ps) == 0:
		a.action.Warningf("NO valid paths were found.\n")
		// Not erroring, decision can be made in the next steps of the action
		return "undefined"
	case len(ps) == 1:
		// TODO: Needs to be validated against include/exclude
		a.action.Infof("Project path: %s", ps[0])
		return ps[0]
	default:
		// TODO:
		// - validate against include || exclude patherns provided as inputs, https://pkg.go.dev/path/filepath#Match
		// - Check existing label?
		// - How to determine which one to return?
		a.action.Warningf("More then one potential project paths found.")
		a.action.Warningf("Returning the first match.")
		a.action.Infof("Project path: %s", ps[0])
		return ps[0]
	}
}

func cleanDirPath(f []*github.CommitFile, a *ghAction) []string {
	var paths []string

	for _, f := range f {
		dir := filepath.Dir(*f.Filename)
		a.action.Infof("Validating change for %s", dir)
		// Ignore files in the root of repository and directories not valid acording to the desired depth
		if dir != "." {
			ds := strings.Split(dir, string(os.PathSeparator))
			switch {
			case len(ds) == a.inputs.depth:
				paths = append(paths, dir)
			case len(ds) < a.inputs.depth:
				a.action.Warningf("Project path depth is longer, %s path isn't valid", dir)
			case len(ds) > a.inputs.depth:
				// Cut directory path to desired depth
				paths = append(paths, strings.Join(ds[:a.inputs.depth], string(os.PathSeparator)))
			}
		}
	}

	return removeDuplicateValues(paths)
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
