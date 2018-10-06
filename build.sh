#!/bin/bash -eu

source /build-common.sh

BINARY_NAME="docserver"
COMPILE_IN_DIRECTORY="cmd/docserver"
GOFMT_TARGETS="cmd/"

standardBuildProcess
