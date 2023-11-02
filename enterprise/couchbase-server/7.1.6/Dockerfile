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

ARG CB_RELEASE_URL=https://packages.couchbase.com/releases/7.1.6
ARG CB_PACKAGE=couchbase-server-enterprise_7.1.6-linux_@@ARCH@@.deb
ARG CB_SKIP_CHECKSUM=false
ENV PATH=$PATH:/opt/couchbase/bin:/opt/couchbase/bin/tools:/opt/couchbase/bin/install

# Create Couchbase user with UID 1000 (necessary to match default
# boot2docker UID)
RUN groupadd -g 1000 couchbase && useradd couchbase -u 1000 -g couchbase -M

# Install couchbase
RUN set -x \
    && export INSTALL_DONT_START_SERVER=1 \
    && dpkgArch="$(dpkg --print-architecture)" \
    && case "${dpkgArch}" in \
         'arm64') \
           CB_SHA256=d4462e7228c372f761bd83f96fa63a7211544df885c2d2e065202ff663dad6f9 \
           ;; \
         'amd64') \
           CB_SHA256=abf410a1f97dd14171cd260d4abb853be003db5ec0a44c8324c846068eb90ce7 \
           ;; \
       esac \
    && CB_PACKAGE=$(echo ${CB_PACKAGE} | sed -e "s/@@ARCH@@/${dpkgArch}/") \
    && wget -N --no-verbose $CB_RELEASE_URL/$CB_PACKAGE \
    && { ${CB_SKIP_CHECKSUM} || echo "$CB_SHA256  $CB_PACKAGE" | sha256sum -c - ; } \
    && ${UPDATE_COMMAND} \
    && apt-get install -y ./$CB_PACKAGE \
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

# 8091: Cluster administration REST/HTTP traffic, including Couchbase Web Console
# 8092: Views and XDCR access
# 8093: Query service REST/HTTP traffic
# 8094: Search Service REST/HTTP traffic
# 8095: Analytics service REST/HTTP traffic
# 8096: Eventing service REST/HTTP traffic
# 8097: Backup service REST/HTTP traffic
# 9123: Analytics prometheus
# 11207: Data Service (SSL)
# 11210: Data Service
# 11280: Data Service prometheus
# 18091: Cluster administration REST/HTTP traffic, including Couchbase Web Console (SSL)
# 18092: Views and XDCR access (SSL)
# 18093: Query service REST/HTTP traffic (SSL)
# 18094: Search Service REST/HTTP traffic (SSL)
# 18095: Analytics service REST/HTTP traffic (SSL)
# 18096: Eventing service REST/HTTP traffic (SSL)
# 18097: Backup service REST/HTTP traffic (SSL)
EXPOSE 8091 \
       8092 \
       8093 \
       8094 \
       8095 \
       8096 \
       8097 \
       9123 \
       11207 \
       11210 \
       11280 \
       18091 \
       18092 \
       18093 \
       18094 \
       18095 \
       18096 \
       18097

VOLUME /opt/couchbase/var
