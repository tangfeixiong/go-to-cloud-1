package osoc

import (
	"bytes"
	//"log"
	"testing"

	"github.com/helm/helm-classic/codec"
	//buildapi "github.com/openshift/origin/pkg/build/api/v1"
	//projectapi "github.com/openshift/origin/pkg/project/api/v1"

	//"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

var (
	factory = &integrationFactory{server: _grpc_client_endpoint}
)

func TestProject_retrieve(t *testing.T) {
	//  cc, err := grpc.Dial(_grpc_client_endpoint, grpc.WithInsecure())
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

func TestDocker_Builder(t *testing.T) {
	exam := dockerbuilder_example()
	util := &DockerBuildRequestDataUtility{}
	data, err := util.Builder(exam["Project"].(string), exam["Name"].(string)).
		Dockerfile(exam["Dockerfile"].(string)).
		Git(exam["GitURI"].(string), exam["GitRef"].(string), exam["GitPath"].(string)).
		DockerBuildStrategy(_override_baseimage, "", ".", true, false).
		DockerBuildOutputOption(exam["DockerPushRepo"].(string), _dockerpush_secret).RequestDataForPOST()
	if err != nil {
		t.Fatal(err)
	}
	result, err := factory.CreateDockerBuildIntoImage(data)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Log("Received nothing")
	} else if result.Raw != nil && len(result.Raw.ObjectJSON) > 0 {
		t.Logf("Result: %s", string(result.Raw.ObjectJSON))
	} else {
		t.Logf("Received: %+v", result)
	}
}

func TestDockerBuild_retrieve(t *testing.T) {
	//  cc, err := grpc.Dial(_grpc_client_endpoint, grpc.WithInsecure())
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
	respBuild, err := factory.QueryDockerBuilderIntoBuilding(reqBuild)
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

func TestDirect_origindockerbuild(t *testing.T) {
	reqBuild := dockerbuilder_data()
	respBuild, err := factory.CreateDockerBuildIntoImage(reqBuild)
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(respBuild); err != nil {
		t.Fatal(err)
	}
	t.Logf("Received: \n%+v", b.String())
}
