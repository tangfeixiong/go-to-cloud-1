package service

import (
	"fmt"
	"log"
	"net/http"
	"os"

	restful "github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
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
