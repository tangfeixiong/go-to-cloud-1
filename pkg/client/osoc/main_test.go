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
	_host    = "0.0.0.0:50051"
	_server  = "172.17.4.50:50051"
	_grpcsvr *grpc.Server
)

func startServerGRPC() {

	lstn, err := net.Listen("tcp", _host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	_grpcsvr = grpc.NewServer()
	osopb3.RegisterSimpleServiceServer(_grpcsvr, service.Usrs)

	fmt.Printf("grpc server is running on %s\n", _host)

	if err := _grpcsvr.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("quit application\n")

}

func stopServerGRPC() {
	if _grpcsvr != nil {
		time.Sleep(1000)
		_grpcsvr.Stop()
	}
}

func TestData_mock(t *testing.T) {
	in := internalDockerBuildRequestData()

	t.Log(in)

	util := &DockerBuildRequestDataUtility{}
	data, err := util.BuilderName("default", "example").
		Dockerfile("From busybox\nCMD [\"sh\"]").
		Git("https://github.com/docker-library/busybox", "a0558a9006ce0dd6f6ec5d56cfd3f32ebeeb815f", "uclibc/").
		DockerBuildStrategy("", "", "", true, false).
		DockerBuildOutputOption("hub.qingyuanos.com/admin/busybox:latest", "dockercfg").Result()

	if err != nil {
		t.Fatal(err)
	}
	t.Log(data)
}

func origindockerbuild() *osopb3.DockerBuildRequestData {
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
