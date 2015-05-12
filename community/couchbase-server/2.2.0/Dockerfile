FROM centos:centos6.6

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV CB_VERSION=2.2.0 \
    CB_RELEASE_URL=http://packages.couchbase.com/releases \
    CB_PACKAGE=couchbase-server-community_2.2.0_x86_64.rpm \
    PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Install yum dependencies
RUN yum install -y \
    hostname \
    initscripts \
    openssl \
    pkgconfig \
    sudo \
    tar

# Install couchbase
RUN rpm --install $CB_RELEASE_URL/$CB_VERSION/$CB_PACKAGE

# Modify /etc/passwd to add a login shell, otherwise running
#    su - couchbase -c "/opt/couchbase/bin/couchbase-server -- -noinput"
# will give an error: 
#    This account is currently not available.
# This is only an issue on Couchbase Server 3.x, and it's a no-op on 2.x
RUN sed -i -e 's/\/opt\/couchbase:\/sbin\/nologin/\/opt\/couchbase:\/bin\/sh/' /etc/passwd

COPY scripts/couchbase-start /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/couchbase-start"]

EXPOSE 8091 8092 11207 11210 11211 18091 18092
