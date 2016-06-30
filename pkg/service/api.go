package service

import (
	"log"
	"net/http"
	"os"
	//"time"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/gengo/grpc-gateway/runtime"

	"golang.org/x/net/context"
)

type user struct {
	Id   string
	Name string
	*basicAuth
}

type basicAuth struct{ username, password string }
type x509Auth struct{ certificateAuthority string }
type tlsAuth struct{ clientCertificate, clientKey string }

type UserResource struct {
	// normally one would use DAO (data access object)
	users     map[string]user
	user      *user
	workspace string

	ContextBase context.Context
	HttpMuxs    []*http.ServeMux
	GatewayMux  *runtime.ServeMux
}

var (
	Usrs   *UserResource
	logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

	blankEntity []byte = []byte("{}")
	//var errorEntity []byte = []byte(`{"description": "Something went wrong. Please contact support at http://support.example.com."}`)
)

func init() {
	Usrs = new(UserResource)
}

func (u *UserResource) Register(ws *restful.WebService, container *restful.Container) {

	/*
	   Services of building Dockerfile and image, ACI
	*/
	if ws == nil {
		ws := new(restful.WebService)
		ws.
			Path("/api/v1").
			Consumes(restful.MIME_JSON, restful.MIME_XML).
			Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well
	}

	ws.Route(ws.Consumes("application/json").POST("/projects").To(u.createProject))
	ws.Route(ws.Consumes("application/json").GET("/projects").To(u.todo))
	ws.Route(ws.Consumes("application/json").GET("/projects/{id}").To(u.todo))
	ws.Route(ws.Consumes("application/json").PUT("/projects/{id}").To(u.todo))
	ws.Route(ws.Consumes("application/json").DELETE("/projects/{id}").To(u.todo))
	ws.Route(ws.Consumes("application/json").PATCH("/projects/{id}").To(u.todo))

	ws.Route(ws.POST("/dockerfiles").To(u.todo))

	ws.Route(ws.GET("/todo/{any}").To(u.todo).
		// docs
		Doc("to do").
		Operation("todo").
		Param(ws.QueryParameter("qs", "query string").DataType("string")).
		Param(ws.PathParameter("any", "service path").DataType("string")))

	if container != nil {
		container.Add(ws)
	} else {
		restful.Add(ws)
	}
}

func (u *UserResource) todo(request *restful.Request, response *restful.Response) {
	response.Write(blankEntity)
}

func Run() error {
	// to see what happens in the package, uncomment the following
	restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	//wsContainer := restful.NewContainer()
	ws := new(restful.WebService)
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

	ws.Route(ws.GET("/").To(todo("/api/v1")))

	Usrs.Register(ws, nil)
	//restful.Add(ws)

	// Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
	// You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
	// Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
	config := swagger.Config{
		WebServices:    restful.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath: "/apidocs/",
		//SwaggerFilePath: "/Users/emicklei/Projects/swagger-ui/dist",
		SwaggerFilePath: "/data/src/github.com/swagger-api/swagger-ui/dist",
	}
	//swagger.InstallSwaggerService(config)
	//swagger.RegisterSwaggerService(config, wsContainer)
	swagger.RegisterSwaggerService(config, restful.DefaultContainer)

	//server := &http.Server{Addr: ":8080", Handler: wsContainer}
	//log.Fatal(server.ListenAndServe())

	log.Printf("start listening on %s:8085", "")
	return http.ListenAndServe(":8085", nil)
}

//type handler func(request *restful.Request, response *restful.Response)

func todo(context string) restful.RouteFunction {
	switch context {
	case "/":
		return func(request *restful.Request, response *restful.Response) {
			response.Write([]byte(`
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
			response.Write([]byte(`
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
		response.WriteHeader(404)
	}
}
