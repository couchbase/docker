FROM alpine:3.8

MAINTAINER Couchbase Docker Team <docker@couchbase.com>

ARG CO_SHA256={{ .CO_SHA256 }}
ARG CO_VERSION={{ .CO_VERSION }}
ARG CO_RELEASE_URL={{ .CO_RELEASE_URL }}
ARG CO_PACKAGE={{ .CO_PACKAGE }}
RUN wget $CO_RELEASE_URL/$CO_VERSION/$CO_PACKAGE && \
    echo "$CO_SHA256  $CO_PACKAGE" | sha256sum -c - && \
    tar xvf $CO_PACKAGE && \
    rm -f $CO_PACKAGE
