#!/bin/sh -eu

#build_inside_docker_image=golang:1.8-alpine

echo "Installing git"

apk add --no-cache git

cd src/

echo "Downloading dependencies (but not installing)"

go get -d

echo "Compiling"

# compile statically, so this works on Alpine as well
CGO_ENABLED=0 go build --ldflags '-extldflags "-static"' docserver.go

chmod +x docserver
