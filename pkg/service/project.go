package service

import (
	"bytes"
	"errors"
	"log"
	"os"
	"time"

	restful "github.com/emicklei/go-restful"
	//google_protobuf "github.com/golang/protobuf/ptypes/any"
	"github.com/helm/helm-classic/codec"

	//projectapi "github.com/openshift/origin/pkg/project/api"
	projectapi "github.com/openshift/origin/pkg/project/api/v1"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
)

func (u *UserResource) CreateProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectCreationRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {

	logger = log.New(os.Stdout, "[CreateProjectIntoArbitrary] ", log.LstdFlags|log.Lshortfile)

	if req.Name == "" {
		return nil, errors.New("Unexpected")
	}

	obj := new(projectapi.Project)
	obj.Kind = "Project"
	obj.APIVersion = projectapi.SchemeGroupVersion.Version
	obj.Name = req.Name
	obj.Labels = make(map[string]string)
	for k, v := range req.Labels {
		obj.Labels[k] = v
	}
	obj.Annotations = make(map[string]string)
	for k, v := range req.Annotations {
		obj.Annotations[k] = v
	}
	obj.Spec.Finalizers = []kapi.FinalizerName{projectapi.FinalizerOrigin, kapi.FinalizerKubernetes}
	for _, v := range req.Finalizers {
		obj.Spec.Finalizers = append(obj.Spec.Finalizers, kapi.FinalizerName(v))
	}
	//obj.Spec.Finalizers = append(obj.Spec.Finalizers, kapi.FinalizerName(origin.FinalizerVender))

	b := bytes.Buffer{}
	if err := codec.JSON.Encode(&b).One(obj); err != nil {
		return nil, err
	}
	data, _, err := origin.CreateProjectFromArbitray(b.Bytes())
	if err != nil {
		return nil, err
	}
	logger.Printf("Project created: %s\n", string(data))

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		return nil, err
	}
	result := new(projectapi.Project)
	if err := hco.Object(result); err != nil {
		return nil, err
	}
	resp := &osopb3.ProjectResponseDataArbitrary{
		Name:          result.Name,
		Result:        string(result.Status.Phase),
		Finalizers:    []string{},
		ResultingCode: osopb3.K8SNamespacePhase(osopb3.K8SNamespacePhase_value[string(result.Status.Phase)]),
	}

	for _, v := range result.Spec.Finalizers {
		resp.Finalizers = append(resp.Finalizers, string(v))
	}

	resp.Datatype = &result.TypeMeta
	resp.Metadata = &result.ObjectMeta
	gvk := result.APIVersion + "/" + result.Kind
	resp.Raw = &osopb3.RawData{
		ObjectName:  gvk,
		ObjectBytes: data,
	}
	return resp, nil
}

func (u *UserResource) RetrieveProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectRetrieveRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {

	logger = log.New(os.Stdout, "[RetrieveProjectIntoArbitrary] ", log.LstdFlags|log.Lshortfile)

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
	resp := &osopb3.ProjectResponseDataArbitrary{
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
		resp.Finalizers = append(resp.Finalizers, string(v))
	}

	resp.Datatype = &obj.TypeMeta
	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		return nil, err
	}
	resp.Metadata = &kapi.ObjectMeta{}
	if err := hco.Object(resp.Metadata); err != nil {
		return nil, err
	}
	gvk := obj.APIVersion + "/" + obj.Kind
	resp.Raw = &osopb3.RawData{
		ObjectName:  gvk,
		ObjectBytes: raw,
	}
	return resp, nil
}

func (u *UserResource) UpdateProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectUpdationRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {
	return nil, errNotImplemented
}

func (u *UserResource) DeleteProjectIntoArbitrary(ctx context.Context,
	req *osopb3.ProjectDeletionRequestData) (*osopb3.ProjectResponseDataArbitrary, error) {

	logger = log.New(os.Stdout, "[DeleteProjectIntoArbitrary] ", log.LstdFlags|log.Lshortfile)

	if req.Name == "" {
		return nil, errors.New("Unexpected")
	}

	if err := origin.DeleteProject(req.Name); err != nil {
		return nil, err
	}

	//raw, obj, err := origin.RetrieveProject(req.Name)
	//if err != nil {
	//	return nil, err
	//}

	resp := &osopb3.ProjectResponseDataArbitrary{
		Name: req.Name,
	}
	//if len(raw) > 0 || obj != nil {
	//	resp.Result = string(obj.Status.Phase)
	//	resp.ResultingCode = osopb3.K8SNamespacePhase(osopb3.K8SNamespacePhase_value[string(obj.Status.Phase)])
	//	return resp, errUnexpected
	//}
	//
	//logger.Printf("Project removed: %s\n", string(raw))
	return resp, nil
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
