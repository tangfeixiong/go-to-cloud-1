package osoc

import (
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

type Builder interface {
	CreateDockerBuild(ctx context.Context,
		in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error)
}

type builder struct {
	//	out             io.Writer
	//	build           *api.Build
	//	sourceSecretDir string
	//	dockerClient    *docker.Client
	//	dockerEndpoint  string
	//	buildsClient    client.BuildInterface
	server string
}

func NewBuilder(server string) Builder {
	if server == "" {
		return &builder{server: ":50051"}
	}
	return &builder{server: server}
}

func (bd *builder) CreateDockerBuild(ctx context.Context,
	in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {

	conn, err := grpc.Dial(bd.server, grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %v", err)
		return nil, err
	}
	defer conn.Close()
	c := osopb3.NewSimpleServiceClient(conn)

	sendProject := osopb3.FindProjectRequest{Name: in.ProjectName}
	recvProject, err := c.FindProject(ctx, &sendProject)
	if err != nil {
		glog.Fatalf("could not request: %v", err)
		return nil, err
	}

	hobProject, err := codec.JSON.Decode(recvProject.Odefv1RawData).One()
	if err != nil {
		glog.Errorf("could not create decoder into object: %s", err)
		return nil, err
	}

	v1Project := new(projectapiv1.Project)
	if err := hobProject.Object(v1Project); err != nil {
		glog.Errorf("could not create decoder into object: %s", err)
		return nil, err
	}

	out, err := c.CreateDockerImageBuild(ctx, in)
	if err != nil {
		glog.Fatalf("could not request: %v", err)
		return nil, err
	}
	return out, nil
}
