package service

import (
	//"bytes"
	//"log"
	"os"

	//"github.com/helm/helm-classic/codec"
	buildapi "github.com/openshift/origin/pkg/build/api/v1"

	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	//kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/gnatsd"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/dispatcher"
)

var (
	_builderServiceAccount string = "builder"
	_dockerfile            string = "FROM busybox\nCMD [\"sh\"]"
	_timeout               int64  = 900
)

func toOriginBuildOutputType(t string) string {
	return t[len("output_"):]
}

func toOriginBuildStrategyType(t string) string {
	return t[len("strategy_"):]
}

func convertIntoBuildObject(req *osopb3.DockerBuildRequestData) (*buildapi.BuildConfig, *buildapi.Build) {
	common := buildapi.CommonSpec{
		ServiceAccount: _builderServiceAccount,
		Source: buildapi.BuildSource{
			Type:       buildapi.BuildSourceNone,
			Binary:     (*buildapi.BinaryBuildSource)(nil),
			Dockerfile: &_dockerfile,
			Git:        (*buildapi.GitBuildSource)(nil),
			Images:     []buildapi.ImageSource{
			/*buildapi.ImageSource{
				From: kapi.ObjectReference{
					Kind: "DockerImage",
					Name: "alpine:edge",
				},
				Paths: []buildapi.ImageSourcePath{
					{
						SourcePath:     "",
						DestinationDir: "",
					},
				},
				PullSecret: &kapi.LocalObjectReference{},
			},*/
			},
			ContextDir:   "",
			SourceSecret: (*kapi.LocalObjectReference)(nil),
			Secrets:      []buildapi.SecretBuildSource{
			/*{
				Secret:         &kapi.LocalObjectReference{},
				DestinationDir: "/root/.docker/config.json",
			},*/
			},
		},
		Revision: (*buildapi.SourceRevision)(nil),
		Strategy: buildapi.BuildStrategy{
			Type: buildapi.DockerBuildStrategyType,
			DockerStrategy: &buildapi.DockerBuildStrategy{
				From: &kapi.ObjectReference{
					Kind: "DockerImage",
					Name: "busybox:latest",
				},
				PullSecret:     (*kapi.LocalObjectReference)(nil),
				NoCache:        true,
				Env:            []kapi.EnvVar{},
				ForcePull:      false,
				DockerfilePath: ".",
			},
			SourceStrategy:          (*buildapi.SourceBuildStrategy)(nil),
			CustomStrategy:          (*buildapi.CustomBuildStrategy)(nil),
			JenkinsPipelineStrategy: (*buildapi.JenkinsPipelineBuildStrategy)(nil),
		},
		Output: buildapi.BuildOutput{
			To:         (*kapi.ObjectReference)(nil),
			PushSecret: (*kapi.LocalObjectReference)(nil),
		},
		Resources: kapi.ResourceRequirements{},
		PostCommit: buildapi.BuildPostCommitSpec{
			Command: []string{},
			Args:    []string{},
			Script:  "",
		},
		CompletionDeadlineSeconds: &_timeout,
	}

	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Source != nil &&
		req.Configuration.CommonSpec.Source.Git != nil &&
		req.Configuration.CommonSpec.Source.Git.Uri != "" {
		common.Source.Type = buildapi.BuildSourceGit
		common.Source.Git = &buildapi.GitBuildSource{
			URI:        req.Configuration.CommonSpec.Source.Git.Uri,
			Ref:        req.Configuration.CommonSpec.Source.Git.Ref,
			HTTPProxy:  nil,
			HTTPSProxy: nil,
		}
		common.Source.ContextDir = req.Configuration.CommonSpec.Source.ContextDir
		common.Source.SourceSecret = req.Configuration.CommonSpec.Source.SourceSecret
		if req.Configuration.CommonSpec.Source.Secrets != nil {
			for _, val := range req.Configuration.CommonSpec.Source.Secrets {
				if val != nil && val.Secret != nil {
					common.Source.Secrets = append(common.Source.Secrets,
						buildapi.SecretBuildSource{*val.Secret, val.DestinationDir})
				}
			}
		}
	}

	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Source != nil &&
		len(req.Configuration.CommonSpec.Source.Images) > 0 {
		for _, img := range req.Configuration.CommonSpec.Source.Images {
			if img != nil && img.From != nil && len(img.Paths) > 0 {
				var paths []buildapi.ImageSourcePath
				for _, val := range img.Paths {
					if val != nil && val.SourcePath != "" && val.DestinationDir != "" {
						paths = append(paths,
							buildapi.ImageSourcePath{
								SourcePath:     val.SourcePath,
								DestinationDir: val.DestinationDir,
							})
					}
				}
				if len(paths) > 0 {
					ele := buildapi.ImageSource{
						From:       *img.From,
						Paths:      paths,
						PullSecret: img.PullSecret,
					}
					common.Source.Images = append(common.Source.Images, ele)
				}
			}
		}
		common.Source.SourceSecret = req.Configuration.CommonSpec.Source.SourceSecret
		if req.Configuration.CommonSpec.Source.Secrets != nil {
			for _, val := range req.Configuration.CommonSpec.Source.Secrets {
				if val.Secret != nil {
					common.Source.Secrets = append(common.Source.Secrets,
						buildapi.SecretBuildSource{*val.Secret, val.DestinationDir})
				}
			}
		}
		if len(common.Source.Images) > 0 &&
			common.Source.Type == buildapi.BuildSourceNone {
			common.Source.Type = buildapi.BuildSourceImage
		}
	}

	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Source != nil &&
		req.Configuration.CommonSpec.Source.Dockerfile != "" {
		common.Source.Dockerfile = &req.Configuration.CommonSpec.Source.Dockerfile
		if common.Source.Type == buildapi.BuildSourceNone {
			common.Source.Type = buildapi.BuildSourceDockerfile
		}
	}

	//revision
	/*if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Revision != nil &&
		req.Configuration.CommonSpec.Revision.Git != nil {
		common.Revision = &buildapi.SourceRevision{
			Type: common.Source.Type,
			Git: &buildapi.GitSourceRevision{
				Commit:  req.Configuration.CommonSpec.Revision.Git.Commit,
				Message: req.Configuration.CommonSpec.Revision.Git.Message,
			},
		}
		if req.Configuration.CommonSpec.Revision.Git.Author != nil &&
			req.Configuration.CommonSpec.Revision.Git.Author.Name != "" &&
			req.Configuration.CommonSpec.Revision.Git.Author.Email != "" {
			common.Revision.Git.Author = buildapi.SourceControlUser{
				req.Configuration.CommonSpec.Revision.Git.Author.Name,
				req.Configuration.CommonSpec.Revision.Git.Author.Email,
			}
		}
		if req.Configuration.CommonSpec.Revision.Git.Committer != nil &&
			req.Configuration.CommonSpec.Revision.Git.Committer.Name != "" &&
			req.Configuration.CommonSpec.Revision.Git.Committer.Email != "" {
			common.Revision.Git.Committer = buildapi.SourceControlUser{
				req.Configuration.CommonSpec.Revision.Git.Committer.Name,
				req.Configuration.CommonSpec.Revision.Git.Committer.Email,
			}
		}
	}*/

	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Strategy != nil {
		switch {
		case req.Configuration.CommonSpec.Strategy.DockerStrategy != nil:
			common.Strategy.Type = buildapi.DockerBuildStrategyType
			common.Strategy.DockerStrategy = &buildapi.DockerBuildStrategy{
				From:           req.Configuration.CommonSpec.Strategy.DockerStrategy.From,
				PullSecret:     req.Configuration.CommonSpec.Strategy.DockerStrategy.PullSecret,
				NoCache:        req.Configuration.CommonSpec.Strategy.DockerStrategy.NoCache,
				Env:            []kapi.EnvVar{},
				ForcePull:      req.Configuration.CommonSpec.Strategy.DockerStrategy.ForcePull,
				DockerfilePath: req.Configuration.CommonSpec.Strategy.DockerStrategy.DockerfilePath,
			}
			for _, val := range req.Configuration.CommonSpec.Strategy.DockerStrategy.Env {
				if val != nil {
					common.Strategy.DockerStrategy.Env = append(common.Strategy.DockerStrategy.Env,
						*val)
				}
			}
		case req.Configuration.CommonSpec.Strategy.SourceStrategy != nil:
			common.Strategy.Type = buildapi.SourceBuildStrategyType
			common.Strategy.SourceStrategy = &buildapi.SourceBuildStrategy{
				From:        *req.Configuration.CommonSpec.Strategy.SourceStrategy.From,
				PullSecret:  req.Configuration.CommonSpec.Strategy.SourceStrategy.PullSecret,
				Env:         []kapi.EnvVar{},
				Scripts:     req.Configuration.CommonSpec.Strategy.SourceStrategy.Scripts,
				Incremental: req.Configuration.CommonSpec.Strategy.SourceStrategy.Incremental,
				ForcePull:   req.Configuration.CommonSpec.Strategy.SourceStrategy.ForcePull,
			}
			for _, val := range req.Configuration.CommonSpec.Strategy.SourceStrategy.Env {
				if val != nil {
					common.Strategy.SourceStrategy.Env = append(common.Strategy.SourceStrategy.Env,
						*val)
				}
			}
		case req.Configuration.CommonSpec.Strategy.JenkinsPipelineStrategy != nil:
			common.Strategy.Type = buildapi.JenkinsPipelineBuildStrategyType
			common.Strategy.JenkinsPipelineStrategy = &buildapi.JenkinsPipelineBuildStrategy{
				JenkinsfilePath: req.Configuration.CommonSpec.Strategy.JenkinsPipelineStrategy.JenkinsfilePath,
				Jenkinsfile:     req.Configuration.CommonSpec.Strategy.JenkinsPipelineStrategy.Jenkinsfile,
			}
		default:
			common.Strategy.Type = buildapi.CustomBuildStrategyType
			common.Strategy.CustomStrategy = &buildapi.CustomBuildStrategy{
				From:               *req.Configuration.CommonSpec.Strategy.CustomStrategy.From,
				PullSecret:         req.Configuration.CommonSpec.Strategy.CustomStrategy.PullSecret,
				Env:                []kapi.EnvVar{},
				ExposeDockerSocket: req.Configuration.CommonSpec.Strategy.CustomStrategy.ExposeDockerSocket,
				ForcePull:          req.Configuration.CommonSpec.Strategy.CustomStrategy.ForcePull,
				Secrets:            []buildapi.SecretSpec{},
				BuildAPIVersion:    req.Configuration.CommonSpec.Strategy.CustomStrategy.BuildAPIVersion,
			}
			for _, val := range req.Configuration.CommonSpec.Strategy.CustomStrategy.Env {
				if val != nil {
					common.Strategy.CustomStrategy.Env = append(common.Strategy.CustomStrategy.Env,
						*val)
				}
			}
			for _, val := range req.Configuration.CommonSpec.Strategy.CustomStrategy.Secrets {
				if val != nil {
					common.Strategy.CustomStrategy.Secrets = append(common.Strategy.CustomStrategy.Secrets,
						buildapi.SecretSpec{*val.SecretSource, val.MountPath})
				}
			}
		}
	}

	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Output != nil {
		common.Output = buildapi.BuildOutput{
			To:         req.Configuration.CommonSpec.Output.To,
			PushSecret: req.Configuration.CommonSpec.Output.PushSecret,
		}
	}
	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.Resources != nil {
		common.Resources = *req.Configuration.CommonSpec.Resources
	}
	//PostCommit
	if req.Configuration != nil && req.Configuration.CommonSpec != nil &&
		req.Configuration.CommonSpec.PostCommit != nil {
		common.PostCommit.Command = req.Configuration.CommonSpec.PostCommit.Command
		common.PostCommit.Args = req.Configuration.CommonSpec.PostCommit.Args
		common.PostCommit.Script = req.Configuration.CommonSpec.PostCommit.Script
	}

	bld := &buildapi.Build{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Build",
			APIVersion: buildapi.SchemeGroupVersion.Version,
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:              req.Name,
			Namespace:         req.ProjectName,
			CreationTimestamp: unversioned.Now(),
			Labels:            req.Labels,
			Annotations:       req.Annotations,
		},
		Spec: buildapi.BuildSpec{
			CommonSpec:  common,
			TriggeredBy: []buildapi.BuildTriggerCause{},
		},
		Status: buildapi.BuildStatus{
			Phase: buildapi.BuildPhaseNew,
		},
	}
	//TriggeredBy
	/*for _, val := range req.TriggeredBy {
		if val != nil {
			cause := buildapi.BuildTriggerCause{
				Message: val.Message,
			}
			switch {
			case val.GenericWebHook != nil:
				cause.GenericWebHook = &buildapi.GenericWebHookCause{
					Secret: val.GenericWebHook.Secret,
				}
				if val.GenericWebHook.Revision != nil {
					cause.GenericWebHook.Revision = &buildapi.SourceRevision{
						Type: buildapi.BuildSourceType(val.GenericWebHook.Revision.Type),
					}
					if val.GenericWebHook.Revision.Git != nil {
						cause.GenericWebHook.Revision.Git = &buildapi.GitSourceRevision{
							Commit:  val.GenericWebHook.Revision.Git.Commit,
							Message: val.GenericWebHook.Revision.Git.Message,
						}
						if val.GenericWebHook.Revision.Git.Author != nil &&
							val.GenericWebHook.Revision.Git.Author.Name != "" &&
							val.GenericWebHook.Revision.Git.Author.Email != "" {
							cause.GenericWebHook.Revision.Git.Author = buildapi.SourceControlUser{
								val.GenericWebHook.Revision.Git.Author.Name,
								val.GenericWebHook.Revision.Git.Author.Email,
							}
						}
						if val.GenericWebHook.Revision.Git.Committer != nil &&
							val.GenericWebHook.Revision.Git.Committer.Name != "" &&
							val.GenericWebHook.Revision.Git.Committer.Email != "" {
							cause.GenericWebHook.Revision.Git.Committer = buildapi.SourceControlUser{
								val.GenericWebHook.Revision.Git.Committer.Name,
								val.GenericWebHook.Revision.Git.Committer.Email,
							}
						}
					}
				}
			case val.GithubWebHook != nil:
				cause.GitHubWebHook = &buildapi.GitHubWebHookCause{
					Secret: val.GithubWebHook.Secret,
				}
				if val.GithubWebHook.Revision != nil {
					cause.GitHubWebHook.Revision = &buildapi.SourceRevision{
						Type: buildapi.BuildSourceType(val.GithubWebHook.Revision.Type),
					}
					if val.GithubWebHook.Revision.Git != nil {
						cause.GitHubWebHook.Revision.Git = &buildapi.GitSourceRevision{
							Commit:  val.GithubWebHook.Revision.Git.Commit,
							Message: val.GithubWebHook.Revision.Git.Message,
						}
						if val.GithubWebHook.Revision.Git.Author != nil &&
							val.GithubWebHook.Revision.Git.Author.Name != "" &&
							val.GithubWebHook.Revision.Git.Author.Email != "" {
							cause.GitHubWebHook.Revision.Git.Author = buildapi.SourceControlUser{
								val.GithubWebHook.Revision.Git.Author.Name,
								val.GithubWebHook.Revision.Git.Author.Email,
							}
						}
						if val.GithubWebHook.Revision.Git.Committer != nil &&
							val.GithubWebHook.Revision.Git.Committer.Name != "" &&
							val.GithubWebHook.Revision.Git.Committer.Email != "" {
							cause.GitHubWebHook.Revision.Git.Committer = buildapi.SourceControlUser{
								val.GithubWebHook.Revision.Git.Committer.Name,
								val.GithubWebHook.Revision.Git.Committer.Email,
							}
						}
					}
				}
			case val.ImageChangeBuild != nil:
				cause.ImageChangeBuild = &buildapi.ImageChangeCause{
					ImageID: val.ImageChangeBuild.ImageID,
					FromRef: val.ImageChangeBuild.FromRef,
				}
			}
			bld.Spec.TriggeredBy = append(bld.Spec.TriggeredBy, cause)
		}
	}*/

	bldconf := &buildapi.BuildConfig{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "BuildConfig",
			APIVersion: buildapi.SchemeGroupVersion.Version,
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:              req.Name,
			Namespace:         req.ProjectName,
			CreationTimestamp: unversioned.Now(),
			Labels:            req.Labels,
			Annotations:       req.Annotations,
		},
		Spec: buildapi.BuildConfigSpec{
			Triggers:   []buildapi.BuildTriggerPolicy{},
			RunPolicy:  buildapi.BuildRunPolicySerial,
			CommonSpec: common,
		},
	}
	if req.Configuration != nil {
		bldconf.Spec.RunPolicy = buildapi.BuildRunPolicy(req.Configuration.RunPolicy)
		for _, val := range req.Configuration.Triggers {
			if val != nil {
				switch {
				case val.GithubWebHook != nil:
					ele := buildapi.BuildTriggerPolicy{
						Type:          buildapi.GitHubWebHookBuildTriggerType,
						GitHubWebHook: &buildapi.WebHookTrigger{val.GithubWebHook.Secret, val.GithubWebHook.AllowEnv}}
					bldconf.Spec.Triggers = append(bldconf.Spec.Triggers, ele)
				case val.GenericWebHook != nil:
					ele := buildapi.BuildTriggerPolicy{
						Type:           buildapi.GenericWebHookBuildTriggerType,
						GenericWebHook: &buildapi.WebHookTrigger{val.GenericWebHook.Secret, val.GenericWebHook.AllowEnv}}
					bldconf.Spec.Triggers = append(bldconf.Spec.Triggers, ele)
				case val.ImageChange != nil:
					ele := buildapi.BuildTriggerPolicy{
						Type:        buildapi.ImageChangeBuildTriggerType,
						ImageChange: &buildapi.ImageChangeTrigger{val.ImageChange.LastTriggeredImageID, val.ImageChange.From}}
					bldconf.Spec.Triggers = append(bldconf.Spec.Triggers, ele)
				}
			}
		}
	}
	return bldconf, bld
}

