#!/bin/bash
set -e

staticConfigFile=/opt/enterprise-analytics/etc/couchbase/static_config
restPortValue=8091

# see https://developer.couchbase.com/documentation/server/current/install/install-ports.html
function overridePort() {
    portName=$1
    portNameUpper=$(echo $portName | awk '{print toupper($0)}')
    portValue=${!portNameUpper}

    # only override port if value available AND not already contained in static_config
    if [ "$portValue" != "" ]; then
        if grep -Fq "{${portName}," ${staticConfigFile}
        then
            echo "Don't override port ${portName} because already available in $staticConfigFile"
        else
            echo "Override port '$portName' with value '$portValue'"
            echo "{$portName, $portValue}." >> ${staticConfigFile}

            if [ ${portName} == "rest_port" ]; then
                restPortValue=${portValue}
            fi
        fi
    fi
}

overridePort "rest_port"
overridePort "mccouch_port"
overridePort "memcached_port"
overridePort "query_port"
overridePort "ssl_query_port"
overridePort "fts_http_port"
overridePort "moxi_port"
overridePort "ssl_rest_port"
overridePort "ssl_capi_port"
overridePort "ssl_proxy_downstream_port"
overridePort "ssl_proxy_upstream_port"


[[ "$1" == "enterprise-analytics" ]] && {

    # runit's supervise directory; /run may be an empty tmpfs
    supervise=/run/enterprise-analytics/supervise
    mkdir -p ${supervise} 2>/dev/null || true
    if [ ! -w ${supervise} ]; then
        echo "${supervise} is not writable by the container user"
        echo "Aborting: /run must be writable, e.g. --tmpfs /run"
        exit 1
    fi

    if [ "$(whoami)" = "couchbase" ]; then
        datadir=/opt/enterprise-analytics/var/lib/couchbase
        if [ -d ${datadir} ]; then
            # May be a root-owned, group-writable volume (k8s fsGroup)
            if [ ! -w ${datadir} ]; then
                echo "${datadir} is not writable by UID 1000"
                echo "Aborting as Enterprise Analytics will likely not run"
                exit 1
            fi
        elif [ ! -w /opt/enterprise-analytics/var -o \
            "$(find /opt/enterprise-analytics/var -maxdepth 0 -printf '%u')" != "couchbase" ]; then
            echo "/opt/enterprise-analytics/var is not owned and writable by UID 1000"
            echo "Aborting as Enterprise Analytics will likely not run"
            exit 1
        fi
    fi

    # Ensure running on sufficient hardware
    if [ -e /opt/enterprise-analytics/bin/validate-cpu-microarchitecture.sh ]; then
        source /opt/enterprise-analytics/bin/validate-cpu-microarchitecture.sh
        validate_cpu_microarchitecture
    fi

    echo "Starting Enterprise Analytics -- Web UI available at http://<ip>:$restPortValue"
    echo "and logs available in /opt/enterprise-analytics/var/lib/couchbase/logs"
    exec runsvdir -P /etc/service
}

exec "$@"
