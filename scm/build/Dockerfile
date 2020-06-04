# Set defaults
ARG USERNAME=danm
ARG UID=147
ARG BUILDER=golang:1.13-alpine3.10

FROM ${BUILDER} AS builder
MAINTAINER Levente Kale <levente.kale@nokia.com>
ARG USERNAME
ARG UID

RUN apk add --no-cache ca-certificates make git bash

RUN mkdir -p ${GOPATH}/src/github.com/nokia/danm-utils
COPY / ${GOPATH}/src/github.com/nokia/danm-utils/

WORKDIR ${GOPATH}/src/github.com/nokia/danm-utils

RUN scm/build/build.sh \
 && adduser -u ${UID} -D -H -s /sbin/nologin ${USERNAME} \
 && chown root:${USERNAME} /go/bin/cleaner /go/bin/policer \
 && chmod 0750 /go/bin/*

FROM alpine:latest as cleaner
MAINTAINER Levente Kale <levente.kale@nokia.com>

RUN adduser -u 666 -D -H -s /sbin/nologin cleaner
COPY --from=builder /go/bin/cleaner /usr/local/bin/cleaner
USER cleaner
WORKDIR /
ENTRYPOINT ["/usr/local/bin/cleaner"]

FROM alpine:latest AS policer
MAINTAINER Levente Kale <levente.kale@nokia.com>
ARG USERNAME
ARG UID
RUN adduser -u ${UID} -D -H -s /sbin/nologin ${USERNAME}
WORKDIR /
USER ${USER}
COPY --from=builder /go/bin/policer /usr/local/bin/policer
RUN apk add --no-cache --virtual .tools libcap  \
 && setcap cap_net_raw,cap_net_admin=eip /usr/local/bin/policer \
 && apk del .tools
ENTRYPOINT ["/usr/local/bin/policer"]