func (u *UserResource) CreateIntoBuildDockerImage(ctx context.Context,
	req *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[service, .CreateIntoBuildDockerImage] ")

	var raw []byte
	var obj *buildapi.Build
	var bc *buildapi.BuildConfig
	var err error

	bc, obj = convertIntoBuildObject(req)

	op := new(origin.PaaS)
	raw, obj, bc, err = op.CreateNewBuild(obj, bc)
	//raw, obj, err = origin.DirectlyRunOriginDockerBuilder(obj)
	if err != nil {
		logger.Printf("Failed to docker build with config (%+v)\n", bc)
		return &osopb3.DockerBuildResponseData{}, err
	}
	if len(raw) == 0 || obj == nil {
		logger.Printf("Nothing received from docker build with config (%+v)", bc)
		return &osopb3.DockerBuildResponseData{}, nil
	}

	//return origin.GenerateResponseData(raw, obj), nil
	return u.trackCreatingIntoBuildDockerImage(ctx, req, op, raw, obj, bc), nil
}

func (u *UserResource) trackCreatingIntoBuildDockerImage(ctx context.Context,
	req *osopb3.DockerBuildRequestData,
	op *origin.PaaS, raw []byte, obj *buildapi.Build, bc *buildapi.BuildConfig) (resp *osopb3.DockerBuildResponseData) {
	logger.SetPrefix("[service, .trackCreatingIntoBuildDockerImage] ")
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
			logger.Printf("Schedule docker builder into tracker: %s/%s(%s)\n", obj.Namespace, obj.Name, bc.Name)
			return o.TrackWith(ctx, req, resp, op, raw, obj, bc)
		}(),
	)
	return
}

