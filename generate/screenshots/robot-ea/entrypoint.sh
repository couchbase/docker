#!/bin/bash

EA_HOST=${EA_HOST:-enterprise-analytics}

while ! curl --fail "http://${EA_HOST}:8091" &>/dev/null
do
    echo "Waiting for http://${EA_HOST}:8091"
    sleep 1
done

set -e
node ea.js
