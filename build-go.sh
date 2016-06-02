#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

GITHUB_DIR=$(cd $(dirname "${BASH_SOURCE[0]}")/../.. && pwd )
SYMB_CTX=$GITHUB_DIR/openshift/origin/Godeps/_workspace/src/github.com/

if [[ ! -L ${SYMB_CTX}tangfeixiong ]]; then
	echo "Create link with tangfeixiong"
	ln -s $GITHUB_DIR/tangfeixiong $SYMB_CTX
fi

if [[ ! -L ${SYMB_CTX}openshift/origin ]]; then
	echo "Create link with openshift/origin"
	ln -s $GITHUB_DIR/openshift/origin ${SYMB_CTX}openshift/
fi

GOPATH=$GITHUB_DIR/openshift/origin/Godeps/_workspace
GOBIN=/data/bin
GOPATH=$GOPATH go build -o "$GOBIN/getting-started-2" -v github.com/tangfeixiong/go-to-cloud-1/cmd/staging/kubernetes/openshift-dockerbuilder/getting-started-2

unlink ${SYMB_CTX}tangfeixiong
unlink ${SYMB_CTX}openshift/origin
 
