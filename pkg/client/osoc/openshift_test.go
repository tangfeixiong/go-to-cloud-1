package osoc

import (
	//"log"
	"testing"

	//"github.com/helm/helm-classic/codec"
	//buildapi "github.com/openshift/origin/pkg/build/api/v1"
	//projectapi "github.com/openshift/origin/pkg/project/api/v1"

	//"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

var (
	factory = &integrationFactory{server: _server}
)

func TestData_build(t *testing.T) {
	in := internalDockerBuildRequestData()

	t.Log(in)
}

func TestProject_retrieve(t *testing.T) {
	//  cc, err := grpc.Dial(_server, grpc.WithInsecure())
	//	if err != nil {
	//		log.Fatalf("did not connect: %v", err)
	//	}
	//	defer cc.Close()

	//	c := osopb3.NewSimpleServiceClient(cc)
	//	opts := []grpc.CallOption{}

	// Contact the server and print out its response.
	reqProject := &osopb3.ProjectRetrieveRequestData{
		Name: "gogogo",
	}
	respProject, err := factory.RetrieveProjectByName(reqProject)
	if err != nil {
		t.Fatal(err)
	}
	if respProject.Raw != nil && len(respProject.Raw.ObjectBytes) > 0 {
		t.Logf("Result: %s", string(respProject.Raw.ObjectBytes))
	} else {
		t.Logf("Received: %+v", respProject)
	}

}

func TestDocker_build(t *testing.T) {
	reqBuild := internalDockerBuildRequestData()
	reqBuild.ProjectName = "tangfx"
	reqBuild.Configuration.ProjectName = "tangfx"
	respBuild, err := factory.CreateDockerBuildIntoImage(reqBuild)
	if err != nil {
		t.Fatal(err)
	}
	if respBuild == nil {
		t.Log("Received nothing")
	} else if respBuild.Raw != nil && len(respBuild.Raw.ObjectJSON) > 0 {
		t.Logf("Result: %s", string(respBuild.Raw.ObjectJSON))
	} else {
		t.Logf("Received: %+v", respBuild)
	}
}

func TestDockerBuild_retrieve(t *testing.T) {
	//  cc, err := grpc.Dial(_server, grpc.WithInsecure())
	//	if err != nil {
	//		log.Fatalf("did not connect: %v", err)
	//	}
	//	defer cc.Close()

	//	c := osopb3.NewSimpleServiceClient(cc)
	//	opts := []grpc.CallOption{}

	// Contact the server and print out its response.
	reqBuild := &osopb3.DockerBuildRequestData{
		Name:        "fake",
		ProjectName: "default",
		Configuration: &osopb3.DockerBuildConfigRequestData{
			Name:              "fake",
			ProjectName:       "default",
			Triggers:          []*osopb3.OsoBuildTriggerPolicy{},
			RunPolicy:         "",
			CommonSpec:        (*osopb3.OsoCommonSpec)(nil),
			OsoBuildRunPolicy: osopb3.DockerBuildConfigRequestData_Serial,
			Labels:            map[string]string{},
			Annotations:       map[string]string{},
		},
		TriggeredBy: []*osopb3.OsoBuildTriggerCause{},
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	respBuild, err := factory.RetrieveDockerBuildIntoImage(reqBuild)
	if err != nil {
		t.Fatal(err)
	}
	if respBuild == nil {
		t.Log("Received nothing")
	} else if respBuild.Raw != nil && len(respBuild.Raw.ObjectJSON) > 0 {
		t.Logf("Result: %s", string(respBuild.Raw.ObjectJSON))
	} else {
		t.Logf("Received: %+v", respBuild)
	}
}
