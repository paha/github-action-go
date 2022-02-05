# github-action-go

GitHub Action to identify a path of changed files on monorepos, with regex and depth validation.

Example use-case is execution path for Terraform monorepo. The GitHub action fetches changed files from the GitHub PR API, validates paths to determine what directory Terraform execution needs to happen.

- Using a container for action delivery avoids spending time fetching repo, dependencies, and compiling. Recommended use is container delivery via: `docker://paha/github-action-tf-path`
- PR id is taken from `GITHUB_REF_NAME`, if fails, `pr_number` input is taken.
- The path is persisted as a Pull Request label `tf_path: PATH` for visibility.


Required action inputs:

| Input name | Description | Suggested value 
| --- | --- | --- 
| token | GitHub token |  `${{ github.token }}`

Optional action inputs:
| Input name | Description | Example value | default
| --- | --- | --- | ---
| include | Only output paths that match the regex | `myproject` | None
| exclude | Exclude paths that match the regex | `mynotes` | `^.`
| depth | Terrafrom project path depth | `1` | `1`
| pr_number | GitHub PR id | `${{ github.event.number }}` | `$GITHUB_REF_NAME`

___Note___:
- _Only base paths are validated._
- _Validation for include/exclude regex is using https://pkg.go.dev/regexp#MatchString._
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
        uses: docker://paha/github-action-tf-path:v0.0.2
        with:
          token: ${{ github.token }}

      - name: my path
        run: echo ${{ steps.path.outputs.tf_path }}
```

## References:

- [GitHub Actions SDK][1]
- The action container on [Docker Hub][2]

---
[1]: https://github.com/sethvargo/go-githubactions
[2]: https://hub.docker.com/r/paha/github-action-tf-path
