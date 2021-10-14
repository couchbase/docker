FROM ubuntu:20.04

LABEL maintainer="docker@couchbase.com"

# Install dependencies:
#  runit: for container process management
#  wget: for downloading .deb
#  chrpath: for fixing curl, below
#  tzdata: timezone info used by some N1QL functions
# Additional dependencies for system commands used by cbcollect_info:
#  lsof: lsof
#  lshw: lshw
#  sysstat: iostat, sar, mpstat
#  net-tools: ifconfig, arp, netstat
#  numactl: numactl
RUN set -x && \
    apt-get update && \
    apt-get install -yq runit wget chrpath tzdata \
    lsof lshw sysstat net-tools numactl bzip2 && \
    apt-get autoremove && apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# http://smarden.org/runit/useinit.html#sysv - at some point the script
# runsvdir-start was moved/renamed to this odd name, so we put it back
# somewhere sensible. This appears to be necessary for Ubuntu 20 but
# not Ubuntu 16.
RUN if [ ! -x /usr/sbin/runsvdir-start ]; then \
        cp -a /etc/runit/2 /usr/sbin/runsvdir-start; \
    fi

ARG CB_VERSION=7.0.2
ARG CB_RELEASE_URL=https://packages.couchbase.com/releases/7.0.2
ARG CB_PACKAGE=couchbase-server-community_7.0.2-ubuntu20.04_amd64.deb
ARG CB_SHA256=f935dcad5c04b553a3c56d782c8d9cb782cbd1cf88878a425ba5f9d45d08120e

ENV PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Create Couchbase user with UID 1000 (necessary to match default
# boot2docker UID)
RUN groupadd -g 1000 couchbase && useradd couchbase -u 1000 -g couchbase -M

# Install couchbase
RUN set -x && \
    export INSTALL_DONT_START_SERVER=1 && \
    wget -N --no-verbose $CB_RELEASE_URL/$CB_PACKAGE && \
    echo "$CB_SHA256  $CB_PACKAGE" | sha256sum -c - && \
    apt-get update && \
    apt-get install -y ./$CB_PACKAGE && \
    rm -f ./$CB_PACKAGE && \
    apt-get autoremove && apt-get clean && \
    rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Update VARIANT.txt to indicate we're running in our Docker image
RUN sed -i -e '1 s/$/\/docker/' /opt/couchbase/VARIANT.txt

# Add runit script for couchbase-server
COPY scripts/run /etc/service/couchbase-server/run
RUN mkdir -p /etc/runit/runsvdir/default/couchbase-server/supervise \
    && chown -R couchbase:couchbase \
                /etc/service \
                /etc/runit/runsvdir/default/couchbase-server/supervise

# Add dummy script for commands invoked by cbcollect_info that
# make no sense in a Docker container
COPY scripts/dummy.sh /usr/local/bin/
RUN ln -s dummy.sh /usr/local/bin/iptables-save && \
    ln -s dummy.sh /usr/local/bin/lvdisplay && \
    ln -s dummy.sh /usr/local/bin/vgdisplay && \
    ln -s dummy.sh /usr/local/bin/pvdisplay

# Fix curl RPATH
RUN chrpath -r '$ORIGIN/../lib' /opt/couchbase/bin/curl

# Add bootstrap script
COPY scripts/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["couchbase-server"]

# 8091: Couchbase Web console, REST/HTTP interface
# 8092: Views, queries, XDCR
# 8093: Query services (4.0+)
# 8094: Full-text Search (4.5+)
# 8095: Analytics (5.5+)
# 8096: Eventing (5.5+)
# 11207: Smart client library data node access (SSL)
# 11210: Smart client library/moxi data node access
# 11211: Legacy non-smart client library data node access
# 18091: Couchbase Web console, REST/HTTP interface (SSL)
# 18092: Views, query, XDCR (SSL)
# 18093: Query services (SSL) (4.0+)
# 18094: Full-text Search (SSL) (4.5+)
# 18095: Analytics (SSL) (5.5+)
# 18096: Eventing (SSL) (5.5+)
EXPOSE 8091 8092 8093 8094 8095 8096 11207 11210 11211 18091 18092 18093 18094 18095 18096
VOLUME /opt/couchbase/var
