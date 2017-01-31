#!/bin/sh -eu

#build_inside_docker_image=golang:1.8-alpine

apk add --no-cache git

cd src/

# download dependencies, but don't install
go get -d

# compile statically, so this works on Alpine as well
CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' docserver.go

chmod +x docserver
