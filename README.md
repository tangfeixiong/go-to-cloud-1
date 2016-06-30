# Instruction

## Prerequisites

* Primary dependency

[Kubernetes v1.3.0](https://github.com/kubernetes/kubernetes)

[OpenShift Origin](https://github.com/openshift/origin)

* Networking dependency

[gRPC](https://github.com/grpc/grpc)

[Google Protocol Buffers](https://github.com/google/protobuf)

[Protobuf Golang Generater](https://github.com/golang/protobuf)

## Get gRPC

## Get, Make and Install Google Protobuf

## Get, Build and Install Golang Protobuf

### Generate Golang Code

Enter

    [vagrant@localhost go-to-cloud-1]$ cd _proto/

Generate

* Golang

Generate message and stub

    [vagrant@localhost _proto]$ protoc --go_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --proto_path=/usr/local/include --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --proto_path=/data/src/github.com/gengo/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto

* Java

Generage message only

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --java_out=/data/src/github.com/tangfeixiong/go-to-cloud-1/_java_generated/openshift-project-and-build/src/main/java/ /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto 

mvn protobuf:compile -DprotocExecutable=/usr/local/bin/protoc -Dos.detected.classifier=linux-x86_64

### Make

Install

    [vagrant@localhost go-to-cloud-1]$ make install GOFLAGS=-v
    go install -v github.com/tangfeixiong/go-to-cloud-1/cmd/ociacibuilds

...


