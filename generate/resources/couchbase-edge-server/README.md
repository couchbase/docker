
This README will guide you through running Couchbase Edge Server with Docker Containers.

For configuration references or any additional information, please visit [Edge Server documentation site](https://docs.couchbase.com/edge-server/current/index.html).

# Running Edge Server with Docker

1. Start Edge Server container with the default configuration file
```
$ docker run --name edge-server couchbase/edge-server
```

2. Start Edge Server container with an external configuration file
```
$ docker run --name edge-server -d -v /tmp/my-edge-server.json:/tmp/my-edge-server.json couchbase/edge-server /tmp/my-edge-server.json
```
