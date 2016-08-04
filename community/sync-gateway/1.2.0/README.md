
[Marketing description of what Sync Gateway is goes here]

## Quickstart

```
$ docker run -p 4984:4984 -d couchbase-sync-gateway
```

NOTE: until the official docker image is approved, use the staging repository via:

```
$ docker run -p 4984:4984 -d couchbase/sync-gateway
```

At this point you should be able to run a curl request against the running Sync Gateway on the port 4984 public port:

```
$ curl http://localhost:4984
{"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

NOTE: if you are running on OSX using docker-machine, you will need to use the IP address of the running docker machine rather than localhost (eg, http://192.168.99.100)

You can view the Sync Gateway logs via the `docker logs` command:

```
$ docker logs container-id
2016-08-04T17:53:44.513Z Enabling logging: [HTTP+]
2016-08-04T17:53:44.513Z ==== Couchbase Sync Gateway/1.3.0(274;8c3ee28) ====
2016-08-04T17:53:44.513Z requestedSoftFDLimit < currentSoftFdLimit (5000 < 1048576) no action needed
etc ..
```

## Enabling access to the Admin port



## Getting a "shell" on a running container

```
$ docker exec -ti container-id bash
```

From the container shell, you can use `curl` to make requests against the running Sync Gateway by running:

```
# curl localhost:4984
```


## Using a mounted volume to persist data across container instances


## Running with Couchbase Server

## Running with Couchbase Server + App using Docker Compose

## Using sgcollect_info


