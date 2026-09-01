Couchbase Server Sandbox
========================

# Running

You should need nothing installed on your machine except Docker. Type:

    docker run -d --name couchbase-sandbox -p 8091-8094:8091-8094 -p 11210:11210 -v $(pwd)/couchbase_demo:/opt/couchbase/var couchbase/server-sandbox:SANDBOX_VERSION

(Replace "SANDBOX_VERSION" with the version of Couchbase Server you wish to explore.)

Then visit [http://localhost:8091/](http://localhost:8091/) for the Server user interface. The login credentials are Administrator / password. You can also
see this information by typing "docker logs couchbase-sandbox".

This image is configured as follows:

    * Couchbase Server
    * All services enabled with small but sufficient memory quotas
    * travel-sample bucket installed
    * FTS index named "hotels"
    * Admin credentials: Administrator / password
    * RBAC user with admin access to travel-sample bucket, with
      credentials admin / password
