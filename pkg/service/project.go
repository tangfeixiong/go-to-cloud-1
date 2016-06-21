package service

import (
	"time"

	restful "github.com/emicklei/go-restful"

	"golang.org/x/net/context"

	_ "google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/api/paas/ci/openshift"
)

func (u *UserResource) createProject(request *restful.Request, response *restful.Response) {
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

	// according kapi.Context, namespaceKey=0, userKey=1
	context.WithValue(ctx, 1, &u)
}

func (u *UserResource) CreateProject(context.Context, *openshift.CreateProjectRequest) (*openshift.CreateProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) LookupProjects(context.Context, *openshift.LookupProjectsRequest) (*openshift.LookupProjectsResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) OpenProject(context.Context, *openshift.OpenProjectRequest) (*openshift.OpenProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) DeleteProject(ctx context.Context, in *openshift.DeleteProjectRequest) (*openshift.DeleteProjectResponse, error) {
	return nil, errNotImplemented
}
