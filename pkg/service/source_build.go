package service

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"
	buildapi "github.com/openshift/origin/pkg/build/api/v1"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/gnatsd"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/build-builder"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/dispatcher"
)

func gitSourceBuildSkeleton(req *osopb3.DockerBuildRequestData) (*buildapi.BuildConfig, *buildapi.Build, error) {
	if req.Configuration == nil || req.Configuration.CommonSpec == nil ||
		req.Configuration.CommonSpec.Source == nil ||
		req.Configuration.CommonSpec.Source.Git == nil ||
		len(req.Configuration.CommonSpec.Source.Git.Uri) == 0 {
		return nil, nil, errUnexpected
	}
	if req.Configuration.CommonSpec.Strategy == nil ||
		req.Configuration.CommonSpec.Strategy.SourceStrategy == nil ||
		req.Configuration.CommonSpec.Strategy.SourceStrategy.From == nil ||
		len(req.Configuration.CommonSpec.Strategy.SourceStrategy.From.Kind) == 0 ||
		len(req.Configuration.CommonSpec.Strategy.SourceStrategy.From.Name) == 0 ||
		req.Configuration.CommonSpec.Output == nil ||
		req.Configuration.CommonSpec.Output.To == nil ||
		len(req.Configuration.CommonSpec.Output.To.Kind) == 0 ||
		len(req.Configuration.CommonSpec.Output.To.Name) == 0 {
		return nil, nil, errUnexpected
	}
	opt := builder.CommonSpecTemplateOption{
		SimpleGitOption: builder.SimpleGitOption{
			GitURI:    req.Configuration.CommonSpec.Source.Git.Uri,
			GitRef:    "master",
			FromKind:  "DockerImage",
			FromName:  "",
			ForcePull: false,
			ToKind:    "DockerImage",
			ToName:    "",
		},
		ContextDir: "/",
	}
	if len(req.Configuration.CommonSpec.Source.Git.Ref) > 0 {
		opt.GitRef = req.Configuration.CommonSpec.Source.Git.Ref
	}
	if len(req.Configuration.CommonSpec.Source.ContextDir) > 0 {
		opt.ContextDir = req.Configuration.CommonSpec.Source.ContextDir
	}
	opt.FromKind = req.Configuration.CommonSpec.Strategy.SourceStrategy.From.Kind
	opt.FromName = req.Configuration.CommonSpec.Strategy.SourceStrategy.From.Name
	opt.ForcePull = req.Configuration.CommonSpec.Strategy.SourceStrategy.ForcePull
	opt.ToKind = req.Configuration.CommonSpec.Output.To.Kind
	opt.ToName = req.Configuration.CommonSpec.Output.To.Name

	bTmplOpt := builder.BuildTemplateOption{
		ObjectMetaTemplateOption: builder.ObjectMetaTemplateOption{
			Name:      req.Name,
			Namespace: req.ProjectName,
		},
	}
	bcTmplOpt := builder.BuildConfigTemplateOption{
		ObjectMetaTemplateOption: builder.ObjectMetaTemplateOption{
			Name:      req.Configuration.Name,
			Namespace: req.Configuration.ProjectName,
		},
	}
	bTmpl := builder.SourceBuildConfigTemplate["BuildForGitByManuallyTriggered"]
	bcTmpl := builder.SourceBuildConfigTemplate["BuildConfigForGitWithoutTriggers"]

	tmpl := template.New(req.Name)

	tmpl = template.Must(tmpl.Parse(bTmpl))
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, bTmplOpt); err != nil {
		glog.Errorf("Failed to excute template engine: %+v", err)
		return nil, nil, err
	}
	if glog.V(2) {
		glog.V(2).Infof("Received build object: %+v", buf.String())
	} else {
		glog.Infof("Received build object: %+v", buf.String())
	}

	hco, err := codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		glog.Errorf("Failed to docode into build object: %+v", err)
		return nil, nil, err
	}
	bld := new(buildapi.Build)
	if err := hco.Object(bld); err != nil {
		glog.Errorf("Failed to docode into build object: %+v", err)
		return nil, nil, err
	}

	tmpl = template.Must(tmpl.Parse(bcTmpl))
	buf.Reset()
	if err := tmpl.Execute(buf, bcTmplOpt); err != nil {
		glog.Errorf("Failed to excute template engine: %+v", err)
		return nil, nil, err
	}
	if glog.V(2) {
		glog.V(2).Infof("Received buildconfig object: %+v", buf.String())
	} else {
		glog.Infof("Received buildconfig object: %+v", buf.String())
	}

	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		glog.Errorf("Failed to docode into buildconfig object: %+v", err)
		return nil, nil, err
	}
	bc := new(buildapi.BuildConfig)
	if err := hco.Object(bld); err != nil {
		glog.Errorf("Failed to docode into buildconfig object: %+v", err)
		return nil, nil, err
	}
	return bc, bld, nil
}

func (u *UserResource) CreateStiBuilderIntoImage(ctx context.Context, req *osopb3.StiBuildRequestData) (*osopb3.StiBuildResponseData, error) {
	if len(req.Project) == 0 || len(req.BuildRequests) == 0 {
		glog.Errorln("Request body required")
		return &osopb3.StiBuildResponseData{}, errBadRequest
	}
	result := &osopb3.StiBuildResponseData{
		Authorized:     true,
		Project:        req.Project,
		BuildResponses: make([]*osopb3.DockerBuildResponseData, 0),
	}
	for i := range req.BuildRequests {
		req, err := u.createStiBuilderIntoImage(ctx, req.BuildRequests[i])
		if err != nil {
			return result, err
		}
		result.BuildResponses = append(result.BuildResponses, req)
	}
	return result, nil
}

