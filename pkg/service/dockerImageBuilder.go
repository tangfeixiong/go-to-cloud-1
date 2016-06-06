/*
 Copyright 2016, All rights reserved.
 
 Author <tangfx128@gmail.com>
*/

package servers

import (
    _ "fmt"
    "log"
    "net/http"
    "os"
    
    restful "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful/swagger"

)

type ApiServerInterface interface {
	Dispatch()
}

type apiServer struct {
    Port string
    //Ip   string
}

func NewApiServer() *ApiServer {
    s := apiServer{
        Port: "8080",
    }
	
	if port := os.Getenv("PORT"); port != "" {
	    s.Port = port
	} 
    
    return &s
}

func (s *ApiServer) Dispatch() {
	
}

func (s *ApiServer) Run() error {
    // to see what happens in the package, uncomment the following
	restful.TraceLogger(log.New(os.Stdout, "[restful] ", log.LstdFlags|log.Lshortfile))

	ctx := userResource { AppCtx: s }
	ctx.Register()
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
	
	log.Printf("Start Listening on port %s\n", s.Port)
	err := http.ListenAndServe(":" + s.Port, nil)
	
	//server := &http.Server{Addr: ":8080", Handler: wsContainer}
	//log.Fatal(server.ListenAndServe())
	    
	return err
}
