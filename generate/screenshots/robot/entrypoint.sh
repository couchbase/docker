#!/bin/bash

set -Eeuo pipefail

# Wait for server
max_wait_seconds=180
elapsed=0
until curl --fail --silent http://couchbase:8091 >/dev/null 2>&1; do
    if [ "$elapsed" -ge "$max_wait_seconds" ]; then
        echo "Timed out waiting for http://couchbase:8091 after ${max_wait_seconds}s" >&2
        exit 1
    fi
    echo "Waiting for http://couchbase:8091"
    sleep 1
    elapsed=$((elapsed + 1))
done

node app.js
