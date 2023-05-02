package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v52/github"
	githubactions "github.com/sethvargo/go-githubactions"
)

const identifier = "tf_path"

func main() {
	a := &ghAction{}
	a.setup()
	a.getIssueLabels()

	files := a.getChangedFiles()
	a.path = identifyPath(cleanDirPath(files, a.inputs, a.action), a.action)
	fmt.Printf("Identified project path: %s\n", a.path)
	// Set the path as the action output
	// Example: ${{ steps.STEP_ID.outputs.tf_path }}
	a.action.SetOutput(identifier, a.path)

	label := fmt.Sprintf("%s: %s", identifier, a.path)
	currentLabel := a.getCurrentPathLabel()
	if currentLabel == nil {
		fmt.Printf("No lable for path is set for the Pull request.\n")
		a.addLabel(label)
	} else if label != *currentLabel.Name {
		fmt.Printf("Pull Request path label needs to be updated.\n")
		a.rmLabel(*currentLabel.Name)
		a.addLabel(label)
	} else {
		fmt.Printf("Pull Request path label is confirmed.\n")
	}
}

func identifyPath(ps []string, a *githubactions.Action) string {
	a.Infof("Valid paths under witch changes are made in this PR: %+q", ps)

	switch {
	case len(ps) == 0:
		a.Warningf("NO valid paths were found.\n")
		// Not erroring, decision can be made in the next steps of the action
		return "undefined"
	case len(ps) == 1:
		return ps[0]
	default:
		a.Warningf("More then one potential project paths found.")
		a.Warningf("Returning the first match.")
		return ps[0]
	}
}

func cleanDirPath(f []*github.CommitFile, i *inputs, a *githubactions.Action) []string {
	var paths []string

	for _, f := range f {
		dir := filepath.Dir(*f.Filename)
		a.Infof("Validating change for %s", dir)
		// Ignore files in the root of repository, directories that start with a '.', and directories not valid acording to the desired depth
		if !strings.HasPrefix(dir, ".") {
			ds := strings.Split(dir, string(os.PathSeparator))
			switch {
			case len(ds) == i.depth:
				paths = append(paths, dir)
				a.Infof("Adding %s to futher validation.", dir)
			case len(ds) < i.depth:
				a.Warningf("Project path depth is longer, %s path isn't valid", dir)
			case len(ds) > i.depth:
				// Cut directory path to desired depth
				p := strings.Join(ds[:i.depth], string(os.PathSeparator))
				paths = append(paths, p)
				a.Infof("Adding %s to futher validation. Actual change path is %s", p, ds)
			}
		}
	}

	return matchRegex(removeDuplicateValues(paths), i.include, i.exclude)
}

func matchRegex(paths []string, include, exclude string) []string {
	if len(include) == 0 && len(exclude) == 0 {
		fmt.Println("No regex defined, skipping matching.")
		return paths
	}

	var p []string
	if len(include) > 0 {
		for _, d := range paths {
			r, _ := regexp.MatchString(include, d)
			if r {
				fmt.Printf("Matched include %s regex for path %s\n", include, d)
				p = append(p, d)
			}
		}
	}

	var px []string
	if len(exclude) > 0 {
		for _, d := range p {
			e, _ := regexp.MatchString(exclude, d)
			if !e {
				fmt.Printf("Not matched exclude %s regex for path %s. \n", include, d)
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
	fmt.Printf("Found and removed %d duplicates\n", len(ps)-len(res))
	return res
}
