#!/bin/sh -ex
export GOOS=linux
export CGO_ENABLED=0
cd $GOPATH/src/github.com/nokia/danm-utils
go mod download
go install -ldflags "-extldflags '-static'" github.com/nokia/danm-utils/cmd/cleaner
