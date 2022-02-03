# github-action-go

Terrafrom project path identification for Terraform monorepos. The GitHub action fetches changed files from the GitHub PR API, validates paths in order to determin what directory Terrafrom execution needs to happen.

Required action inputs:

| Input name | Description | Suggested value 
| --- | --- | --- 
| depth | Terrafrom project path depth | `1`
| token | GitHub token |  `${{ github.token }}`
| pr_number | GitHub PR id | `${{ github.event.number }}`

Optional action inputs:
| Input name | Description | Suggested value 
| --- | --- | --- 
| include | Only output paths that match the regex | `aws-*`
| exclude | Exclude paths that match the regex | `notes-*`

*Note: Only Base paths are validated.*

NOTES: 
- Action delivery options. It can be used as a normal action but using a container avoids on spending time of fetching deps and compiling. Recomended use is container delivery via: `docker://paha/github-action-tf-path`
- The event must be `pull_request` to fetch the PR number.

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

      - name: Debug action
        id: path
        uses: docker://paha/github-action-tf-path:latest
        with:
          depth: 1
          token: ${{ github.token }}
          pr_number: ${{ github.event.number }}

      - name: echo output
        run: echo ${{ steps.path.outputs.tf_path }}
```

## Docker Hub

Using the personal account to store the container on Docker Hub.

```bash
docker login -u paha -p $DH_TOKEN
docker build -t paha/github-action-tf-path .
docker push paha/github-action-tf-path
# tags
docker tag paha/github-action-tf-path paha/github-action-tf-path:v0.0.1
docker push paha/github-action-tf-path:v0.0.1 

```

References:

---
[1]: https://github.com/sethvargo/go-githubactions
[2]: https://docs.github.com/en/actions/creating-actions/creating-a-docker-container-action
[3]: https://github.com/posener/goaction
