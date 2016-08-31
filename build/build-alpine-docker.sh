#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

VER=0.1
TAG=${VER}-$(git rev-parse --short=7 HEAD)
if [[ $# > 0 ]]; then
	TAG=$1
elif [[ ! -z $(git status --porcelain) ]]; then
    #echo "0.1-$(git show-ref --abbrev=7 --heads | awk '{ print $1 }')-$(date +%m%dT%H%M)"
    TAG=${TAG}-$(date +%m%dT%H%M)
fi
DOCKER_IMAGE="hub.qingyuanos.com/admin/apaas:${TAG}"

BUILD_ROOT=$(dirname "${BASH_SOURCE[0]}")
if [[ ! -f ${BUILD_ROOT}/docker/apaas ]]; then
	if [[ -d "/work" ]]; then
		GOPATH=/work CGO_ENABLED=0 go build -a -v -installsuffix CGO -o ${BUILD_ROOT}/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
	else 
	    CGO_ENABLED=0 go build -a -v -installsuffix CGO -o $BUILD_ROOT/docker/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
	fi
fi

DOCKER_BUILD_CONTEXT=$(mktemp -d)
cp -r ${BUILD_ROOT}/docker/* $DOCKER_BUILD_CONTEXT

cat <<DF >${DOCKER_BUILD_CONTEXT}/Dockerfile
FROM gliderlabs/alpine
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

LABEL name="apaas" version="0.1" description="openshift origin, GitVersion: v1.3.0-alpha.2"

RUN apk add --update bash ca-certificates git libc6-compat && rm -rf /var/cache/apk/*

ADD apaas /bin/
ADD ./openshift.local.config/ /openshift.local.config/
ADD ./ssl/ /root/.kube/
RUN cp /root/.kube/kubeconfig /root/.kube/config

ENV KUBE_CONFIG /root/.kube/config
ENV KUBE_CONTEXT kube
ENV OSO_CONFIG /openshift.local.config/master/admin.kubeconfig
ENV OSO_CONTEXT default/20-0-0-64:8443/system:admin
ENV ORIGIN_VERSION v1.3.0-alpha.3
ENV APAAS_GRPC_PORT :50051
ENV GNATSD_ADDRESSES 10.3.0.39:4222
ENV ETCD_V3_ADDRESSES 10.3.0.212:2379

VOLUME ["/root/.kube", "/openshift.local.config"]

EXPOSE 50051

CMD ["/bin/apaas", "--loglevel=5"]
DF

docker build -t ${DOCKER_IMAGE} ${DOCKER_BUILD_CONTEXT}

# Cleanup
rm -rf $DOCKER_BUILD_CONTEXT

docker rmi $(docker images --all --quiet --filter=dangling=true)