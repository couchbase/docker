FROM ubuntu:24.04

LABEL maintainer="docker@couchbase.com"

ARG PKG_COMMAND="apt-get"
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
    && ${PKG_COMMAND} install -y -q wget tzdata tzdata-legacy \
      lsof lshw sysstat net-tools numactl bzip2 \
    && ${CLEANUP_COMMAND}

# Add runit
RUN set -x \
    && apt-get update \
    && apt-get install -y gcc git make \
    && cd /usr/src \
    && git clone https://github.com/couchbasedeps/runit \
    && cd runit \
    && git checkout edb631449d89d5b452a5992c6ffaa1e384fea697 \
    && ./package/compile \
    && cp ./command/* /sbin/ \
    && apt-get purge -y --autoremove gcc git make \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /usr/src/runit

ARG CB_RELEASE_URL=https://packages.couchbase.com/releases/enterprise-analytics/2.2.0
ARG CB_PACKAGE=enterprise-analytics_2.2.0-linux_@@ARCH@@.deb
ENV PATH=$PATH:/opt/enterprise-analytics/bin:/opt/enterprise-analytics/bin/tools:/opt/enterprise-analytics/bin/install

# Create Couchbase user with UID 1000 (necessary to match default
# boot2docker UID)
RUN set -x \
    && if getent group 1000 >/dev/null; then \
          existing_group=$(getent group 1000 | cut -d: -f1); \
          groupmod --new-name couchbase "${existing_group}"; \
       else \
          groupadd -g 1000 couchbase; \
       fi \
    && if getent passwd 1000 >/dev/null; then \
          existing_user=$(getent passwd 1000 | cut -d: -f1); \
          usermod --login couchbase -d /home/couchbase -m -g couchbase -s /bin/sh "${existing_user}"; \
       else \
          useradd couchbase -u 1000 -g couchbase -M -s /bin/sh; \
       fi

# Install enterprise-analytics
RUN \
    set -x \
    && ${UPDATE_COMMAND} \
    && export INSTALL_DONT_START_SERVER=1 \
    && dpkgArch="$(dpkg --print-architecture)" \
    && CB_PACKAGE=$(echo ${CB_PACKAGE} | sed -e "s/@@ARCH@@/${dpkgArch}/") \
    && wget -N --no-verbose $CB_RELEASE_URL/$CB_PACKAGE \
    && ${PKG_COMMAND} install -y ./$CB_PACKAGE \
    && rm -f ./$CB_PACKAGE \
    && ${CLEANUP_COMMAND} \
    && rm -rf /tmp/* /var/tmp/*

# Update VARIANT.txt to indicate we're running in our Docker image
RUN sed -i -e '1 s/$/\/docker/' /opt/enterprise-analytics/VARIANT.txt

# Add runit service script for enterprise-analytics
COPY scripts/run /etc/service/enterprise-analytics/run
RUN set -x \
    && mkdir -p /etc/service/enterprise-analytics/supervise \
    && chown -R couchbase:couchbase \
                /etc/service \
                /etc/service/enterprise-analytics/supervise

# Add dummy script for commands invoked by cbcollect_info that
# make no sense in a Docker container
COPY scripts/dummy.sh /usr/local/bin/
RUN set -x \
    && ln -s dummy.sh /usr/local/bin/iptables-save \
    && ln -s dummy.sh /usr/local/bin/lvdisplay \
    && ln -s dummy.sh /usr/local/bin/vgdisplay \
    && ln -s dummy.sh /usr/local/bin/pvdisplay

# Replace the bundled Python certifi CA bundle with a pinned Mozilla bundle
# from curl.se, bumping certifi's version to match and keeping its dist-info
# RECORD valid
ARG CACERT_DATE=2026-10-20
ARG CACERT_SHA256=796f1389bfea13c0ed78f663b765baaff54a508e5483786bd6ece73c5e3e3cbe
RUN set -ex \
    && PYTHON_LIB_DIR=$(find /opt/enterprise-analytics/lib/python/interp/lib -mindepth 1 -maxdepth 1 -type d -name "python3.*") \
    && [ $(echo "${PYTHON_LIB_DIR}" | wc -l) -eq 1 ] \
    && cd "${PYTHON_LIB_DIR}/site-packages" \
    && OLD_DIST_INFO=$(ls -d certifi-*.dist-info) \
    && OLD_VERSION=$(sed -n 's/^Version: //p' "${OLD_DIST_INFO}/METADATA") \
    && wget -q https://curl.se/ca/cacert-${CACERT_DATE}.pem -O certifi/cacert.pem \
    && echo "${CACERT_SHA256}  certifi/cacert.pem" | sha256sum -c - \
    && VERSION=$(date -d "${CACERT_DATE}" +"%Y.%-m.%-d") \
    && PADDED_VERSION=$(date -d "${CACERT_DATE}" +"%Y.%m.%d") \
    && [ "$(printf '%s\n%s\n' "${OLD_VERSION}" "${VERSION}" | sort -V | tail -1)" = "${VERSION}" ] \
    && DIST_INFO="certifi-${VERSION}.dist-info" \
    && if [ "${OLD_DIST_INFO}" != "${DIST_INFO}" ]; then \
           mv "${OLD_DIST_INFO}" "${DIST_INFO}"; \
           sed -i "s|^${OLD_DIST_INFO}/|${DIST_INFO}/|" "${DIST_INFO}/RECORD"; \
       fi \
    && sed -i -E "s/^Version: .*/Version: ${VERSION}/" "${DIST_INFO}/METADATA" \
    && sed -i -E "s/^__version__ = .*/__version__ = \"${PADDED_VERSION}\"/" certifi/__init__.py \
    && for f in certifi/cacert.pem certifi/__init__.py "${DIST_INFO}/METADATA"; do \
           hash=$(sha256sum "$f" | cut -d' ' -f1 | perl -ne 'chomp; print pack("H*", $_)' \
                  | base64 | tr '+/' '-_' | tr -d '=\n'); \
           size=$(stat -c %s "$f"); \
           sed -i "s|^$f,.*|$f,sha256=${hash},${size}|" "${DIST_INFO}/RECORD"; \
       done

# Add bootstrap script
COPY scripts/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["enterprise-analytics"]

# 8091: Cluster administration REST/HTTP traffic, including Web Console
# 8095: Enterprise Analytics service REST/HTTP traffic
# 9123: Enterprise Analytics prometheus
# 11207: Data Service (SSL)
# 11210: Data Service
# 11280: Data Service prometheus
# 18091: Cluster administration REST/HTTP traffic, including Web Console (SSL)
# 18095: Enterprise Analytics service REST/HTTP traffic (SSL)
EXPOSE 8091 \
       8095 \
       9123 \
       11207 \
       11210 \
       11280 \
       18091 \
       18095

VOLUME /opt/enterprise-analytics/var
