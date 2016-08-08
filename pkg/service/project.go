package service

import (
	"bytes"
	"errors"
	"time"

	restful "github.com/emicklei/go-restful"
	//google_protobuf "github.com/golang/protobuf/ptypes/any"
	"github.com/helm/helm-classic/codec"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
)

func (u *UserResource) CreateProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectCreationRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {
	return nil, errNotImplemented
}

func (u *UserResource) RetrieveProjectIntoArbitrary(ctx context.Context,
	in *osopb3.ProjectRetrieveRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {
	if in.Name == "" {
		return nil, errors.New("Unexpected")
	}
	raw, obj, err := origin.RetrieveProject(in.Name)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 || obj == nil {
		return nil, errUnexpected
	}
	out := &osopb3.ProjectResponseDataArbitrary{
		Name:          obj.Name,
		Result:        string(obj.Status.Phase),
		Finalizers:    []string{},
		ResultingCode: osopb3.K8SNamespacePhase(osopb3.K8SNamespacePhase_value[string(obj.Status.Phase)]),
		//Project: &google_protobuf.Any{
		//	TypeUrl: "type.googleapis.com/github.com/openshift/origin/pkg/project/api/v1",
		//	Value:   raw,
		//},
	}
	for _, v := range obj.Spec.Finalizers {
		out.Finalizers = append(out.Finalizers, string(v))
	}
	var b *bytes.Buffer = &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(obj.TypeMeta); err == nil {
		o, err := codec.JSON.Decode(b.Bytes()).One()
		if err == nil {
			out.Datatype = new(osopb3.K8STypeMeta)
			_ = o.Object(out.Datatype)
		}
	}
	b.Reset()
	if err := codec.JSON.Encode(b).One(obj.ObjectMeta); err == nil {
		o, err := codec.JSON.Decode(b.Bytes()).One()
		if err == nil {
			out.Metadata = new(osopb3.K8SObjectMeta)
			_ = o.Object(out.Metadata)
		}
	}
	b.Reset()
	if err := codec.JSON.Encode(b).One(obj); err == nil {
		_, err := codec.JSON.Decode(b.Bytes()).One()
		if err == nil {
			out.Raw = &osopb3.RawData{
				ObjectName:  obj.Name,
				ObjectBytes: b.Bytes(),
			}
		}
	}
	return out, nil
}

func (u *UserResource) UpdateProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectUpdationRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {
	return nil, errNotImplemented
}

func (u *UserResource) DeleteProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectDeletionRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {
	return nil, errNotImplemented
}

func (u *UserResource) CreateOriginProject(ctx context.Context,
	req *osopb3.CreateOriginProjectRequest) (*osopb3.CreateOriginProjectResponse, error) {
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
