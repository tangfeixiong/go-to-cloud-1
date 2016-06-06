package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"

	"github.com/golang/glog"

	"google.golang.org/grpc"
)

type ApiServer struct {
	GrpcRootServer *grpc.Server
	RestWebService *restful.WebService
}

var ApiServer *ApiServer

func init() {
	ApiServer = new(ApiServer)
	ApiServer.GrpcRootServer = grpc.NewServer()
	Api.Server.RestWebService = new(restful.WebService)
}

func Run() {
	go runREST()

	lstn, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("grpc server is running on %s\n", "")

	if err := ApiServer.GrpcRootServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}
}

func runREST() {
	// to see what happens in the package, uncomment the following
	restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	//ws := new(restful.WebService)
	ws := ApiServer.RestWebService
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/").To(todo("/api/v1")))

	restful.Add(ws)

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

	log.Printf("start listening on %s:8085", "")
	log.Fatal(http.ListenAndServe(":8085", nil))

	//wsContainer := restful.NewContainer()
	//server := &http.Server{Addr: ":8080", Handler: wsContainer}
	//log.Fatal(server.ListenAndServe())
}

type handler func(request *restful.Request, response *restful.Response)

func todo(string context) handler {
	switch context {
	case "/":
		return func(request *restful.Request, response *restful.Response) {
			response.Write(byte(`
{
  "kind": "APIVersions",
  "versions": [
    "v1"
  ],
  "serverAddressByClientCIDRs": [
    {
      "clientCIDR": "0.0.0.0/0",
      "serverAddress": "172.17.4.50:443"
    }
  ]
}`))
		}
	case "/api/v1":
		return func(request *restful.Request, response *restful.Response) {
			response.Write(byte(`
{
  "kind": "APIResourceList",
  "groupVersion": "v1",
  "resources": [
    {
      "name": "projects",
      "namespaced": false,
      "kind": "project"
    },
    {
      "name": "builds",
      "namespaced": true,
      "kind": "build"
    }
  ]
}`))
		}
	}

	return func(request *restful.Request, response *restful.Response) {
		response.Header().Set(404, "Not Found")
	}
}
