
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

## Accessing the SyncGateway Admin port from the container

By default, the port 4985, which is the Sync Gateway Admin port, is only accessible via localhost.  This means that it's only accessible *from within the container*.

To access it from within the container, you can get a bash shell on the running container and then use curl to connect to the admin port as follows:

```
$ docker exec -ti container-id bash
```

Note: replace `container-id` above with the actual running container id (ie, `9d004a24a4d1`), which you can find by running `docker ps | grep sync_gateway`.

From the container shell (indicated by the `#` prompt), you can use `curl` to make requests against the running Sync Gateway by running:

```
# curl http://localhost:4985

{"ADMIN":true,"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

## Exposing accessing to the SyncGateway Admin port to the host

If you need to expose port 4985 to the host machine, you can do so with the following steps.

You may want to stop any currently running Sync Gateway containers with `docker stop container-id`.

Start a container with these arguments:

```
$ docker run -p 4984-4985:4984-4985 -d couchbase-sync-gateway -adminInterface :4985 /etc/sync_gateway/config.json
```

Now, from the *host* machine, you should be able to run a curl request against the admin port of 4985:

```
$ curl http://localhost:4985

{"ADMIN":true,"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

NOTE: if you are running on OSX using docker-machine, you will need to use the IP address of the running docker machine rather than localhost (eg, http://192.168.99.100)

## Customizing Sync Gateway configuration

### Using a Docker volume

### Using a URL

## Using a volume to persist data across container instances


## Running with Couchbase Server


## Running with Couchbase Server + App using Docker Compose

## Using sgcollect_info


