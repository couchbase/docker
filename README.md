
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
$ go run generate.go <project-dir>
```

Where `<project-dir>` is the absolute path to the directory where you checked out this repo, for example: `/Users/tleyden/Development/couchbase-docker`

Expected output:

```
2015/04/24 15:41:40 generateDockerfile called with: {community couchbase-server 2.2.0}
2015/04/24 15:41:41 generateDockerfile called with: {community couchbase-server 3.0.1}
2015/04/24 15:41:41 generateDockerfile called with: {community sync-gateway 1.0.4}
2015/04/24 15:41:41 generateDockerfile called with: {enterprise couchbase-server 2.5.2}
2015/04/24 15:41:41 generateDockerfile called with: {enterprise couchbase-server 3.0.2}
2015/04/24 15:41:41 Successfully finished!
```

At this point, you should push your changes to github.

# Adding a new Couchbase Server version + dockerhub tag

**Create directory**

```
$ cd <project-dir>/enterprise/couchbase-server
```

**Regerate from templates**

See instructions above.

**Push to github**

Commit and push to github

**Kick off dockerhub build**

Login to dockerhub (you need to be on the couchbase team for this step) and create a new build that corresponds to that directory, and enter a matching tag, eg:

* **Branch**: master
* **Dockerfile Location**: /enterprise/couchbase-server/3.1.3
* **Docker Tag Name**: enterprise-3.1.3