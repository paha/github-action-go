# github-action-go

GitHub Action for Terrafrom project path identification on Terraform monorepos. The GitHub action fetches changed files from the GitHub PR API, validates paths in order to determin what directory Terrafrom execution needs to happen.

- Using a container for action delivery avoids spending time of fetching repo, deps and compiling. Recomended use is container delivery via: `docker://paha/github-action-tf-path`
- PR id is taken from `GITHUB_REF_NAME` 

Required action inputs:

| Input name | Description | Suggested value 
| --- | --- | --- 
| token | GitHub token |  `${{ github.token }}`

Optional action inputs:
| Input name | Description | Example value | default
| --- | --- | --- | ---
| include | Only output paths that match the regex | `aws-` | None
| exclude | Exclude paths that match the regex | `notes-` | `^.`
| depth | Terrafrom project path depth | `1` | `1`
| pr_number | GitHub PR id | `${{ github.event.number }}` | `$GITHUB_REF_NAME`

___Note___:
- _Only base paths are validated._
- _Validation is using https://pkg.go.dev/regexp#MatchString._
- _Anything that starts with a `.` is not validated._

## Example


```yaml
on:
  pull_request:

jobs:
  terraform:
    name: Debug
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v2

      - name: TF path action
        id: path
        uses: docker://paha/github-action-tf-path:v0.0.1
        with:
          token: ${{ github.token }}

      - name: echo output
        run: echo ${{ steps.path.outputs.tf_path }}
```

## References:

- [GitHub Actions SDK][1]
- The action container on [Docker Hub][2]

---
[1]: https://github.com/sethvargo/go-githubactions
[2]: https://hub.docker.com/r/paha/github-action-tf-path
