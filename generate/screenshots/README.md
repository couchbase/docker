# Screenshot Robots

This directory contains tooling used to automatically generate the screenshots embedded in the Docker Hub overview pages. It needs enough free memory for docker to bring up a node, and will probably need tweaked over time as UI elements evolve.

There are two robots:

| Robot | Product | Compose file | Output |
|-------|---------|--------------|--------|
| `robot` | `couchbase/server` | `docker-compose.yml` | `./output` |
| `robot-oi` | `couchbase/operational-insights` | `docker-compose.oi.yml` | `./output/oi` |

## How it works

A docker-compose stack creates the product service, and a robot service (playwright/node/chromium), the robot service opens the product UI, walks through the initial setup wizard resizing the viewport and taking screenshots as it goes. Screenshots are stored in the output directory listed above when complete.

Output filenames must stay in sync with the image URLs in the corresponding `generate/resources/<product>/README.md`, since that README is what gets published as the Docker Hub description.

Note: we use `--renew-anon-volumes` to ensure the product container is coming up on clean volumes and not bringing up an initialised cluster on subsequent runs. With `--exit-code-from robot` we ensure the robot is responsible for the lifecycle of the stack.

## Example Usage: Couchbase Server

`docker-compose build && TAG=6.6.2 docker-compose up --renew-anon-volumes --exit-code-from robot`

## Example Usage: Operational Insights

The Operational Insights stack also brings up [Versity S3 Gateway](https://github.com/versity/versitygw) to stand in for S3-compatible blob storage, which creates its own bucket on startup, and the robot configures the cluster against it, mirroring the quickstart in the Docker Hub overview.

```
TAG=3.0.0 docker compose -f docker-compose.oi.yml build
TAG=3.0.0 docker compose -f docker-compose.oi.yml up --renew-anon-volumes --abort-on-container-exit --exit-code-from robot
```

To capture a build that is not yet on Docker Hub, set `IMAGE` to its repository, e.g. `IMAGE=<registry>/operational-insights TAG=3.0.0-<build>`.

The run takes several minutes: after initializing the cluster the robot waits for the Analytics service to warm up, then loads `travel-sample` and waits for ingestion to finish before running the sample query.

Notes:

* `robot-oi` builds on `mcr.microsoft.com/playwright`, which is multi-arch, so it runs natively on both amd64 and arm64. The older `robot` image is amd64-only and runs under emulation on Apple Silicon.
* On failure the robot writes `FAILURE.png` to its output directory and dumps the page text, so you can see which step broke when the UI changes.
* If a step times out on an element that does exist, check whether the input is visually hidden behind a styled label. Several of them are, which is why the robots click those via the DOM and wait for `state: 'attached'` rather than for visibility.

## Uploading the output

Once the images have been generated and you have visually confirmed they are correct, upload them to the origin bucket behind `https://d774lla4im6mk.cloudfront.net`.

Couchbase Server images live at the root of the bucket:

`aws s3 cp --recursive output/ s3://cb-dockerhub-screenshots-origin --exclude 'oi/*'`

Operational Insights images live in a folder per release, `oi/<version>/`, which the overview's image URLs name:

`aws s3 cp --recursive output/oi/ s3://cb-dockerhub-screenshots-origin/oi/3.0.0/ --exclude FAILURE.png`

Because each release gets new object keys, no CloudFront invalidation is needed, and the overview pages of earlier releases keep their own images. When a later release needs new screenshots, upload them to a new folder and update the image URLs in the README to match.

If you ever do overwrite existing objects, note that they carry no `Cache-Control` header, so CloudFront keeps serving the previous images until its default TTL expires. Invalidate the changed paths after uploading:

`aws cloudfront create-invalidation --distribution-id <id> --paths '/oi/3.0.0/*'`

Look up `<id>` with:

`aws cloudfront list-distributions --query "DistributionList.Items[?contains(DomainName, 'd774lla4im6mk')].Id" --output text`
