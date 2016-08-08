# Instruction

## Prerequisites

* Primary dependency

[Kubernetes](https://github.com/kubernetes/kubernetes)

[Kubernetes for Openshift](https://github.com/openshift/kubernetes)

[OpenShift Origin](https://github.com/openshift/origin)

* Networking dependency

[gRPC for Golang](https://github.com/grpc/grpc-go)

[Google Protocol Buffers](https://github.com/google/protobuf)

[Protobuf for Golang](https://github.com/golang/protobuf)

[gRPC to JSON proxy for Golang](https://github.com/grpc-ecosystem/grpc-gateway)

## Get, Make and Install Google Protobuf

## Get, Build and Install Golang Protobuf

## Get Golang gRPC

### Generate Golang stub code

Generate

    [vagrant@localhost go-to-cloud-1]$ cd _proto/

* Golang

Generate message and stub

    [vagrant@localhost _proto]$ protoc --go_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --proto_path=/usr/local/include --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=/go/src --proto_path=/data/src --proto_path=/work/src/github.com/tangfeixiong/go-to-cloud-1/_proto --gofast_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../pkg/proto /work/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto

Latest

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/usr/local/include --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=/go/src --proto_path=/data/src --proto_path=/work/src/github.com/tangfeixiong/go-to-cloud-1/_proto --gofast_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:pkg/api/proto /work/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/osopb3/*.proto

Generate gateway

    [vagrant@localhost _proto]$ protoc --proto_path=/usr/local/include --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --proto_path=/data/src/github.com/gengo/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto

* Java

Generage message only

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --java_out=/data/src/github.com/tangfeixiong/go-to-cloud-1/_java_generated/openshift-project-and-build/src/main/java/ /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto 

    mvn protobuf:compile -DprotocExecutable=/usr/local/bin/protoc -Dos.detected.classifier=linux-x86_64

### Make

Build with vendoring

    [vagrant@localhost go-to-cloud-1]$ GOPATH=/work go build -a -v -o /data/bin/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas


Install without vendor

    [vagrant@localhost go-to-cloud-1]$ mv vendor vendor-exclude && make install GOFLAGS=-v
    go install -v github.com/tangfeixiong/go-to-cloud-1/cmd/apaas && mv vendor-exclude vendor

...