func (u *UserResource) createStiBuilderIntoImage(ctx context.Context, req *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	var raw []byte
	var obj *buildapi.Build
	var bc *buildapi.BuildConfig
	var err error

	if len(req.Name) == 0 || len(req.ProjectName) == 0 ||
		req.Configuration == nil || len(req.Configuration.Name) == 0 ||
		len(req.Configuration.ProjectName) == 0 ||
		req.Configuration.CommonSpec == nil ||
		req.Configuration.CommonSpec.Source == nil {
		return &osopb3.DockerBuildResponseData{}, fmt.Errorf("Bad request")
	}

	bc, obj, err = gitSourceBuildSkeleton(req)
	if err != nil {
		if err != errUnexpected {
			return nil, errBadRequest
		}
		return nil, errUnexpected
	}

	op := new(origin.PaaS).WithOCctl("", "", "", "").WithEtcdCtl([]string{}, 0, 0)
	err = op.VerifyProject(bc.Namespace)
	if err != nil {
		glog.Errorf("Failed to create origin project (%+v)\n", bc)
		return &osopb3.DockerBuildResponseData{}, err
	}
	if strings.Compare(bc.Namespace, obj.Namespace) != 0 {
		err = op.VerifyProject(obj.Namespace)
		if err != nil {
			glog.Errorf("Failed to create origin project (%+v)\n", bc)
			return &osopb3.DockerBuildResponseData{}, err
		}
	}

	if err = verifyDockerSecret(req, obj); err != nil {
		return &osopb3.DockerBuildResponseData{}, err
	}

	raw, obj, bc, err = op.CreateNewBuild(obj, bc)
	if err != nil {
		glog.Errorf("Failed to docker build with config (%+v)\n", bc)
		return &osopb3.DockerBuildResponseData{}, err
	}
	if len(raw) == 0 || obj == nil {
		glog.Errorf("Nothing received from docker build with config (%+v)", bc)
		return &osopb3.DockerBuildResponseData{}, nil
	}

	return u.scheduleStiBuildTracker(ctx, req, op, raw, obj, bc), nil
}

func (u *UserResource) scheduleStiBuildTracker(ctx context.Context,
	req *osopb3.DockerBuildRequestData,
	op *origin.PaaS, raw []byte, obj *buildapi.Build, bc *buildapi.BuildConfig) (resp *osopb3.DockerBuildResponseData) {
	cmd, o := origin.NewCmdStartBuild("osoc", op.Factory(), os.Stdin, os.Stdout)
	o.In = os.Stdin
	o.Out = os.Stdout
	o.ErrOut = cmd.Out()
	o.StartBuildOptions.WaitForComplete = true
	o.StartBuildOptions.Follow = true
	o.StartBuildOptions.Namespace = obj.Namespace
	o.StartBuildOptions.Client = op.OC()
	resp = origin.GenerateResponseData(raw, obj)
	u.Schedulers["DockerBuilder"].WithPaylodHandler(
		func() dispatcher.HandleFunc {
			glog.Errorf("Schedule docker builder into tracker: %s/%s(%s)\n", obj.Namespace, obj.Name, bc.Name)
			return o.TrackWith(ctx, req, resp, op, raw, obj, bc)
		}(),
	)
	return
}

func (u *UserResource) TrackStiBuild(ctx context.Context, req *osopb3.StiBuildRequestData) (*osopb3.StiBuildResponseData, error) {
	data := &osopb3.StiBuildResponseData{
		Authorized:     true,
		Project:        req.Project,
		BuildResponses: make([]*osopb3.DockerBuildResponseData, 0),
	}
	if len(req.Project) == 0 || len(req.BuildRequests) == 0 {
		glog.Errorln("Request body required")
		return data, errBadRequest
	}

	op := new(origin.PaaS).WithOCctl("", "", "", "").WithEtcdCtl([]string{}, 0, 0)
	for i := range req.BuildRequests {
		resp := new(osopb3.DockerBuildResponseData)
		if etcdctl := op.EtcdCtl(); etcdctl != nil {
			prefix := origin.EtcdV3BuildCacheKey("v1", "default",
				req.BuildRequests[i].ProjectName, req.BuildRequests[i].Configuration.Name, req.BuildRequests[i].Name, false)
			result, err := etcdctl.GetWithPrefix(prefix)
			if err != nil {
				return data, err
			}
			if result == nil || len(result.Kvs) == 0 {
				return data, fmt.Errorf("Unexpected as nothing to result")
			}
			for _, val := range result.Kvs {
				if strings.Compare(prefix, string(val.Key)) == 0 {
					if err := resp.Unmarshal(val.Value); err != nil {
						glog.Errorf("Could not unmarshal into response: %+v", err)
						return data, err
					}
				} else {
					resp.Status.Message = fmt.Sprintf("%s\n%s", resp.Status.Message, val.Value)
				}
			}
		} else {
			b, err := gnatsd.Subscribe([]string{}, nil, nil, origin.Subject(req.BuildRequests[i].ProjectName, req.BuildRequests[i].Name))
			if err != nil {
				return data, err
			}
			if err := resp.Unmarshal(b); err != nil {
				glog.Errorf("Could not unmarshal into response: %+v", err)
				return data, err
			}
		}
		data.BuildResponses = append(data.BuildResponses, resp)
	}
	return data, nil
}
