package service

import (
	"errors"
	"time"

	restful "github.com/emicklei/go-restful"

	//google_protobuf "github.com/golang/protobuf/ptypes/any"

	"golang.org/x/net/context"

	_ "google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/openshift/client"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/paas/ci/openshift"
)

func (u *UserResource) CreateOriginProject(context.Context, *openshift.CreateOriginProjectRequest) (*openshift.CreateOriginProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) CreateOriginProjectArbitrary(context.Context, *openshift.CreateOriginProjectArbitraryRequest) (*openshift.CreateOriginProjectArbitraryResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) FindProject(ctx context.Context, req *openshift.FindProjectRequest) (*openshift.FindProjectResponse, error) {
	if req.Name == "" {
		return nil, errors.New("Unexpected")
	}
	raw, obj, err := client.RetrieveProject(req.Name)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || obj == nil {
		return nil, errUnexpected
	}
	resp := &openshift.FindProjectResponse{
		Odefv1RawData: raw,
		//Project: &google_protobuf.Any{
		//	TypeUrl: "type.googleapis.com/github.com/openshift/origin/pkg/project/api/v1",
		//	Value:   raw,
		//},
	}
	return resp, nil
}

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
