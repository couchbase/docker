FROM centos:centos7

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV PATH $PATH:/opt/couchbase-sync-gateway/bin

# Install dependencies:
#  wget: for downloading Sync Gateway package installer
RUN yum -y update && \
    yum install -y \
    wget && \
    yum clean all

# Install Sync Gateway
RUN wget http://packages.couchbase.com/releases/couchbase-sync-gateway/1.5.0/couchbase-sync-gateway-enterprise_1.5.0-377_x86_64.rpm && \
    rpm -i couchbase-sync-gateway-enterprise_1.5.0-377_x86_64.rpm && \
    rm couchbase-sync-gateway-enterprise_1.5.0-377_x86_64.rpm

# Create directory where the default config stores memory snapshots to disk
RUN mkdir /opt/couchbase-sync-gateway/data

# copy the default config into the container
COPY config/sync_gateway_config.json /etc/sync_gateway/config.json

# Add bootstrap script
COPY scripts/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

# If user doesn't specify any args, use the default config
CMD ["/etc/sync_gateway/config.json"]

# Expose ports
#  port 4984: public port
EXPOSE 4984
