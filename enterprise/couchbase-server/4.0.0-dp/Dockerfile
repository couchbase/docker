FROM ubuntu:12.04

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV CB_VERSION		3.0.1
ENV CB_RELEASE_URL	http://packages.couchbase.com/releases
ENV CB_PACKAGE		couchbase-server-community_3.0.1-ubuntu12.04_amd64.deb

# Add couchbase binaries to PATH
ENV PATH $PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Install dependancies
RUN apt-get update && \
    apt-get install -yq runit wget python3-minimal && \
    apt-get autoremove && apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Add bootstrap script
ADD scripts/entrypoint.sh /entrypoint.sh

# Add runnit script for couchbase-server
ADD scripts/run /etc/service/couchbase-server/run

# Install couchbase
RUN wget -N $CB_RELEASE_URL/$CB_VERSION/$CB_PACKAGE && \
    dpkg -i ./$CB_PACKAGE

ENTRYPOINT ["/entrypoint.sh"]
CMD ["couchbase-server"]

EXPOSE 8091 8092 11207 11210 11211 18091 18092
VOLUME /opt/couchbase/var
