# Screenshot Robot

This directory contains tooling used to automatically generate screenshots for the couchbase/server docker hub repository. It needs enough free memory for docker to bring up a node, and will probably need tweaked over time as UI elements evolve.

## How it works

A docker-compose stack creates a couchbase/server:${TAG} service, and a robot service (playwright/node/chromium), the robot service opens the couchbase UI, walks through the initial setup wizard resizing the viewport and taking screenshots as it goes.

## Example Usage

`docker-compose build && TAG=6.6.2 docker-compose up --renew-anon-volumes --exit-code-from robot`

This will build the robot image and bring up the stack, screenshots will be stored in the `output` directory (./output) when complete.

Note: we use `--renew-anon-volumes` to ensure the couchbase container is coming up on clean volumes and not bringing up an initialised cluster on subsequent runs. With `--exit-code-from robot` we ensure the robot is responsible for the lifecycle of the stack.

## Uploading the output

Once the images have been generated and you have visually confirmed they are correct, upload the contents of the output folder with:

`aws s3 cp --recursive output/ s3://cb-dockerhub-screenshots-origin`
