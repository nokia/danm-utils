#!/bin/sh -ex
export GOOS=linux
cd $GOPATH/src/github.com/nokia/danm-utils
$GOPATH/bin/glide install --strip-vendor
go get -d github.com/vishvananda/netlink
go build github.com/nokia/danm-utils/cmd/cleaner
