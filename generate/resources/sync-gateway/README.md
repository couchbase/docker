
Sync Gateway is REST API server that allows Couchbase Lite mobile databases to synchronize data. It can also be used as a standalone data storage system.

For more information, see the [Couchbase Mobile Overview](http://developer.couchbase.com/mobile).

## Quickstart

```
$ docker run -p 4984:4984 -d couchbase/sync-gateway
```

At this point you should be able to run a curl request against the running Sync Gateway on the port 4984 public port:

```
$ curl http://localhost:4984

{"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

> **Note:** if you are running on OSX using docker-machine, you will need to use the IP address of the running docker machine rather than localhost (eg, http://192.168.99.100)

You can view the Sync Gateway logs via the `docker logs` command:

```
$ docker logs container-id
2016-08-04T17:53:44.513Z Enabling logging: [HTTP+]
2016-08-04T17:53:44.513Z ==== Couchbase Sync Gateway/1.3.0(274;8c3ee28) ====
2016-08-04T17:53:44.513Z requestedSoftFDLimit < currentSoftFdLimit (5000 < 1048576) no action needed
etc ...
```

> **Note:** replace `container-id` above with the actual running container id (ie, `9d004a24a4d1`), which you can find by running `docker ps | grep sync_gateway`.

## Accessing the Sync Gateway Admin port from the container

By default, the port 4985, which is the Sync Gateway Admin port, is only accessible via localhost. This means that it's only accessible *from within the container*.

To access it from within the container, you can get a bash shell on the running container and then use curl to connect to the admin port as follows:

```
$ docker exec -ti container-id bash
```

> **Note:** replace `container-id` above with the actual running container id (ie, `9d004a24a4d1`), which you can find by running `docker ps | grep sync_gateway`.

From the container shell (indicated by the `#` prompt), you can use `curl` to make requests against the running Sync Gateway by running:

```
# curl http://localhost:4985

{"ADMIN":true,"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

## Exposing accessing to the Sync Gateway Admin port to the host

If you need to expose port 4985 to the host machine, you can do so with the following steps.

You may want to stop any currently running Sync Gateway containers with `docker stop container-id`.

Start a container with these arguments:

```
$ docker run -p 4984-4985:4984-4985 -d couchbase/sync-gateway -adminInterface :4985 /etc/sync_gateway/config.json
```

Now, from the *host* machine, you should be able to run a curl request against the admin port of 4985:

```
$ curl http://localhost:4985

{"ADMIN":true,"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.3},"version":"Couchbase Sync Gateway/1.3.0(274;8c3ee28)"}
```

> **Note:** if you are running on OSX using docker-machine, you will need to use the IP address of the running docker machine rather than localhost (eg, http://192.168.99.100)


## Using a volume to persist data across container instances

Sync Gateway uses an in-memory storage backend by default, called [Walrus](https://www.ihasabucket.com/), which has the ability to store snapshots of it's memory contents to disk. *This should never be used in production*, and is included for development purposes.

The default configuration file used by the Sync Gateway Docker container saves Walrus memory snapshots of it's data in the `/opt/couchbase-sync-gateway/data` directory inside the container.  If you want to persist this data *across container instances*, you just need to launch the container with a volume that mounts a local directory on your host, for example, your `/tmp` directory.

```
$ docker run -p 4984:4984 -v /tmp:/opt/couchbase-sync-gateway/data -d couchbase/sync-gateway
```

You can verify it worked by looking in your `/tmp` directory on your host, and you will see a `.walrus` memory snapshot file.

```
$ ls /tmp/*.walrus

db.walrus
```

> **Note:** if you are running on OSX using docker-machine, you will need to either use a directory under `/User` instead of `/tmp`, or run `docker-machine ssh` and run the commands from within the docker-machine Linux VM.

If you add data to a Sync Gateway in a container instance, then stop that container instance and start a new one and mount the volume where the memory snapshots were stored, you should see data from the earlier container instance.

> **WARNING:** if you have multiple container instances trying to write memory snapshots to the same files on the same volumes, it will corrupt the memory snapshots.

## Running with Couchbase Server

Create a docker network called `couchbase`.

```
$ docker network create --driver bridge couchbase
```

Run Couchbase Server in a docker container, and put it in the `couchbase` network.

```
$ docker run --net=couchbase -d --name couchbase-server -p 8091-8094:8091-8094 -p 11210:11210 couchbase
```

Now go to the Couchbase Server Admin UI on [http://localhost:8091](http://localhost:8091) (on OSX, replace localhost with the docker machine host IP) and go through the Setup Wizard.  See [Couchbase Server on Dockerhub](https://hub.docker.com/r/couchbase/server/) for more info.

In Couchbase Server, create a RBAC user for Sync Gateway, in the "Security" section.

There are 3 alternative options to create the JSON config file for Sync Gateway, all described below in the configuration sections.

## Customizing Sync Gateway configuration

### Configure with environment variables

You can use environment variables for 2 purposers: 

* To generate a `config.json` file.
* To make Sync Gateway wait for Couchbase server to be ready before connecting (instead of crashing).

Start the Sync Gateway using the following environment variables, some have defaults already:

* `COUCHBASE_HOST`: set it to the name of your Couchbase server container, e.g. `couchbase-server`.
* `COUCHBASE_PORT`: the Couchbase port if you use a different one. The default value is: `8091`.
* `COUCHBASE_BUCKET_NAME`: the bucket name you want to use Sync Gateway with. The default value is: `app`.
* `COUCHBASE_SYNC_GATEWAY_USER`: set it to the RBAC user you created for Sync Gateway.
* `COUCHBASE_SYNC_GATEWAY_PASSWORD`: set it to the RBAC user password you created for Sync Gateway.
* `COUCHBASE_SYNC_GATEWAY_DATABASE`: the database JSON config field. The default value is: `db`.
* `COUCHBASE_SYNC_GATEWAY_CORS_ORIGINS`: the origins that you want to allow for CORS (cross origin resource sharing). Use it if you are connecting pure browser, web based, frontend applications directly. Set it to a list of origins separated by commas. E.g.: `http://localhost,https://example.com,https://web.example.com`
* `COUCHBASE_SYNC_GATEWAY_DISABLE_GUEST_USER`: set it to `true` or `1` to disable the guest user.
* `COUCHBASE_SYNC_GATEWAY_NUM_INDEX_REPLICAS`: the number of replicas for the Sync Gateway indexes. If you are connecting it to a single Couchbase node, you have to leave it at `0`, you can increase it only with a multi-node cluster. The default value is `0`.
* `COUCHBASE_SYNC_GATEWAY_LOG`: set the type of log for the config. The default value is: `HTTP+`.
* `COUCHBASE_SYNC_GATEWAY_ADMIN_INTERFACE`: to enable the admin interface (only for local development) set it to `:4985`. By default it is disabled.

For example, you can have:

* a Couchbase Server container named `couchbase-server` in the network `couchbase`
* a RBAC user in Couchbase Server for the Sync Gateway named `sync`
* a RBAC user password `secret`

Then you can start the container with:

```bash
docker run --net=couchbase -p 4984:4984 -d -e COUCHBASE_HOST=couchbase-server -e COUCHBASE_SYNC_GATEWAY_USER=sync -e COUCHBASE_SYNC_GATEWAY_PASSWORD=secret couchbase/sync-gateway
```

Verify that Sync Gateway started by running `docker logs container-id` and trying to run a curl request against it:

```
$ curl http://localhost:4984
```

#### Custom JS sync function

If you create a file inside the container in `/sync/sync-function.js` with the pure JavaScript function code, it will be copied within the generated `config.json` file. This allows you to write the custom Sync function in pure JavaScript, with editor support, etc, and have it integrated at Docker runtime.

You could have a `sync-function.js` file with:

```JavaScript
function (doc, oldDoc) {
    requireAdmin();
    channel(doc.channels);
}
```

And then a `Dockerfile` that copies that file to `/sync/` with:

```Dockerfile
FROM couchbase/sync-gateway

COPY ./sync-function.js /sync/
```

And when you run the container, that JavaScript function will be integrated into the generated `config.json` file.

**Note**: the `/sync/sync-function.js` file will be integrated in the `config.json` only when it is generated from environment variables. If you provide an existing `config.json` file with any of the other options it will not be integrated.


### Configure by copying a file to /sync/config.json

If you put a JSON config file in `/sync/config.json` it will be used instead of a generated file.

One option to do that would be with a custom file `config.json`, like:

```JSON
{
  "logging": {"console": {"log_keys": ["*"]}},
  "databases": {
    "db": {
      "server": "http://couchbase-server:8091",
      "bucket": "default",
      "users": { "GUEST": { "disabled": false, "admin_channels": ["*"] } }
    }
  }
}
```

And then a `Dockefile` based on this image that copies it:

```Dockerfile
FROM couchbase/sync-gateway

COPY ./config.json /sync/
```

### Configure with mounted config file volume

Create a `/tmp/my-sg-config.json` file on your host machine, with the following:

```JSON
{
  "logging": {"console": {"log_keys": ["*"]}},
  "databases": {
    "db": {
      "server": "http://couchbase-server:8091",
      "bucket": "default",
      "users": { "GUEST": { "disabled": false, "admin_channels": ["*"] } }
    }
  }
}
```

Start a Sync Gateway container in the `couchbase` network and use the `/tmp/my-sg-config.json` file:

```
$ docker run --net=couchbase -p 4984:4984 -v /tmp:/tmp/config -d couchbase/sync-gateway /tmp/config/my-sg-config.json
```

If you pass a config file as part of the command, it will take precedence over the previous options.

Verify that Sync Gateway started by running `docker logs container-id` and trying to run a curl request against it:

```
$ curl http://localhost:4984
```

### Configure with mounted config file volume, without Couchbase

Prepare the Sync Gateway configuration file on your local machine:

```
$ cd /tmp
$ wget https://raw.githubusercontent.com/couchbase/sync_gateway/master/examples/basic-walrus-bucket.json
$ mv basic-walrus-bucket.json my-sg-config.json
$ vi my-sg-config.json  # make edits
```

Run Sync Gateway and use that configuration file:

```
$ docker run -p 4984:4984 -d -v /tmp:/tmp/config couchbase/sync-gateway /tmp/config/my-sg-config.json
```

> **Note:** If you are running on OSX using docker-machine, you will need to either use a directory under `/Users` instead of `/tmp`, or run `docker-machine ssh` and run the commands from within the docker-machine Linux VM.

### Configure using a URL

Sync Gateway can also load it's configuration directly from a URL.

First upload a configuration file to a publicly available hosting site of your choice (Amazon S3, Github, etc)

Then start Sync Gateway and give it the URL to the raw JSON data:

```
$ docker run -p 4984:4984 -d couchbase/sync-gateway https://raw.githubusercontent.com/couchbase/sync_gateway/master/examples/basic-walrus-bucket.json
```

## Using sgcollect_info

This section only applies if you need to run the `sgcollect_info` tool to collect Sync Gateway diagnostics for Sync Gateway running in a docker container. In order to collect the logs you will need to do the following workaround:

```
$ docker logs container-id > /tmp/sync_gateway.log 2>&1
$ docker exec container-id mkdir -p /var/log/sync_gateway/
$ docker cp /tmp/sync_gateway.log contaner-id:/var/log/sync_gateway/sync_gateway_error.log
```

Once that is done, you can run `sgcollect_info` via:

```
$ docker exec container-id /opt/couchbase-sync-gateway/tools/sgcollect_info --help
```

## Support

[Couchbase Forums](https://forums.couchbase.com/)

## Licensing

Sync Gateway comes in 2 Editions: Enterprise Edition and Community Edition. You can find details on the differences between the 2 and licensing details on the [Sync Gateway Editions](http://developer.couchbase.com/documentation/server/4.5/introduction/editions.html) page.

-	Enterprise Edition -- free for development, testing and POCs. Requires a paid subscription for production deployment. Please refer to the [subscribe](http://www.couchbase.com/subscriptions-and-support) page for details on enterprise edition agreements.

-	Community Edition -- free for unrestricted use for community users.
