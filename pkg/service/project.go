package service

import (
	"errors"
	"time"

	restful "github.com/emicklei/go-restful"

	//google_protobuf "github.com/golang/protobuf/ptypes/any"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
)

func (u *UserResource) CreateOriginProject(ctx context.Context,
	req *osopb3.CreateOriginProjectRequest) (*osopb3.CreateOriginProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) CreateOriginProjectArbitrary(ctx context.Context,
	req *osopb3.CreateOriginProjectArbitraryRequest) (*osopb3.CreateOriginProjectArbitraryResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) FindProject(ctx context.Context,
	req *osopb3.FindProjectRequest) (*osopb3.FindProjectResponse, error) {
	if req.Name == "" {
		return nil, errors.New("Unexpected")
	}
	raw, obj, err := origin.RetrieveProject(req.Name)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || obj == nil {
		return nil, errUnexpected
	}
	resp := &osopb3.FindProjectResponse{
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

func (u *UserResource) CreateProject(ctx context.Context,
	req *osopb3.CreateOriginProjectRequest) (*osopb3.CreateOriginProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) OpenProject(ctx context.Context,
	req *osopb3.FindProjectRequest) (*osopb3.FindProjectResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) DeleteProject(ctx context.Context,
	req *osopb3.DeleteProjectRequest) (*osopb3.DeleteProjectResponse, error) {
	return nil, errNotImplemented
}
