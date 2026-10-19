#!/bin/bash

OI_HOST=${OI_HOST:-operational-insights}

while ! curl --fail "http://${OI_HOST}:8091" &>/dev/null
do
    echo "Waiting for http://${OI_HOST}:8091"
    sleep 1
done

set -e
node oi.js
