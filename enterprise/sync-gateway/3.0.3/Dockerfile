FROM centos:centos7

LABEL maintainer="docker@couchbase.com"

ENV PATH $PATH:/opt/couchbase-sync-gateway/bin

# Install dependencies:
#  wget: for downloading Sync Gateway package installer
RUN yum -y update && \
    yum install -y \
    wget && \
    yum clean all

# Install Sync Gateway
RUN SGW_PACKAGE=$(echo "http://packages.couchbase.com/releases/couchbase-sync-gateway/3.0.3/couchbase-sync-gateway-enterprise_3.0.3_@@ARCH@@.rpm" | sed -e "s/@@ARCH@@/$(uname -m)/") && \
    SGW_PACKAGE_FILENAME=$(echo "couchbase-sync-gateway-enterprise_3.0.3_@@ARCH@@.rpm" | sed -e "s/@@ARCH@@/$(uname -m)/") && \
    wget "${SGW_PACKAGE}" && \
    rpm -i ${SGW_PACKAGE_FILENAME} && \
    rm ${SGW_PACKAGE_FILENAME}

# Create directory where the default config stores memory snapshots to disk
RUN mkdir /opt/couchbase-sync-gateway/data

# copy the default config into the container
COPY config/sync_gateway_config.json /etc/sync_gateway/config.json

# Create log dir
RUN set -x \
    && mkdir -p /var/log/sync_gateway \
    && chown sync_gateway:sync_gateway /var/log/sync_gateway

# Add bootstrap script
COPY scripts/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

# If user doesn't specify any args, use the default config
CMD ["/etc/sync_gateway/config.json"]

# Expose ports
#  port 4984: public port
EXPOSE 4984
