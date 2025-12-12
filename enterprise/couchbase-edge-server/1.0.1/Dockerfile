FROM ubuntu:22.04

LABEL maintainer="docker@couchbase.com"

# Ubuntu 23.04 and newer have "ubuntu" as the default user, 1000:1000
# Remove it before creating couchbase user as the default user.
RUN set -x \
    && USER1000=$(getent passwd 1000 | cut -d: -f1) \
    && if [ -n "${USER1000}" ]; then userdel --remove ${USER1000}; fi \
    && GROUP1000=$(getent group 1000 | cut -d: -f1) \
    && if [ -n "${GROUP1000}" ]; then groupdel ${GROUP1000}; fi \
    && groupadd --gid 1000 couchbase \
    && useradd couchbase -u 1000 -g couchbase -M -d / -s /usr/bin/bash

# Install dependencies:
RUN set -x \
    && apt update \
    && apt install -y \
           curl \
           lsb-release \
           systemctl \
           wget \
           zlib1g \
    && apt clean

# Install Couchbase-Edge-Server
ARG EDGE_SERVER_RELEASE_URL="https://packages.couchbase.com/releases/couchbase-edge-server/1.0.1"
ARG EDGE_SERVER_PACKAGE_FILENAME="couchbase-edge-server_1.0.1_@@ARCH@@.deb"
RUN set -x \
    && EDGE_SERVER_PACKAGE_FILENAME=$(echo ${EDGE_SERVER_PACKAGE_FILENAME} | sed -e "s/@@ARCH@@/$(dpkg --print-architecture)/") \
    && wget ${EDGE_SERVER_RELEASE_URL}/${EDGE_SERVER_PACKAGE_FILENAME} \
    && apt install -y ./${EDGE_SERVER_PACKAGE_FILENAME} \
    && rm ${EDGE_SERVER_PACKAGE_FILENAME} \
    && rm -f /usr/lib/systemd/system/couchbase-edge-server.service \
    && apt autoremove \
    && apt clean

ENV PATH=$PATH:/opt/couchbase-edge-server/bin

ENTRYPOINT ["couchbase-edge-server"]
CMD ["/opt/couchbase-edge-server/etc/config.json"]

USER couchbase
WORKDIR /opt/couchbase-edge-server/etc

EXPOSE 59840
