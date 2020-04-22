ARG BUILDER=golang:1.13-alpine3.10
ARG APK_EXTRA_ARGS
FROM ${BUILDER} as base

RUN apk add --no-cache --no-progress ${APK_EXTRA_ARGS} ca-certificates make git bash

ENV GOOS=linux
ENV CGO_ENABLED=0
WORKDIR ${GOPATH}/src/github.com/nokia/danm-utils

FROM base as builder
ADD go.* ./
RUN go mod download
ADD cmd cmd
ADD pkg pkg

RUN go install -ldflags "-extldflags '-static'" github.com/nokia/danm-utils/cmd/cleaner
RUN chmod 750 ${GOPATH}/bin/cleaner
ADD . .
ENTRYPOINT ./scm/build/build.sh

FROM alpine:latest as cleaner
MAINTAINER Levente Kale <levente.kale@nokia.com>

RUN adduser -u 666 -D -H -s /sbin/nologin cleaner
COPY --from=builder --chown=root:cleaner  /go/bin/cleaner /usr/local/bin/cleaner
USER cleaner
WORKDIR /
ENTRYPOINT ["/usr/local/bin/cleaner"]
