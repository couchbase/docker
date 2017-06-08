#!/bin/bash
source /opt/util/retry.sh

function initialize(){
    echo "try initialization"
    couchbase-cli cluster-init -c $CLUSTER \
            --cluster-username=$USER \
            --cluster-password=$PASS \
            --cluster-port=$PORT \
            --services=data,index,query,fts \
            --cluster-ramsize=$RAMSIZEMB \
            --cluster-index-ramsize=$RAMSIZEINDEXMB \
            --cluster-fts-ramsize=$RAMSIZEFTSMB \
            --index-storage-setting=default
    if [[ $? != 0 ]]; then 
        echo "failed to initialize cluster." >&2
        return 1
    fi

    echo "try to create buckets"
    i=0
    sizes=($BUCKETSIZES)
    for bucket in $BUCKETS; do 
        size=${sizes[$i]}
        echo "create bucket $bucket with size $size"
        couchbase-cli bucket-create -c $CLUSTER -u $USER -p $PASS \
            --bucket=$bucket \
            --bucket-type=couchbase \
            --bucket-ramsize=$size \
            --bucket-replica=1 \
            --wait
        if [[ $? != 0 ]]; then 
            echo "failed to create buckets." >&2
            return 1
        fi
        let i+=1
    done
}

[[ "$1" == "couchbase-server" ]] && {
    echo "now starting couchbase ..."
    exec /usr/sbin/runsvdir-start &
    if [[ $? != 0 ]]; then 
        echo "failed to start couchbase. exiting now." >&2
        exit 1
    fi

    echo "now initializing couchbase ..."
    retry 5 "cannot initialize couchbase" initialize
    if [[ $? != 0 ]]; then 
        echo "init failed. exiting now." >&2
        exit 1
    fi

    # wait for couchbase server
    echo "now wait for couchbase forever"
    wait
}
exec "$@"