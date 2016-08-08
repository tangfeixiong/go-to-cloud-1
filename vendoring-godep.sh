#!/bin/bash -e

pushd $(dirname "${BASH_SOURCE}")

#cd /go/src/k8s.io/kubernetes && git checkout oso-v1.3.0 && cd pkg/apis && ln -s authentication.k8s.io authentication
cur_rev=
cd /data/src/k8s.io/kubernetes/pkg/apis && cur_rev=$(git rev-parse HEAD) && echo ${cur_rev} && \
  mv authentication.k8s.io authentication.k8s.io.exclude && \
  ln -s /data/src/github.com/openshift/origin/vendor/k8s.io/kubernetes/pkg/apis/authentication authentication
cd /data/src/k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed && \
  ln -s /data/src/github.com/openshift/origin/vendor/k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed/authentication authentication
cd /data/src/k8s.io/kubernetes && git add --all && git commit -a -m "temp commit"
cd /data/src/github.com/golang && mv protobuf protobuf.exclude && cd protobuf.exclude && echo $(git rev-parse HEAD)

cd /work/src/github.com/tangfeixiong/go-to-cloud-1

GOPATH=/data:/go:/work godep save -v ./pkg/...

#cd /go/src/k8s.io/kubernetes/pkg/apis && rm authentication && cd ../.. && git checkout tangfeixiong
cd /data/src/k8s.io/kubernetes/pkg/apis && unlink authentication && mv authentication.k8s.io.exclude authentication.k8s.io
cd /data/src/k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset/typed && unlink authentication
git reset --hard ${cur_rev}
cd /data/src/github.com/golang && mv protobuf.exclude protobuf

popd
