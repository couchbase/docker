FROM alpine:3.8

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ARG CO_SHA256=7dffde5070da06d047f1f645ff64ade0d14e8b6a25ef96aba87035e124cbf922
ARG CO_VERSION=1.0.0
ARG CO_RELEASE_URL=https://packages.couchbase.com/kubernetes
ARG CO_PACKAGE=couchbase-autonomous-operator-dist_1.0.0.tar.gz
RUN wget $CO_RELEASE_URL/$CO_VERSION/$CO_PACKAGE && \
    echo "$CO_SHA256  $CO_PACKAGE" | sha256sum -c - && \
    tar xvf $CO_PACKAGE && \
    rm -f $CO_PACKAGE
