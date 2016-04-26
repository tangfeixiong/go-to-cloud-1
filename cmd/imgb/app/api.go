
package app

import (

    restful "github.com/emicklei/go-restful"
    
    "golang.org/x/net/context"

)

type ScopeCtxInterface interface {
    BasicAuthentication() (username, password string)
}

type user struct {
    Id                          string
    *basicAuth
}

type basicAuth struct { username, password string }
type x509Auth struct { certificateAuthority string}
type tlsAuth struct { clientCertificate, clientKey string }

type userResource struct {
	AppCtx                      AppCtxInterface
    // normally one would use DAO (data access object)
    users                       map[string]user
    user                        *user
    workspace                   string
    // io
    request                     *restful.Request
    response                    *restful.Response
    develop                     string
}

var (
    logger *log.Logger = log.New(os.Stdout, "[debug-k8s-broker] ",
        log.LstdFlags|log.Lshortfile)    

    blankEntity []byte = []byte("{}")
    //var errorEntity []byte = []byte(`{"description": "Something went wrong. Please contact support at http://support.example.com."}`)
)


func (u userResource) createImage(request *restful.Request, response *restful.Response) {
	// ctx is the Context for this handler. Calling cancel closes the
	// ctx.Done channel, which is the cancellation signal for requests
	// started by this handler.
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	timeout, err := time.ParseDuration(request.Request.FormValue("timeout"))
	if err == nil {
		// The request has a timeout, so create a context that is
		// canceled automatically when the timeout expires.
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel() // Cancel ctx as soon as handleSearch returns.

    u.develop = example.DevelopFromRequest(request.Request)
    ctx = example.NewContextR2(ctx, &u)
    
}

func (u userResource) Register() {

/*
 Services of building Dockerfile and image, ACI
*/
    ws = new(restful.WebService)
	ws.
		Path("/imgf/").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well
		
	ws.Route(ws.POST("/").To(u.todo))
	ws.Route(ws.GET("/").To(u.todo))
	ws.Route(ws.PUT("/").To(u.todo))
	ws.Route(ws.DELETE("/").To(u.todo))
	ws.Route(ws.PATCH("/").To(u.todo))

	ws.Route(ws.POST("/dockerfiles").To(u.createImage))
/*
	ws.Route(ws.GET("/{image-builder-id}").To(u.FindImage).
		// docs
		Doc("build images from docker file or App Container Image manifest").
		Operation("buildImage").
		Param(ws.QueryParameter("opts", "one of Dockerfile, DockerImage, ACIM, ACIA").DataType("string"))
		Param(ws.PathParameter("image-builder-id", "identifier of the build operation").DataType("string")))
*/

    restful.Add(ws)
}

func (u *userResource) todo(request *restful.Request, response *restful.Response) {
	response.Write(blankEntity)
}
