#!/bin/bash

#######################################
# function retry
# Runs a command for several times and gives up with
# an error message after max retries.
#
# Arguments:
#   maxretries      number of retries before giving up
#   errormessage    error message for each failed try
#   command         command to be executed
#
# Returns:
#   1 on failure. 0 on success
#
# Author: Christopher Hauser <post@c-ha.de>
#######################################
function retry(){
    maxretries=$1; shift
    errormessage=$1; shift
    command="$@"

    tries=0
    while (( $tries >= 0 )); do
        $command
        if [[ $? != 0 ]]; then
            let tries+=1
            if (( $tries < $maxretries )); then
                echo "$errormessage"
                echo "$tries retries. retry in 5s." >&2
                sleep 5
            else
                echo "$errormessage"
                echo "$tries retries. Will give up." >&2
                return 1
            fi
        else
            tries=-1 # try success
            return 0
        fi
    done   
}

#######################################
# function initializeAddHost
# Inizialize a couchbase server: join a cluster
#
# Arguments:
#   none
#
# Returns:
#   1 on failure. 0 on success
#
# Author: Christopher Hauser <post@c-ha.de>
#######################################
function initializeAddHost(){

    # validate if host is part of cluster, then skip
    if [[ $(couchbase-cli server-list -c $CLUSTER -u $USER -p $PASS | grep $(hostname --ip-address):$PORT) ]]; then
        echo "server is already part of the cluster. skipping."
        return 0;
    fi

    # add host to cluster
    couchbase-cli server-add -c $CLUSTER -u $USER -p $PASS \
        --server-add=$(hostname --ip-address):$PORT \
        --server-add-username=$USER \
        --server-add-password=$PASS
    if [[ $? != 0 ]]; then 
        echo "failed to add server to cluster." >&2
        return 1                
    fi

    # if auto rebalance is activated, start rebalancing cluster
    if [[ "$AUTOREBALANCE" == "true" ]]; then
        couchbase-cli rebalance -c $CLUSTER -u $USER -p $PASS                     
        if [[ $? != 0 ]]; then 
            echo "failed to rebalance cluster. please trigger again manually." >&2
        fi
    fi
}

#######################################
# function initializeAddBuckets
# Inizialize a couchbase server: add buckets to cluster
#
# Arguments:
#   none
#
# Returns:
#   1 on failure. 0 on success
#
# Author: Christopher Hauser <post@c-ha.de>
#######################################
function initializeAddBuckets(){
    
    # validate if all env vars are present    
    params=("BUCKETSIZES BUCKETS")
    for var in $params ; do 
        if [ -z ${!var+x} ]; then 
            echo "$var is unset. skipping initialization of buckets."
            return 0
        fi        
    done

    i=0
    sizes=($BUCKETSIZES)
    for bucket in $BUCKETS; do 
        if [[ $(couchbase-cli bucket-list -c $CLUSTER -u $USER -p $PASS | grep -Fx $bucket) ]]; then
            echo "bucket $bucket exists. skipping."
            continue
        fi
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

#######################################
# function initializeCluster
# Inizialize a couchbase server: create or join a cluster,
# create buckets if specified.
#
# Arguments:
#   none
#
# Returns:
#   1 on failure. 0 on success
#
# Author: Christopher Hauser <post@c-ha.de>
#######################################
function initializeCluster(){

    # validate if all env vars are present
    params=("CLUSTER USER PASS PORT RAMSIZEMB RAMSIZEINDEXMB RAMSIZEFTSMB")
    for var in $params ; do 
        if [ -z ${!var+x} ]; then 
            echo "$var is unset. skipping initialization of cluster."
            return 0
        fi        
    done

    # initialize cluster
    echo "try initialization"
    initOutput=$(couchbase-cli cluster-init -c $CLUSTER -u $USER -p $PASS \
            --cluster-username=$USER \
            --cluster-password=$PASS \
            --cluster-port=$PORT \
            --services=data,index,query,fts \
            --cluster-ramsize=$RAMSIZEMB \
            --cluster-index-ramsize=$RAMSIZEINDEXMB \
            --cluster-fts-ramsize=$RAMSIZEFTSMB \
            --index-storage-setting=default)
    if [[ $? != 0 ]]; then 
        echo $initOutput
        echo "failed to initialize cluster." >&2
        return 1
    fi
    echo $initOutput
    
    # add host to cluster
    echo "try to add host (if not already added) to cluster"
    retry 5 "failed to add host to cluster" initializeAddHost
    if [[ $? != 0 ]]; then 
        echo "failed to add host to cluster." >&2
        return 1
    fi
}

#######################################
# function main
# Main function to bootstrap Couchbase in a
# Docker container
#
# Arguments:
#   none
#
# Returns:
#   none
#
# Author: Christopher Hauser <post@c-ha.de>
#######################################
function main(){
    echo "Starting Couchbase Server -- Web UI available at http://<ip>:8091 and logs available in /opt/couchbase/var/lib/couchbase/logs"
    exec /usr/sbin/runsvdir-start &
    if [[ $? != 0 ]]; then 
        echo "failed to start couchbase. exiting now." >&2
        exit 1
    fi

    echo "now initializing couchbase ..."
    retry 5 "cannot initialize couchbase" initializeCluster
    if [[ $? != 0 ]]; then 
        echo "init failed. exiting now." >&2
        exit 1
    fi

    echo "try to create buckets"
    retry 5 "failed to add buckets" initializeAddBuckets
    if [[ $? != 0 ]]; then 
        echo "failed to add buckets to cluster.  please trigger again manually." >&2
    fi

    # wait for couchbase server
    echo "now wait for couchbase forever"
    wait
}

#######################################
# Start the magic ...
#######################################
if [[ "$1" == "couchbase-server" ]]; then
    main
else
    exec "$@"
fi