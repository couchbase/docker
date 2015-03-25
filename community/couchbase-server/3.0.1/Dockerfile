FROM centos:centos6.6

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV CB_VERSION		3.0.1
ENV CB_RELEASE_URL	http://packages.couchbase.com/releases
ENV CB_PACKAGE		couchbase-server-community-3.0.1-centos6.x86_64.rpm

# Add couchbase binaries to PATH
ENV PATH $PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Install yum dependencies
RUN yum install -y \
    hostname \
    initscripts \
    openssl \
    pkgconfig \
    sudo \
    tar \
    wget 

# Install couchbase
RUN rpm --install $CB_RELEASE_URL/$CB_VERSION/$CB_PACKAGE

# Modify /etc/passwd to add a login shell, otherwise running
#    su - couchbase -c "/opt/couchbase/bin/couchbase-server -- -noinput"
# will give an error: 
#    This account is currently not available.
# This is only an issue on Couchbase Server 3.x, and it's a no-op on 2.x
RUN sed -i -e 's/\/opt\/couchbase:\/sbin\/nologin/\/opt\/couchbase:\/bin\/sh/' /etc/passwd

ADD scripts/couchbase-start /usr/local/bin/

ENTRYPOINT /usr/local/bin/couchbase-start
