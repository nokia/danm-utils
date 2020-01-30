#!/bin/sh -e

echo 'Updating alpine base image'
docker pull alpine:latest

echo 'Building DANM utils builder container'
docker build --no-cache --tag=utils_builder:1.0 scm/build

echo 'Running DANM utils build'
docker run --rm --net=host --name=utils_builder -v $GOPATH/bin:/usr/local/go/bin -v $GOPATH/src:/usr/local/go/src utils_builder:1.0

echo 'Cleaning up DANM utils builder container'
docker rmi -f utils_builder:1.0

echo 'DANM utils libraries successfully built!'