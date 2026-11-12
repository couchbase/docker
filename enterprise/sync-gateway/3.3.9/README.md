
This README will guide you through running Couchbase Sync Gateway with Docker Containers.

[Sync Gateway](https://www.couchbase.com/products/sync-gateway) is a horizontally scalable web server that securely manages the access control and synchronization of data between [Couchbase Lite](https://www.couchbase.com/products/lite) and [Couchbase Server](https://www.couchbase.com/products/server).

For configuration references, or additional information about anything described below, visit the [Sync Gateway documentation site](https://docs.couchbase.com/sync-gateway/current/index.html).

For additional questions and feedback, please visit the [Couchbase Forums](https://forums.couchbase.com/c/mobile/sync-gateway) or [Stack Overflow](https://stackoverflow.com/questions/tagged/couchbase+couchbase-sync-gateway).


# QuickStart with Sync Gateway and Docker

## Running Sync Gateway with Docker

```
$ docker run -d --name sgw -p 4984:4984 couchbase/sync-gateway
```

At this point you should be able to send a HTTP request to the Sync Gateway public port `4984` using curl:

```
$ curl http://localhost:4984
{"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":"3.0"},"version":"Couchbase Sync Gateway/3.0.5(8;godeps/) EE"}
```

## Viewing Logs
You can view the Sync Gateway logs via the `docker logs` command:

```
$ docker logs sgw
2023-02-15T17:24:17.965Z ==== Couchbase Sync Gateway/3.0.5(8;godeps/) EE ====
2023-02-15T17:24:17.965Z [INF] Loading content from [/etc/sync_gateway/config.json] ...
2023-02-15T17:24:17.967Z [INF] Config: Starting in persistent mode using config group "default"
2023-02-15T17:24:17.967Z [INF] Logging: Console to stderr
2023-02-15T17:24:17.967Z [INF] Logging: Files to /var/log/sync_gateway
2023-02-15T17:24:17.967Z [INF] Logging: Console level: info
2023-02-15T17:24:17.967Z [INF] Logging: Console keys: [* HTTP]
etc ...
```


# Customizing Sync Gateway configuration

## Using a Docker volume

**Step - 1 :** Prepare the Sync Gateway configuration file on your local machine:

For versions 3.0.0 and newer:
```
$ cd /tmp
$ wget https://raw.githubusercontent.com/couchbase/sync_gateway/master/examples/startup_config/basic.json
$ mv basic.json my-sg-config.json
$ vi my-sg-config.json  # make edits
```

For older versions:
```
$ cd /tmp
$ wget https://raw.githubusercontent.com/couchbase/sync_gateway/master/examples/release/2.8.3/examples/serviceconfig.json
$ mv serviceconfig.json my-sg-config.json
$ vi my-sg-config.json  # make edits
```

**Step - 2 :** Run Sync Gateway and use that configuration file:

`$ docker run -p 4984:4984 -d -v /tmp:/tmp/config couchbase/sync-gateway /tmp/config/my-sg-config.json`

## Using a URL

Sync Gateway can also load its configuration directly from a public URL.

**Step - 1 :** Upload a configuration file to a publicly available hosting site of your choice (Amazon S3, Github, etc)

**Step - 2 :** Then start Sync Gateway and give it the URL to the raw JSON data:

For versions 3.0.0 and newer:

`$ docker run -p 4984:4984 -d couchbase/sync-gateway https://raw.githubusercontent.com/couchbase/sync_gateway/master/examples/startup_config/basic.json`

For older versions:

`$ docker run -p 4984:4984 -d couchbase/sync-gateway https://raw.githubusercontent.com/couchbase/sync_gateway/release/2.8.3/examples/serviceconfig.json`

# Running with a Couchbase Server container

**Step - 1 :** Create a docker network called `couchbase`:

`$ docker network create --driver bridge couchbase`

**Step - 2 :** Run Couchbase Server in a docker container, and put it in the `couchbase` network.

`$ docker run --net=couchbase -d --name couchbase-server -p 8091-8097:8091-8097 -p 9123:9123 -p 11207:11207 -p 11210:11210 -p 11280:11280 -p 18091-18097:18091-18097 couchbase`

Now go to the Couchbase Server Admin UI on [http://localhost:8091](http://localhost:8091) and go through the Setup Wizard.

See [Couchbase Server on Dockerhub](https://hub.docker.com/r/couchbase/server/) for more info on this process.

**Step - 3 :** Create a configuration file as described in the above config section, and customise the server property, i.e.:

```
{
  "bootstrap": {
    "server": "couchbase-server:8091",
    "server_tls_skip_verify": true,
    "username": "username",
    "password": "password"
  },
  "logging": {
    "console": {
      "enabled": true,
      "log_level": "info",
      "log_keys": ["*"]
    }
  }
}
```

**Step - 4 :** Start a Sync Gateway container in the `couchbase` network and use the configuration file you just wrote:

`$ docker run --net=couchbase -p 4984:4984 -v /tmp:/tmp/config -d couchbase/sync-gateway /tmp/config/my-sg-config.json`

# Admin Port

By default, port `4985`, which is the Sync Gateway Admin port, is only accessible via localhost for security purposes.  
Please refers to https://docs.couchbase.com/sync-gateway/current/rest-api-access.html for details and additional information.

# Collecting logs via sgcollect_info

This section only applies if you need to run the `sgcollect_info` tool to collect Sync Gateway diagnostics for support.

**Step - 1 :** Run the following curl command against the admin port of Sync Gateway to run sgcollect_info and put the zip in your log file path.

`# curl -u <username:password> -X POST http://localhost:4985/_sgcollect_info -H 'Content-Type: application/json' -d '{}'`

You can find more information about the parameters used in this request in the [sgcollect_info documentation](https://docs.couchbase.com/sync-gateway/current/admin-rest-api.html#/server/post__sgcollect_info).


# License

Couchbase software typically comes in two editions: Enterprise Edition and Community Edition. For Couchbase Server, you can find details on the differences between the two and licensing information on the [Couchbase Server Editions](https://docs.couchbase.com/server/current/introduction/editions.html) page.

-	**Enterprise Edition** -- The Enterprise Edition license provides for free for development and testing for Couchbase Enterprise Edition. A paid subscription for production deployment is required. Please refer to the [pricing](https://www.couchbase.com/pricing) page for details on Couchbase’s Enterprise Edition.

-	**Community Edition** -- The Community Edition license provides for free deployment of Couchbase Community Edition. For Couchbase Server, the Community Edition may be used for departmental-scale deployments of up to five node clusters.  It has recently been changed to disallow use of XDCR, which is now an exclusive Enterprise Edition feature.
