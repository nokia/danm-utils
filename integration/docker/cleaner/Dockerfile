ARG CALICOCTL_VERSION=v3.16.0

FROM alpine:latest
MAINTAINER Levente Kale <levente.kale@nokia.com>
ARG CALICOCTL_VERSION

COPY cleaner /usr/local/bin/cleaner

RUN adduser -u 666 -D -H -s /sbin/nologin cleaner \
&&  chown root:cleaner /usr/local/bin/cleaner \
&&  chmod 750 /usr/local/bin/cleaner
ADD https://github.com/projectcalico/calicoctl/releases/download/${CALICOCTL_VERSION}/calicoctl /go/bin/
RUN chmod +x /go/bin/calicoctl && chown root:cleaner /go/bin/calicoctl

USER cleaner

WORKDIR /
ENTRYPOINT ["/usr/local/bin/cleaner"]
