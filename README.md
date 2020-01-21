
This repository contains the Dockerfiles and configuration scripts for the Official Couchbase Docker images.

If you are a casual user, you probably don't care about this repo, instead you should head over to one of the Couchbase Dockerhub repositories:

* [Couchbase Server](https://hub.docker.com/r/couchbase/server/)
* [Sync Gateway](https://hub.docker.com/r/couchbase/sync-gateway)

# Directory layout

* `community/*` and `enterprise/*` contain *generated* Dockerfiles + assets -- **do not edit**

* `generate/` contains everything needed to generate the Dockerfiles and assets -- **everything you want to edit is here**

# Regenerating from templates

After you change anything under the `generate` directory, you should regenerate from the templates.

**Prerequisites**

* [Install Go](https://golang.org/doc/install)

**Steps**

```
$ cd <project-dir>/generate/generator
$ go generate
```

Expected output:

```
2020/01/20 16:15:23 ../../community/couchbase-server/5.0.1/Dockerfile exists, skipping...
2020/01/20 16:15:23 ../../community/couchbase-server/5.1.1/Dockerfile exists, skipping...
2020/01/20 16:15:23 ../../community/couchbase-server/6.0.0/Dockerfile exists, skipping...
2020/01/20 16:15:23 ../../community/couchbase-server/6.0.2/Dockerfile exists, skipping...
2020/01/20 16:15:23 generateDockerfile called with: {community couchbase-server 6.5.0 false}
2020/01/20 16:15:25 https://packages.couchbase.com/releases/6.5.0/couchbase-server-community_6.5.0-ubuntu16.04_amd64.deb.sha256
2020/01/20 16:15:25 ../../community/sync-gateway/1.0.4/Dockerfile exists, skipping...
2020/01/20 16:15:25 ../../community/sync-gateway/1.1.0/Dockerfile exists, skipping...
2020/01/20 16:15:25 Successfully finished!
```

At this point, you should push your changes to github.

# Adding a new Couchbase Server version + dockerhub tag

**Create directory**

Suppose you want to create a docker image for the newly released Couchbase Server version 9.0.0:

```
$ cd <project-dir>/enterprise/couchbase-server
$ mkdir 9.0.0
```

**Regerate from templates**

See instructions above.

**Push to github**

Commit and push to github

**Kick off dockerhub build**

Login to dockerhub (you need to be on the couchbase team for this step) and create a new build that corresponds to that directory, and enter a matching tag, eg:

* **Branch**: master
* **Dockerfile Location**: /enterprise/couchbase-server/9.0.0
* **Docker Tag Name**: enterprise-9.0.0


# Overriding download url for a "devbuild" or "release candidate" version

If the package binaries are not available on packages.couchbase.com, this is an alternative way of generating the dockerfile.

1. Create the directory you want: eg `/enterprise/sync-gateway/2.0.0-devbuild`

1. Upload the binary package to a publicly available location.  (see existing entries)

1. Update the `init()` function in `generate.go` to add a new version customization to the list, following suit w/ the existing one(s), and pointing to the binary package url from the previous step.

1. Regenerate as usual

1. Verify that the generated dockerfile has the customized package url.