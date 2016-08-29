package osoc

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"

	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/service"
)

func TestMain(m *testing.M) {
	go startServerGRPC()

	// os.Exit() does not respect defer statements
	ret := m.Run()

	stopServerGRPC()

	os.Exit(ret)
}

var (
	_grpc_host = "0.0.0.0:50051"

	_grpc_server *grpc.Server

	_grpc_client_endpoint = "172.17.4.50:50051"
)

func startServerGRPC() {

	lstn, err := net.Listen("tcp", _grpc_host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	_grpc_server = grpc.NewServer()
	osopb3.RegisterSimpleServiceServer(_grpc_server, service.Usrs)

	fmt.Printf("grpc server is running on %s\n", _grpc_host)

	if err := _grpc_server.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("quit application\n")

}

func stopServerGRPC() {
	if _grpc_server != nil {
		time.Sleep(1000)
		_grpc_server.Stop()
	}
}

var (
	_verbose_level = 5
	_cluster       = "notused"

	_project = "tangfx"

	_build_name = "osobuilds"

	_dockerfile = `#netcat hello world http server
FROM alpine/edge
MAINTAINER tangfeixiong <tangfx128@gmail.com>
RUN apk add --update bash ca-certificates libc6-compat netcat-openbsd && rm -rf /var/cache/apk/*
RUN echo "<html><head><title>welcome</title></head><body><h1>hello world</h1></body></html>" >> /tmp/index.html
EXPOSE 80
CMD while true; do nc -l 80 < /tmp/index.html; done`

	_override_baseimage = "gliderlabs/alpine"

	_dockpull_secret = "localdockerconfig"

	_git_hub      = "https://github.com/tangfeixiong/docker-nc.git"
	_git_ref      = "master"
	_context_path = "/latest"

	_override_dockerfile string = "FROM alpine:3.4\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCMD [\"nc\"]"

	_docker_hub string = "172.17.4.50:30005/tangfx/osobuilds:latest"

	_dockerpush_secret string = "localdockerconfig" // "dockerconfigjson-osobuilds"

	_bc map[string]interface{}
)

func TestData_mock(t *testing.T) {
	in := internalDockerBuildRequestData()

	t.Log(in)

	util := &DockerBuildRequestDataUtility{}
	data, err := util.Builder("tangfx", "osobuilds").
		Dockerfile("From busybox\nCMD [\"sh\"]").
		Git("https://github.com/docker-library/busybox", "a0558a9006ce0dd6f6ec5d56cfd3f32ebeeb815f", "uclibc/").
		DockerBuildStrategy("", "", "", true, false).
		DockerBuildOutputOption("172.17.4.50:30005/busybox:latest", "osobuilds-tangfx").RequestDataForPOST()

	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}

func dockerbuilder_example() map[string]interface{} {
	_bc = map[string]interface{}{
		"Name":           _build_name,
		"Project":        _project,
		"GitURI":         _git_hub,
		"GitRef":         _git_ref,
		"GitPath":        _context_path,
		"Dockerfile":     _dockerfile,
		"DockerPushRepo": _docker_hub,
		"DockerPushAuth": map[string]string{
			"Username":      "tangfx",
			"Password":      "tangfx",
			"ServerAddress": "172.17.4.50:30005",
		},
	}
	return _bc
}

func dockerbuilder_data() *osopb3.DockerBuildRequestData {
	reqBuild := &osopb3.DockerBuildRequestData{
		Name:        "osobuilds",
		ProjectName: "tangfx",
		Configuration: &osopb3.DockerBuildConfigRequestData{
			Name:        "osobuilds",
			ProjectName: "tangfx",
			Triggers:    []*osopb3.OsoBuildTriggerPolicy{},
			RunPolicy:   "",
			CommonSpec: &osopb3.OsoCommonSpec{
				Source: &osopb3.BuildSource{
					Dockerfile: "FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]",
					Git: &osopb3.GitBuildSource{
						Uri: "https://github.com/tangfeixiong/docker-nc.git",
					},
					ContextDir: "edge",
				},
				Strategy: &osopb3.BuildStrategy{
					Type: osopb3.BuildStrategy_Docker.String(),
					DockerStrategy: &osopb3.DockerBuildStrategy{
						From: &kapi.ObjectReference{
							Kind: "DockerImage",
							Name: "alpine:latest",
						},
						NoCache:   false,
						ForcePull: false,
					},
				},
				Output: &osopb3.BuildOutput{
					To: &kapi.ObjectReference{
						Kind: "DockerImage",
						Name: "172.17.4.50:30005/tangfx/osobuilds:latest",
					},
					PushSecret: &kapi.LocalObjectReference{
						Name: "localdockerconfig",
					},
				},
			},
			OsoBuildRunPolicy: osopb3.DockerBuildConfigRequestData_Serial,
			Labels:            map[string]string{},
			Annotations:       map[string]string{},
		},
		TriggeredBy: []*osopb3.OsoBuildTriggerCause{},
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	return reqBuild
}
