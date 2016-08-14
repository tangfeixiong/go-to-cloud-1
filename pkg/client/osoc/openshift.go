package osoc

import (
	"fmt"
	"log"
	"os"

	//"github.com/helm/helm-classic/codec"
	//projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[client/osoc] ", log.LstdFlags|log.Lshortfile)
)

type integrationFactory struct {
	//	out             io.Writer
	//	build           *api.Build
	//	sourceSecretDir string
	//	dockerClient    *docker.Client
	//	dockerEndpoint  string
	//	buildsClient    client.BuildInterface
	server string
	//osoclient osopb3.SimpleServiceClient
	cc *grpc.ClientConn
}

func NewIntegrationFactory(server string) *integrationFactory {
	if server == "" {
		if v, ok := os.LookupEnv("APAAS_HOST"); !ok {
			server = "127.0.0.1:50051"
		} else {
			server = v
		}
	}
	return &integrationFactory{server: server}
}

func CreateProject(client osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.ProjectCreationRequestData) (out *osopb3.ProjectResponseDataArbitrary, err error) {

	if ctx != nil {
		out, err = client.CreateProjectIntoArbitrary(ctx, in)
	} else {
		out, err = client.CreateProjectIntoArbitrary(context.Background(), in)
	}
	if err != nil {
		logger.Printf("Could not receive result: %v\n", err)
		return nil, err
	}
	if out.Raw != nil && len(out.Raw.ObjectBytes) > 0 {
		logger.Printf("Received: %s\n%s\n", out.Raw.ObjectName, string(out.Raw.ObjectBytes))
	}
	return out, nil

}

func RetrieveProjectByName(client osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.ProjectRetrieveRequestData) (out *osopb3.ProjectResponseDataArbitrary, err error) {

	logger = log.New(os.Stdout, "[client/osoc, RetrieveProjectByName] ", log.LstdFlags|log.Lshortfile)

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = client.RetrieveProjectIntoArbitrary(ctx, in, opts...)
	} else {
		out, err = client.RetrieveProjectIntoArbitrary(context.Background(), in, opts...)
	}
	if err != nil {
		logger.Printf("Could not receive result: %v\n", err)
		return nil, err
	}
	if out.Raw != nil && len(out.Raw.ObjectBytes) > 0 {
		logger.Printf("Received: %s\n%s\n", out.Raw.ObjectName, string(out.Raw.ObjectBytes))
	}
	return out, nil

}

func (itft *integrationFactory) CreateProject(in *osopb3.ProjectCreationRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {

	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()
	client := osopb3.NewSimpleServiceClient(cc)

	return CreateProject(client, context.Background(), in)
}

func (itft *integrationFactory) RetrieveProjectByName(in *osopb3.ProjectRetrieveRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {

	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()
	client := osopb3.NewSimpleServiceClient(cc)

	return RetrieveProjectByName(client, context.Background(), in)
}

func CreateDockerBuildIntoImage(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {

	logger = log.New(os.Stdout, "[client/osoc, CreateDockerBuildIntoImage] ", log.LstdFlags|log.Lshortfile)

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.CreateIntoBuildDockerImage(ctx, in, opts...)
	} else {
		out, err = c.CreateIntoBuildDockerImage(context.Background(), in, opts...)
	}
	if err != nil {
		logger.Printf("Could not receive result: %v", err)
		return nil, err
	}
	if out == nil {
		return
	}
	if out.Raw != nil && len(out.Raw.ObjectJSON) > 0 {
		logger.Printf("Received: %s\n%s\n", out.Raw.ObjectGVK, string(out.Raw.ObjectJSON))
	}
	return out, nil
}

func RetrieveDockerBuildIntoImage(client osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {

	logger = log.New(os.Stdout, "[client/osoc, RetrieveDockerBuildIntoImage] ", log.LstdFlags|log.Lshortfile)

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = client.RetrieveIntoBuildDockerImage(ctx, in, opts...)
	} else {
		out, err = client.RetrieveIntoBuildDockerImage(context.Background(), in, opts...)
	}
	if err != nil {
		logger.Printf("Could not receive result: %v\n", err)
		return nil, err
	}
	if out == nil {
		return
	}
	if out.Raw != nil && len(out.Raw.ObjectJSON) > 0 {
		logger.Printf("Received: %s\n%s\n", out.Raw.ObjectGVK, string(out.Raw.ObjectJSON))
	}
	return out, nil

}

func (itft *integrationFactory) CreateDockerBuildIntoImage(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {

	logger = log.New(os.Stdout, "[client/osoc, .CreateDockerBuildIntoImage] ", log.LstdFlags|log.Lshortfile)

	conn, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer conn.Close()
	c := osopb3.NewSimpleServiceClient(conn)

	p, err := RetrieveProjectByName(c, context.Background(),
		&osopb3.ProjectRetrieveRequestData{Name: in.ProjectName})
	if err != nil {
		return nil, err
	}
	if p != nil && p.ResultingCode != osopb3.K8SNamespacePhase_Active {
		return nil, fmt.Errorf("Project not ready: %v", p)
	}

	if p == nil {
		p, err = CreateProject(c, context.Background(),
			&osopb3.ProjectCreationRequestData{Name: in.ProjectName})
		if err != nil {
			return nil, err
		}
		if p == nil || p.ResultingCode != osopb3.K8SNamespacePhase_Active {
			return nil, fmt.Errorf("Project cloud not create: %v", p)
		}
	}
	//	if p.Raw != nil && len(out.Raw.ObjectBytes) > 0 {
	//		helmobj, err := codec.JSON.Decode(p.Raw.ObjectBytes).One()
	//		if err != nil {
	//			logger.Printf("could not create decoder into object: %s", err)
	//		}
	//		logger.Printf("decoder: %v", helmobj)
	//		osoProject := new(projectapiv1.Project)
	//		if err := helmobj.Object(osoProject); err != nil {
	//			logger.Printf("could not decode into object: %s", err)
	//		}
	//	}

	return CreateDockerBuildIntoImage(c, context.Background(), in)
}

func (itft *integrationFactory) RetrieveDockerBuildIntoImage(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {

	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()
	client := osopb3.NewSimpleServiceClient(cc)

	return RetrieveDockerBuildIntoImage(client, context.Background(), in)
}
