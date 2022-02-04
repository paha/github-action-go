package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
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
	// TODO:
	// Potentially GITHUB_REF=refs/pull/2/merge or GITHUB_REF_NAME=2/merge can be used
	i.pr_number, err = strconv.Atoi(a.GetInput("pr_number"))
	if err != nil {
		a.Errorf("Failed to get PR number: %v", err)
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
	// ongoing DEBUG
	a.Infof("ENV: %+v", os.Environ())

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
	a.getPrLabels()

	files := getChangedFiles(a)
	path := identifyPath(files, a)
	a.action.Infof("Identified project path: %s", path)
	// Set the path as the action output
	// Example: ${{ steps.STEP_ID.outputs.tf_path }}
	a.action.SetOutput("tf_path", path)

	// TODO: manage PR label for the project path.
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
		return ps[0]
	default:
		a.action.Warningf("More then one potential project paths found.")
		a.action.Warningf("Returning the first match.")
		return ps[0]
	}
}

func cleanDirPath(f []*github.CommitFile, a *ghAction) []string {
	var paths []string

	for _, f := range f {
		dir := filepath.Dir(*f.Filename)
		a.action.Infof("Validating change for %s", dir)
		// Ignore files in the root of repository, directories that start with a '.', and directories not valid acording to the desired depth
		if !strings.HasPrefix(dir, ".") {
			ds := strings.Split(dir, string(os.PathSeparator))
			switch {
			case len(ds) == a.inputs.depth:
				paths = append(paths, dir)
				a.action.Infof("Adding %s to futher validation.", dir)
			case len(ds) < a.inputs.depth:
				a.action.Warningf("Project path depth is longer, %s path isn't valid", dir)
			case len(ds) > a.inputs.depth:
				// Cut directory path to desired depth
				p := strings.Join(ds[:a.inputs.depth], string(os.PathSeparator))
				paths = append(paths, p)
				a.action.Infof("Adding %s to futher validation. Actual change path is %s", p, ds)
			}
		}
	}

	return matchRegex(removeDuplicateValues(paths), a.inputs.include, a.inputs.exclude)
}

func matchRegex(paths []string, include, exclude string) []string {
	if len(include) == 0 && len(exclude) == 0 {
		return paths
	}

	var p []string
	if len(include) > 0 {
		for _, d := range paths {
			r, _ := regexp.MatchString(include, d)
			if r {
				p = append(p, d)
			}
		}
	}

	var px []string
	if len(exclude) > 0 {
		for _, d := range p {
			e, _ := regexp.MatchString(exclude, d)
			if !e {
				px = append(px, d)
			}
		}
	} else {
		return p
	}

	return px
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
