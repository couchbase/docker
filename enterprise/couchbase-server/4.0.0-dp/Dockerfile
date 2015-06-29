FROM ubuntu:12.04

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

# Install dependancies
RUN apt-get update && \
    apt-get install -yq runit wget && \
    apt-get autoremove && apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Add bootstrap script
ADD scripts/entrypoint.sh /

# Add runnit script for couchbase-server
ADD scripts/run /etc/service/couchbase-server/run

ENV CB_VERSION=feature/dp \
    CB_RELEASE_URL=http://packages.couchbase.com/releases \
    CB_PACKAGE=couchbase-server-enterprise-4.0.0-dp-ubuntu12.04_amd64.deb \
    PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install


# Install couchbase
RUN wget -N $CB_RELEASE_URL/$CB_VERSION/$CB_PACKAGE && \
    dpkg -i ./$CB_PACKAGE

ENTRYPOINT ["/entrypoint.sh"]
CMD ["couchbase-server"]

EXPOSE 8091 8092 11207 11210 11211 18091 18092
VOLUME /opt/couchbase/var
