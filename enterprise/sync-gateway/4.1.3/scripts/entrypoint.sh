#!/bin/bash
set -e

LOGFILE_DIR=${LOGFILE_DIR:-/var/log/sync_gateway}

exec sync_gateway --defaultLogFilePath="${LOGFILE_DIR}" "$@"
