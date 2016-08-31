package osoc

import (
	"fmt"
	"os"

	//"github.com/helm/helm-classic/codec"
	//projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
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
	logger.SetPrefix("[client/osoc, RetrieveProjectByName] ")

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

func CreateDockerBuilderIntoImage(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {
	logger.SetPrefix("[client/osoc, CreateDockerBuilderIntoImage] ")

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.CreateDockerBuilderIntoImage(ctx, in, opts...)
	} else {
		out, err = c.CreateDockerBuilderIntoImage(context.Background(), in, opts...)
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

func TrackDockerBuild(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {
	logger.SetPrefix("[client/osoc, TrackDockerBuild] ")

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.TrackDockerBuild(ctx, in, opts...)
	} else {
		out, err = c.TrackDockerBuild(context.Background(), in, opts...)
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

func RetrieveDockerBuild(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {
	logger.SetPrefix("[client/osoc, RetrieveDockerBuild] ")

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.RetrieveDockerBuild(ctx, in, opts...)
	} else {
		out, err = c.RetrieveDockerBuild(context.Background(), in, opts...)
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

func RetrieveDockerBuilder(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildConfigRequestData) (out *osopb3.DockerBuildConfigResponseData, err error) {

	return nil, fmt.Errorf("Not implemented")
}

func DeleteDockerBuild(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildRequestData) (out *osopb3.DockerBuildResponseData, err error) {
	logger.SetPrefix("[client/osoc, DeleteDockerBuild] ")

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.DeleteDockerBuild(ctx, in, opts...)
	} else {
		out, err = c.DeleteDockerBuild(context.Background(), in, opts...)
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

func DeleteDockerBuilder(c osopb3.SimpleServiceClient,
	ctx context.Context,
	in *osopb3.DockerBuildConfigRequestData) (out *osopb3.DockerBuildConfigResponseData, err error) {
	logger.SetPrefix("[client/osoc, DeleteDockerBuilder] ")

	opts := []grpc.CallOption{}
	if ctx != nil {
		out, err = c.DeleteDockerBuilder(ctx, in, opts...)
	} else {
		out, err = c.DeleteDockerBuilder(context.Background(), in, opts...)
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

func (itft *integrationFactory) CreateDockerBuilderIntoImage(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[client/osoc, .CreateDockerBuilderIntoImage] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return CreateDockerBuilderIntoImage(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}

func (itft *integrationFactory) TrackDockerBuild(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[client/osoc, .TrackDockerBuild] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return TrackDockerBuild(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}

func (itft *integrationFactory) RetrieveDockerBuild(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[client/osoc, .RetrieveDockerBuild] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return RetrieveDockerBuild(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}

func (itft *integrationFactory) RetrieveDockerBuilder(in *osopb3.DockerBuildConfigRequestData) (*osopb3.DockerBuildConfigResponseData, error) {
	logger.SetPrefix("[client/osoc, .RetrieveDockerBuilder] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return RetrieveDockerBuilder(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}

func (itft *integrationFactory) DeleteDockerBuild(in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[client/osoc, .DeleteDockerBuild] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return DeleteDockerBuild(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}

func (itft *integrationFactory) DeleteDockerBuilder(in *osopb3.DockerBuildConfigRequestData) (*osopb3.DockerBuildConfigResponseData, error) {
	logger.SetPrefix("[client/osoc, .DeleteDockerBuilder] ")
	cc, err := grpc.Dial(itft.server, grpc.WithInsecure())
	if err != nil {
		logger.Printf("Did not connect: %v\n", err)
		return nil, err
	}
	defer cc.Close()

	return DeleteDockerBuilder(osopb3.NewSimpleServiceClient(cc), context.Background(), in)
}
