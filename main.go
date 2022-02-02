package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v42/github"
	githubactions "github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

func main() {
	val := githubactions.GetInput("val")
	/*
	   github
	         "repository_owner": "shiftcars",
	         "repository": "shiftcars/tf-fastly",
	         "event_name": "pull_request",
	         event
	               "number": 22
	*/
	if val == "" {
		// githubactions.Fatalf("missing 'val'")
		fmt.Println("missing 'val'")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	gh := github.NewClient(tc)
	files, _, _ := gh.PullRequests.ListFiles(
		ctx,
		"shiftcars",
		"tf-fastly",
		22,
		&github.ListOptions{},
	)

	for _, f := range files {
		fmt.Println(f.GetFilename())
	}
}
