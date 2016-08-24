#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

BUILD_ROOT=$(dirname "${BASH_SOURCE[0]}")

if [[ ! -f ${BUILD_ROOT}/docker/apaas ]]; then
	if [[ -d "/work" ]]; then
		GOPATH=/work go build -a -v -o ${BUILD_ROOT}/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
	else 
	    go build -a -v -o $BUILD_ROOT/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
	fi
fi

DOCKER_BUILD_CONTEXT=$(mktemp -d)
DOCKER_IMAGE="hub.qingyuanos.com/admin/apaas"

cp -r build/docker/* $DOCKER_BUILD_CONTEXT

cat <<DF >${DOCKER_BUILD_CONTEXT}/Dockerfile
FROM gliderlabs/alpine
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

LABEL name="apaas" version="0.1" description="openshift origin, GitVersion: v1.3.0-alpha.2"

RUN apk add --update bash ca-certificates git libc6-compat && rm -rf /var/cache/apk/*

ADD apaas /bin/
ADD ./openshift.local.config/ /openshift.local.config/
ADD ./ssl/ /root/.kube/

ENV PORT :50051
ENV KUBE_CONFIG /root/.kube/config
ENV KUBE_CONTEXT kube
ENV OSO_CONFIG /openshift.local.config/master/admin.kubeconfig
ENV OSO_CONTEXT default/20-0-0-64:8443/system:admin
ENV ORIGIN_VERSION v1.3.0-alpha.3

VOLUME ["/root/.kube", "/openshift.local.config"]

EXPOSE 50051

CMD ["/bin/apaas"]
DF

docker build -t ${DOCKER_IMAGE} ${DOCKER_BUILD_CONTEXT}

# Cleanup
rm -rf $DOCKER_BUILD_CONTEXT

docker rmi $(docker images --all --quiet --filter=dangling=true)