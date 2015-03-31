FROM ubuntu:14.04

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ENV GOPATH /opt/go
ENV SGROOT /opt/sync_gateway
ENV GOROOT /usr/local/go
ENV PATH $PATH:$GOPATH/bin:$GOROOT/bin

# Get dependencies
RUN apt-get update && apt-get install -y \
  bc \
  curl \
  emacs \
  git \
  mercurial \
  wget && \
  apt-get clean

# Download and install Go 1.4
RUN wget http://golang.org/dl/go1.4.2.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz && \
    rm go1.4.2.linux-amd64.tar.gz

# install go packages
RUN go get github.com/tools/godep && \
    go get github.com/nsf/gocode && \
    go get code.google.com/p/go.tools/cmd/goimports && \
    go get github.com/golang/lint/golint && \
    go get code.google.com/p/rog-go/exp/cmd/godef

# clone emacs conf
RUN git clone https://github.com/fgimenez/.emacs.d.git /root/.emacs.d && \
    cd /root/.emacs.d && \
    git checkout origin/go

# Build Sync Gateway
RUN mkdir -p $GOPATH && \
    cd /opt && \ 
    git clone https://github.com/couchbase/sync_gateway.git && \
    cd $SGROOT && \
    git submodule update --init --recursive && \
    ./build.sh && \
    cp bin/sync_gateway /usr/local/bin && \
    mkdir -p $SGROOT/data

# Add Sync Gateway launch script
ADD scripts/sync-gw-start /usr/local/bin/