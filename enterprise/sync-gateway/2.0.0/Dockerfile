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
RUN wget http://packages.couchbase.com/releases/couchbase-sync-gateway/2.0.0/couchbase-sync-gateway-enterprise_2.0.0_x86_64.rpm && \
    rpm -i couchbase-sync-gateway-enterprise_2.0.0_x86_64.rpm && \
    rm couchbase-sync-gateway-enterprise_2.0.0_x86_64.rpm

# Create directory where the default config stores memory snapshots to disk
RUN mkdir /opt/couchbase-sync-gateway/data

# Create sync file for custom config.json or generated from env vars config.json
RUN mkdir /sync

# Copy config generator from environment variables
COPY /scripts/create_config.py /

# Add bootstrap script
COPY scripts/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]

# Expose ports
#  port 4984: public port
EXPOSE 4984
