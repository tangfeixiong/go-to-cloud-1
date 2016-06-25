package service

import (
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

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift"
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
	grpcServer *grpc.Server
)

func startServerGRPC() {

	lstn, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	grpcServer = grpc.NewServer()
	openshift.RegisterSimpleServiceServer(grpcServer, Usrs)
	openshift.RegisterSimpleManageServiceServer(grpcServer, Usrs)

	fmt.Printf("grpc server is running on %s\n", "")

	if err := grpcServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	glog.Info("quit application\n")

}

func stopServerGRPC() {
	if grpcServer != nil {
		time.Sleep(1000)
		grpcServer.Stop()
	}
}

func TestFindProjectGRPC(t *testing.T) {
	go startServerGRPC()
	grpcFindProject()
	time.Sleep(1200)
	stopServerGRPC()
}

func grpcFindProject() {
	conn, err := grpc.Dial(":8086", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := openshift.NewSimpleManageServiceClient(conn)

	// Contact the server and print out its response.
	req := &openshift.FindProjectRequest{
		Name: "tangfeixiong",
	}

	resp, err := c.FindProject(context.Background(), req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %+v, %s", resp, string(resp.Odefv1RawData))
}