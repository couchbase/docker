FROM centos:centos7

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV PATH $PATH:/opt/couchbase-sync-gateway/bin

# install wget
RUN yum -y update && \
        yum install -y \
	wget 

# wget sync gateway URL
# eg: http://cbmobile-packages.s3.amazonaws.com/couchbase-sync-gateway-enterprise_1.3.0-274_x86_64.rpm
RUN wget http://packages.couchbase.com/releases/couchbase-sync-gateway/1.3.0/couchbase-sync-gateway-community_1.3.0-274_x86_64.rpm  

# rpm install it
# eg, couchbase-sync-gateway-enterprise_1.3.0-274_x86_64.rpm
RUN rpm -i couchbase-sync-gateway-community_1.3.0-274_x86_64.rpm

ENTRYPOINT ["/opt/couchbase-sync-gateway/bin/sync_gateway"]

# Expose ports
EXPOSE 4984 4985

