FROM ubuntu:20.04

LABEL maintainer="docker@couchbase.com"

ARG UPDATE_COMMAND="apt-get update -y -q"
ARG CLEANUP_COMMAND="rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*"

# Install dependencies:
#  runit: for container process management
#  wget: for downloading .deb
#  tzdata: timezone info used by some N1QL functions
# Additional dependencies for system commands used by cbcollect_info:
#  lsof: lsof
#  lshw: lshw
#  sysstat: iostat, sar, mpstat
#  net-tools: ifconfig, arp, netstat
#  numactl: numactl
RUN set -x \
    && ${UPDATE_COMMAND} \
    && apt-get install -y -q wget tzdata \
      lsof lshw sysstat net-tools numactl bzip2 runit \
    && ${CLEANUP_COMMAND}

ARG CB_RELEASE_URL=https://packages.couchbase.com/releases/6.6.5
ARG CB_PACKAGE=couchbase-server-enterprise_6.6.5-ubuntu20.04_amd64.deb
ARG CB_SHA256=fb2da1880ea993dc7a5695c6fbe14cde62024d865a71a7d44ab653f0f633d4c6
ARG CB_SKIP_CHECKSUM=false
ARG CB_PACKAGE_NAME=couchbase-server
ENV PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Create Couchbase user with UID 1000 (necessary to match default
# boot2docker UID)
RUN groupadd -g 1000 couchbase && useradd couchbase -u 1000 -g couchbase -M

# Install couchbase
# Note: installers for Server prior to 7.0.0 used a method for detecting
# if they were running in a container that caused installation to fail
# in some environments, such as some GitHub actions. Below we patch the
# detection mid-install to work around this issue.
RUN set -x \
    && export INSTALL_DONT_START_SERVER=1 \
    && wget -N --no-verbose $CB_RELEASE_URL/$CB_PACKAGE \
    && { ${CB_SKIP_CHECKSUM} || echo "$CB_SHA256  $CB_PACKAGE" | sha256sum -c - ; } \
    && ${UPDATE_COMMAND} \
    && dpkg --unpack ./$CB_PACKAGE \
    && sed -i -e '/Best heuristic/ a \ \ \ \ [ -d /run/systemd/system ] && return 1; return 0' /opt/couchbase/bin/install/systemd-ctl \
    && dpkg --configure couchbase-server \
    && apt-get install -yf \
    && rm -f ./$CB_PACKAGE \
    && ${CLEANUP_COMMAND} \
    && rm -rf /tmp/* /var/tmp/*

# Update VARIANT.txt to indicate we're running in our Docker image
RUN sed -i -e '1 s/$/\/docker/' /opt/couchbase/VARIANT.txt

# Add runit script for couchbase-server
COPY scripts/run /etc/service/couchbase-server/run
RUN set -x \
    && mkdir -p /etc/runit/runsvdir/default/couchbase-server/supervise \
    && chown -R couchbase:couchbase \
                /etc/service \
                /etc/runit/runsvdir/default/couchbase-server/supervise

# Add dummy script for commands invoked by cbcollect_info that
# make no sense in a Docker container
COPY scripts/dummy.sh /usr/local/bin/
RUN set -x \
    && ln -s dummy.sh /usr/local/bin/iptables-save \
    && ln -s dummy.sh /usr/local/bin/lvdisplay \
    && ln -s dummy.sh /usr/local/bin/vgdisplay \
    && ln -s dummy.sh /usr/local/bin/pvdisplay

# Fix curl RPATH if necessary - if curl.real exists, it's a new
# enough package that we don't need to do anything. If not, it
# may be OK, but just fix it
RUN set -ex \
    &&  if [ ! -e /opt/couchbase/bin/curl.real ]; then \
            ${UPDATE_COMMAND}; \
            apt-get install -y chrpath; \
            chrpath -r '$ORIGIN/../lib' /opt/couchbase/bin/curl; \
            apt-get remove -y chrpath; \
            apt-get autoremove -y; \
            ${CLEANUP_COMMAND}; \
        fi

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
