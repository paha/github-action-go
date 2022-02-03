package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/go-github/v42/github"
	githubactions "github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	a := githubactions.New()
	gh_token := a.GetInput("token")
	pr, _ := strconv.Atoi(a.GetInput("pr_number"))
	r := a.GetInput("repo")
	a.Infof("GitHub repository: %s\n", r)
	a.Infof("Pull request Number: %d\n", pr)
	// repo - [user, repository]
	repo := strings.Split(r, string(os.PathSeparator))
	depth, _ := strconv.Atoi(a.GetInput("depth"))

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: gh_token})
	gh := github.NewClient(oauth2.NewClient(ctx, ts))

	files, _, err := gh.PullRequests.ListFiles(ctx, repo[0], repo[1], pr, &github.ListOptions{})
	if err != nil {
		fmt.Println(err)
	}
	a.Debugf("Files changed in this PR: %+v\n", files)

	path := getPath(files, depth, a)
	a.SetOutput("tf_path", path)
	// set label
}

func getPath(f []*github.CommitFile, d int, a *githubactions.Action) string {
	var paths []string

	for _, f := range f {
		dir := filepath.Dir(*f.Filename)
		// Ignore files in the root of repository and files not matching desired depth
		if dir != "." && len(strings.Split(dir, string(os.PathSeparator))) == d {
			paths = append(paths, dir)
		}
	}
	n := len(paths)
	if n < 1 {
		a.Warningf("NO valid paths were found.")
		return "undefined"
	}
	a.Infof("Found %d changed paths.\n", n)
	a.Infof("Paths under witch changes are made in this PR: %+q\n", paths)

	// Matching dir path - https://pkg.go.dev/path/filepath#Match
	// for include input

	return paths[0]
}
