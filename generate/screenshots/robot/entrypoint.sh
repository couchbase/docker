#!/bin/bash

npm install --save playwright

while ! curl --fail http://couchbase:8091 &>/dev/null
do
    echo "Waiting for http://couchbase:8091"
    sleep 1
done

set -e
node app.js
