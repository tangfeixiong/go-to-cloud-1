package service

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

type apiServer struct {
	Port string
	//Ip   string
}

func NewApiServer() *apiServer {
	s := apiServer{
		Port: "8080",
	}

	if port := os.Getenv("PORT"); port != "" {
		s.Port = port
	}

	return &s
}

func (s *apiServer) Run() {
	// to see what happens in the package, uncomment the following
	restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	res := UserResource{}
	res.Register(nil, nil)
	//wsContainer := restful.NewContainer()
	//ctx.Register(wsContainer)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "/Users/emicklei/Projects/swagger-ui/dist"}
	//swagger.InstallSwaggerService(config)
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	//log.Printf("start listening on localhost:8080")
	//log.Fatal(http.ListenAndServe(":8080", nil))

	fmt.Printf("Start Listening on port %s\n", s.Port)
	if err := http.ListenAndServe(":"+s.Port, nil); err != nil {
		log.Fatal(err)
	}

	//server := &http.Server{Addr: ":8080", Handler: wsContainer}
	//log.Fatal(server.ListenAndServe())
}

var (
	_host        = "0.0.0.0:50051"
	_server      = "172.17.4.50:50051"
	_grpc_server *grpc.Server
)

func startServerGRPC() {

	lstn, err := net.Listen("tcp", _host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	_grpc_server = grpc.NewServer()
	osopb3.RegisterSimpleServiceServer(_grpc_server, Usrs)

	fmt.Printf("grpc server is running on %s\n", _host)

	if err := _grpc_server.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	glog.Info("quit application\n")

}

func stopServerGRPC() {
	if _grpc_server != nil {
		time.Sleep(1000)
		_grpc_server.Stop()
	}
}

func TestDockerBuildRetrieve_gRPC(t *testing.T) {
	go startServerGRPC()

	var err error
	err = dockerbuildretrieve_gRPC()
	if err != nil {
		time.Sleep(600)
		stopServerGRPC()

		t.Fatal(err)
	}

	time.Sleep(600)
	stopServerGRPC()
}

func dockerbuildretrieve_gRPC() error {
	conn, err := grpc.Dial(_server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := osopb3.NewSimpleServiceClient(conn)
	opts := []grpc.CallOption{}

	// Contact the server and print out its response.
	reqProject := &osopb3.ProjectRetrieveRequestData{
		Name: "tangfx",
	}
	respProject, err := c.RetrieveProjectIntoArbitrary(context.Background(), reqProject, opts...)
	if err != nil {
		return err
	}
	if respProject.Raw != nil && len(respProject.Raw.ObjectBytes) > 0 {
		log.Printf("Result: %s", string(respProject.Raw.ObjectBytes))
	} else {
		log.Printf("Received: %+v", respProject)
	}

	//c = osopb3.NewSimpleServiceClient(conn)
	//opts = []grpc.CallOption{}

	reqBuild := &osopb3.DockerBuildRequestData{
		Name:        "nchellohttp",
		ProjectName: "tangfx",
		Configuration: &osopb3.DockerBuildConfigRequestData{
			Name:        "nchellohttp",
			ProjectName: "tangfx",
			/*Triggers:          []*osopb3.OsoBuildTriggerPolicy{},
			RunPolicy:         "",
			CommonSpec:        (*osopb3.OsoCommonSpec)(nil),
			OsoBuildRunPolicy: osopb3.DockerBuildConfigRequestData_Serial,
			Labels:            map[string]string{},
			Annotations:       map[string]string{},*/
		},
		/*TriggeredBy: []*osopb3.OsoBuildTriggerCause{},
		Labels:      map[string]string{},
		Annotations: map[string]string{},*/
	}
	respBuild, err := c.RetrieveDockerBuild(context.Background(), reqBuild, opts...)
	if err != nil {
		return err
	}
	log.Printf("Received: %+v", respBuild)

	return nil
}

func TestDirect_origindockerbuild(t *testing.T) {
	reqBuild := origindockerbuild_data()
	respBuild, err := Usrs.CreateDockerBuilderIntoImage(context.Background(), reqBuild)
	if err != nil {
		t.Fatal(err)
	}
	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(respBuild); err != nil {
		t.Fatal(err)
	}
	t.Logf("Received: \n%+v", b.String())
}

func TestOriginDockerBuild_gRPC(t *testing.T) {
	go startServerGRPC()

	var err error
	err = origindockerbuild_grpc()
	if err != nil {
		time.Sleep(600)
		stopServerGRPC()

		t.Fatal(err)
	}

	time.Sleep(600)
	stopServerGRPC()
}

func origindockerbuild_grpc() error {
	conn, err := grpc.Dial(_server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := osopb3.NewSimpleServiceClient(conn)
	opts := []grpc.CallOption{}

	reqBuild := origindockerbuild_data()
	respBuild, err := c.CreateDockerBuilderIntoImage(context.Background(), reqBuild, opts...)
	if err != nil {
		return err
	}
	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(respBuild); err != nil {
		return err
	}
	fmt.Printf("Received: \n%+v", b.String())
	return nil
}

func origindockerbuild_data() *osopb3.DockerBuildRequestData {
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
