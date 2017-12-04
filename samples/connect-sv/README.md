Couchbase 5.0.0 Sandbox
=======================

# Running

You should need nothing installed on your machine except Docker. Type:

    docker run -d --name couchbase-sandbox -p 8091-8094:8091-8094 -v `pwd`/couchbase_demo:/opt/couchbase/var couchbase/server-internal:sandbox

Then visit [http://localhost:8091/](http://localhost:8091/) for the Server user interface. The login credentials are Administrator / password. You can also
see this information by typing "docker logs couchbase-sandbox".
