
## Docker Compose: Couchbase Server and Sync Gateway 

### Components

1. Couchbase Server (Single Node)
1. Sync Gateway (Single Node)

### Launch containers

1. `docker-compose up`

### Configure Couchbase Server 

1. Find the local high port that is mapped to Couchbase Server port 8091 by running `docker ps | grep -i couchbase-server` and looking for something that looks like `0.0.0.0:32771->8091/tcp`.
1. Open a web browser and point to `localhost:32771`, replacing `32771` with the actual high port found in previous command
1. Configure Couchbase Server via Web Admin UI.
    * You will probably need to reduce the memory because it tries to take the entire memory of the Linux host machine, even though only a portion of it is available to the container.
    * Create a "default" bucket, since that is what the Sync Gateway configuration tries to connect to

### Verify Sync Gateway

Initially, Sync Gateway will be in a retry loop trying to connect to Couchbase Server, and you can expect to see `Opening Couchbase database default on <http://couchbase-server:8091> as user "default"` repeated in the logs.

Once Couchbase Server has been configured, you should see the following in the Sync Gateway logs:

```
sync-gateway_1      | 2017-04-10T21:14:23.137Z Opening Couchbase database default on <http://couchbase-server:8091> as user "default"
sync-gateway_1      | 2017-04-10T21:14:28.261Z Opening Couchbase database default on <http://couchbase-server:8091> as user "default"
sync-gateway_1      | _time=2017-04-10T21:14:28.286+00:00 _level=INFO _msg=Non-healthy node; node details:
sync-gateway_1      | _time=2017-04-10T21:14:28.286+00:00 _level=INFO _msg=Hostname=172.18.0.2:8091, Status=warmup, CouchAPIBase=http://172.18.0.2:8092/default%2Be573413a4d6119a6b5532f276ee4bd64, ThisNode=true
..... snipped logs .....
sync-gateway_1      | 2017-04-10T21:14:29.156Z Starting admin server on :4985
sync-gateway_1      | 2017-04-10T21:14:29.162Z Starting server on :4984 ...

```

1. Find the local high port that is mapped to Sync Gateway port 4984 by running `docker ps | grep -i sync-gateway` and looking for something that looks like `0.0.0.0:32772->4984/tcp`.
1. Open a web browser and point to `localhost:32772`, replacing `32772` with the actual high port found in previous command.  You should see: `{"couchdb":"Welcome","vendor":{"name":"Couchbase Sync Gateway","version":1.4},"version":"Couchbase Sync Gateway/1.4.0(2;9e18d3e)"}`