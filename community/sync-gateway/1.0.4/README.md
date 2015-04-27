
Docker container for [Couchbase Sync Gateway](https://github.com/couchbase/sync_gateway).

## Quickstart

To run the latest release of Sync Gateway with a configuration file hosted in a github gist:

```
$ docker run -p 4984:4984 -p 4985:4985 couchbase/sync-gateway sync-gw-start -c image -g http://git.io/vfQpe
```
