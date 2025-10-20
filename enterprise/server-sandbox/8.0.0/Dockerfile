FROM couchbase/server:8.0.0

COPY scripts/configure-node.sh /etc/service/config-couchbase/run
RUN chown -R couchbase:couchbase /etc/service
COPY scripts/create-index.json /opt/couchbase
