#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

BUILD_ROOT=$(dirname "${BASH_SOURCE[0]}")

if [[ -d "/work" ]]; then
	GOPATH=/work go build -a -v -o ${BUILD_ROOT}/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
else 
    go build -a -v -o $BUILD_ROOT/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
fi

docker build -t tangfeixiong/apaas ${BUILD_ROOT}/docker