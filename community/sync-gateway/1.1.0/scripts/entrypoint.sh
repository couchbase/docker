#!/bin/bash
set -e

# Set up symlinks to log into a 'normal' location.
mkdir -p /var/log/sync_gateway
ln -sf /stdout.log /var/log/sync_gateway/sync_gateway_access.log
ln -sf /stderr.log /var/log/sync_gateway/sync_gateway_error.log

# Run SG and use tee to append stdout and stderr to separate logfiles
# Process substitution described here: https://stackoverflow.com/a/692407
exec sync_gateway "$@" > >(tee -a /stdout.log) 2> >(tee -a /stderr.log >&2)
