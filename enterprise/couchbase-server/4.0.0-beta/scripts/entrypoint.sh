#!/bin/bash
set -e

[[ "$1" == "couchbase-server" ]] && {
    exec /usr/sbin/runsvdir-start
}

exec "$@"
