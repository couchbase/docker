FROM couchbase/server:7.0.5

COPY scripts/configure-node.sh /etc/service/config-couchbase/run
RUN chown -R couchbase:couchbase /etc/service
COPY scripts/create-index.json /opt/couchbase
