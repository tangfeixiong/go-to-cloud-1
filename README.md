# Instruction

A Kubernetes/Openshift integration

## Prerequisites

* Primary packages

[Kubernetes](https://github.com/kubernetes/kubernetes)

[Kubernetes for Openshift](https://github.com/openshift/kubernetes)

[OpenShift Origin](https://github.com/openshift/origin)

* Networking packages

[gRPC for Golang](https://github.com/grpc/grpc-go)

[Google Protocol Buffers](https://github.com/google/protobuf)

[Protobuf for Golang](https://github.com/golang/protobuf)

[Protocol Buffers for Go with Gadgets](https://github.com/gogo/protobuf)

>This gogoprotobuf is a fork of golang/protobuf with extra code generation features. It is widely used by such: etcd, cockroachdb, cloudfoundry, kubernetes, docker swarm, nats-io, tidb

[gRPC to JSON proxy for Golang](https://github.com/grpc-ecosystem/grpc-gateway)

## Get, Make and Install Google Protobuf

## Get, Build or Install Golang Protobuf (gogoprotobuf)

## Get, Build or Install Golang gRPC

## Get, Build or Install Golang gRPC-gateway for JSON REST HTTP

## Development

### Generate Golang stub

* Golang stub

Generate gRPC and Protobuf stub

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/usr/local/include --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=/go/src --proto_path=/data/src --proto_path=/work/src/github.com/tangfeixiong/go-to-cloud-1/_proto --gofast_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:pkg/api/proto /work/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/cicd/pb3/*.proto

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/usr/local/include --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=/go/src --proto_path=/data/src --proto_path=/work/src/github.com/tangfeixiong/go-to-cloud-1/_proto --gofast_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:pkg/api/proto /work/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/osopb3/*.proto

Deprecated

    [vagrant@localhost go-to-cloud-1]$ cd _proto/

    [vagrant@localhost _proto]$ protoc --go_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --gofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mpaas/ci/openshift=github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift:../pkg/proto paas/ci/openshift/*.proto

    [vagrant@localhost _proto]$ protoc --proto_path=/usr/local/include --proto_path=/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --proto_path=/go/src --proto_path=/data/src --proto_path=/work/src/github.com/tangfeixiong/go-to-cloud-1/_proto --gofast_out=Mgoogle/protobuf/any.proto=github.com/golang/protobuf/ptypes/any,Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../pkg/proto /work/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto

Generate gateway

    [vagrant@localhost _proto]$ protoc --proto_path=/usr/local/include --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --proto_path=/data/src/github.com/gengo/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto

* Java stub

Generage Protobuf only

    [vagrant@localhost go-to-cloud-1]$ protoc --proto_path=/data/src/github.com/tangfeixiong/go-to-cloud-1/_proto --java_out=/data/src/github.com/tangfeixiong/go-to-cloud-1/_java_generated/openshift-project-and-build/src/main/java/ /data/src/github.com/tangfeixiong/go-to-cloud-1/_proto/paas/ci/openshift/manage_service.proto 

    mvn protobuf:compile -DprotocExecutable=/usr/local/bin/protoc -Dos.detected.classifier=linux-x86_64

### Make

Example GO environment

    [vagrant@localhost go-to-cloud-1]$ echo $GOPATH
    /data:/go:/work
    [vagrant@localhost go-to-cloud-1]$ echo $GOBIN
    /data/bin

* Build

CLI

    [vagrant@localhost go-to-cloud-1]$ GOPATH=/work:/go:/data go build -o /data/bin/c5 -v ./cmd/c5/main.go
    
    [vagrant@localhost go-to-cloud-1]$ KUBECONFIG=/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig c5 g2c run-build
    You can play it soon...

Build with vendoring

    [vagrant@localhost go-to-cloud-1]$ GOPATH=/work go build -a -v -o /data/bin/apaas github.com/tangfeixiong/go-to-cloud-1/cmd/apaas

Or install

    [vagrant@localhost go-to-cloud-1]$ GOPATH=/work go install -v github.com/tangfeixiong/go-to-cloud-1/cmd/apaas

Build into Alpine Docker image

    $ GOPATH=/work CGO_ENABLED=0 go build -o build/docker/apaas --installsuffix cgo github.com/tangfeixiong/go-to-cloud-1/cmd/apaas
    
    $ touch -m build/docker/apaas
    
    $ docker build -t tangfeixiong/gotopaas build/docker/

* Develop

Install without vendor

    [vagrant@localhost go-to-cloud-1]$ unlink vendor && go install -v github.com/tangfeixiong/go-to-cloud-1/cmd/apaas && ln -s _vendor/src vendor

CGO_ENABLED=0 go install -v -a -tags netgo -installsuffix netgo -ldflags "-s -w -X github.com/nats-io/gnatsd/version.GITCOMMIT=`git rev-parse --short HEAD`"

Or using Make tool

    [vagrant@localhost go-to-cloud-1]$ make install GOFLAGS=-v

### Issue

Log

    I0820 22:47:46.501235   20850 request.go:782] Error in request: no kind is registered for the type api.Build
    error: An error occurred saving build revision: no kind is registered for the type api.Build

Code

github.com/tangfeixiong/go-to-cloud-1/vendor/github.com/openshift/origin/pkg/build/builder/common.go

Line 91 - 95

	glog.V(4).Infof("Setting build revision to %#v", build.Spec.Revision.Git)
	_, err := c.UpdateDetails(build)
	if err != nil {
		glog.V(0).Infof("error: An error occurred saving build revision: %v", err)
	}


