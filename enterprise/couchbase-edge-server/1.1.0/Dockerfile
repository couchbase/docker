FROM ubuntu:24.04

LABEL maintainer="docker@couchbase.com"

# Create couchbase user/group with fixed UID/GID 1000 for consistency across environments
# (modifies existing user/group in images which already have UID/GID 1000 - e.g. ubuntu:24.04)
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
ARG EDGE_SERVER_RELEASE_URL="https://packages.couchbase.com/releases/couchbase-edge-server/1.1.0"
ARG EDGE_SERVER_PACKAGE_FILENAME="couchbase-edge-server_1.1.0_@@ARCH@@.deb"
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
