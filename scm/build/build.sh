#!/bin/sh -ex
export GOOS=linux
cd $GOPATH/src/github.com/nokia/danm-utils
glide install --strip-vendor
go get -d github.com/vishvananda/netlink
go install github.com/nokia/danm-utils/cmd/cleaner