func (u *UserResource) RetrieveIntoBuildDockerImage(ctx context.Context,
	req *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	logger.SetPrefix("[service, .RetrieveIntoBuildDockerImage] ")

	if req.Name == "" {
		logger.Println("Request body required")
		return (*osopb3.DockerBuildResponseData)(nil), errUnexpected
	}
	b, err := gnatsd.Subscribe([]string{}, nil, nil, origin.Subject(req.ProjectName, req.Name))
	if err != nil {
		return (*osopb3.DockerBuildResponseData)(nil), err
	}
	resp := new(osopb3.DockerBuildResponseData)
	if err := resp.Unmarshal(b); err != nil {
		logger.Printf("Could not unmarshal into response: %+v", err)
		return resp, err
	}

	/*raw, obj, err := origin.RetrieveBuild(req.ProjectName, req.Name)
	if err != nil {
		return (*osopb3.DockerBuildResponseData)(nil), err
	}
	if raw == nil || len(raw) == 0 || obj == nil {
		logger.Println("No data and object retrieved")
		return &osopb3.DockerBuildResponseData{nil, nil}, nil
	}

	status := &osopb3.OsoBuildStatus{
		Phase:                      string(obj.Status.Phase),
		Cancelled:                  obj.Status.Cancelled,
		Reason:                     string(obj.Status.Reason),
		StartTimestamp:             obj.Status.StartTimestamp,
		CompletionTimestamp:        obj.Status.CompletionTimestamp,
		Duration:                   int64(obj.Status.Duration),
		OutputDockerImageReference: obj.Status.OutputDockerImageReference,
		Config: obj.Status.Config,
	}
	status.OsoBuildPhase = osopb3.OsoBuildStatus_OsoBuildPhase(osopb3.OsoBuildStatus_OsoBuildPhase_value[status.Phase])

	gvk := unversioned.GroupVersionKind{
		Group:   "",
		Version: obj.TypeMeta.APIVersion,
		Kind:    obj.TypeMeta.Kind,
	}.String()

	resp := &osopb3.DockerBuildResponseData{
		Status: status,
		Raw: &osopb3.RawJSON{
			ObjectGVK:  gvk,
			ObjectJSON: raw,
		},
	}*/

	return resp, nil
}

func (u *UserResource) UpdateIntoBuildDockerImage(ctx context.Context,
	in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) DeleteIntoBuildDockerImage(ctx context.Context,
	in *osopb3.DockerBuildRequestData) (*osopb3.DockerBuildResponseData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) NewOsoBuildConfig(ctx context.Context,
	in *buildapi.BuildConfig) (*buildapi.BuildConfig, error) {
	return nil, errNotImplemented
}

func (u *UserResource) StartOsoBuild(ctx context.Context,
	in *buildapi.Build) (*buildapi.Build, error) {
	return nil, errNotImplemented
}

func (u *UserResource) BuildDockerImage(ctx context.Context,
	in *osopb3.RawData) (*osopb3.RawData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) RebuildDockerImage(ctx context.Context,
	in *osopb3.RawData) (*osopb3.RawData, error) {
	return nil, errNotImplemented
}
