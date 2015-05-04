FROM tleyden5iwx/forestdb

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV GOPATH /opt/go
ENV SGROOT /opt/sync_gateway
ENV GOROOT /usr/local/go
ENV PATH $PATH:$GOPATH/bin:$GOROOT/bin

# Get dependencies
RUN yum -y update && \
    yum install -y \
    bc \
    curl \
    emacs-nox \
    git \
    mercurial \
    wget 

# Download and install Go 1.4
RUN wget http://golang.org/dl/go1.4.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz && \
    rm go1.4.2.linux-amd64.tar.gz

# install common go packages
RUN go get github.com/tools/godep && \
    go get github.com/nsf/gocode && \
    go get golang.org/x/tools/cmd/goimports && \
    go get github.com/golang/lint/golint && \
    go get github.com/rogpeppe/godef

# Build Sync Gateway
RUN mkdir -p $GOPATH && \
    cd /opt && \ 
    git clone https://github.com/couchbase/sync_gateway.git && \
    cd $SGROOT && \
    git checkout feature/forestdb_bucket && \	
    git submodule update --init --recursive && \
    ./build.sh && \
    cp bin/sync_gateway /usr/local/bin && \
    mkdir -p $SGROOT/data

# Add Sync Gateway launch script
ADD scripts/sync-gw-start /usr/local/bin/

# Expose ports
EXPOSE 4984 4985