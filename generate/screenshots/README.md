# Screenshot Robots

This directory contains tooling used to automatically generate the screenshots embedded in the Docker Hub overview pages. It needs enough free memory for docker to bring up a node, and will probably need tweaked over time as UI elements evolve.

There are two robots:

| Robot | Product | Compose file |
|-------|---------|--------------|
| `robot` | `couchbase/server` | `docker-compose.yml` |
| `robot-ea` | `couchbase/enterprise-analytics` | `docker-compose.ea.yml` |

## How it works

A docker-compose stack creates the product service, and a robot service (playwright/node/chromium), the robot service opens the product UI, walks through the initial setup wizard resizing the viewport and taking screenshots as it goes. Screenshots are stored in the `output` directory (./output) when complete.

Output filenames must stay in sync with the image URLs in the corresponding `generate/resources/<product>/README.md`, since that README is what gets published as the Docker Hub description.

Note: we use `--renew-anon-volumes` to ensure the product container is coming up on clean volumes and not bringing up an initialised cluster on subsequent runs. With `--exit-code-from robot` we ensure the robot is responsible for the lifecycle of the stack.

## Example Usage: Couchbase Server

`docker-compose build && TAG=6.6.2 docker-compose up --renew-anon-volumes --exit-code-from robot`

## Example Usage: Enterprise Analytics

The Enterprise Analytics stack also brings up [Adobe S3Mock](https://github.com/adobe/S3Mock) to stand in for S3-compatible blob storage, and the robot configures the cluster against it, mirroring the quickstart in the Docker Hub overview.

```
TAG=2.2.1 docker compose -f docker-compose.ea.yml build
TAG=2.2.1 docker compose -f docker-compose.ea.yml up --renew-anon-volumes --abort-on-container-exit --exit-code-from robot
```

The run takes several minutes: after initializing the cluster the robot waits for the Analytics service to warm up, then loads `travel-sample` and waits for ingestion to finish before running the sample query.

Notes:

* `robot-ea` builds on `mcr.microsoft.com/playwright`, which is multi-arch, so it runs natively on both amd64 and arm64. The older `robot` image is amd64-only and runs under emulation on Apple Silicon.
* On failure the robot writes `output/FAILURE.png` and dumps the page text, so you can see which step broke when the UI changes.
* If a step times out on an element that does exist, check whether the input is visually hidden behind a styled label. Several of them are, which is why the robots click those via the DOM and wait for `state: 'attached'` rather than for visibility.

## Uploading the output

Once the images have been generated and you have visually confirmed they are correct, upload them to the origin bucket behind `https://d774lla4im6mk.cloudfront.net`.

Couchbase Server images live at the root of the bucket:

`aws s3 cp --recursive output/ s3://cb-dockerhub-screenshots-origin`

Enterprise Analytics images live under the `ea/` prefix:

`aws s3 cp --recursive output/ s3://cb-dockerhub-screenshots-origin/ea/`

The objects carry no `Cache-Control` header, so CloudFront keeps serving the previous images until its default TTL expires. Because the filenames do not change between releases, invalidate the prefix after uploading:

`aws cloudfront create-invalidation --distribution-id <id> --paths '/ea/*'`

Look up `<id>` with:

`aws cloudfront list-distributions --query "DistributionList.Items[?contains(DomainName, 'd774lla4im6mk')].Id" --output text`
