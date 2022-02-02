# github-action-go

## Docker Hub

Using personal account to store the container on Docker Hub.

```bash
docker login -u paha -p $DH_TOKEN
docker build -t paha/github-action-tf-path .
docker push paha/github-action-tf-path
# if tags are desired
docker tag paha/github-action-tf-path paha/github-action-tf-path:v0.0.1
docker push paha/github-action-tf-path:v0.0.1 

```

References:

---
[1]: https://github.com/sethvargo/go-githubactions
[2]: https://docs.github.com/en/actions/creating-actions/creating-a-docker-container-action
[3]: https://github.com/posener/goaction
[4]: 
