FROM alpine:latest
MAINTAINER Levente Kale <levente.kale@nokia.com>

COPY cleaner /usr/local/bin/cleaner

RUN adduser -u 666 -D -H -s /sbin/nologin cleaner \
&&  chown root:cleaner /usr/local/bin/cleaner \
&&  chmod 750 /usr/local/bin/cleaner

USER cleaner

WORKDIR /
ENTRYPOINT ["/usr/local/bin/cleaner"]
