#!/bin/bash
set -e

LOGFILE_DIR=/var/log/sync_gateway
mkdir -p $LOGFILE_DIR

exec sync_gateway --defaultLogFilePath="${LOGFILE_DIR}" "$@"
