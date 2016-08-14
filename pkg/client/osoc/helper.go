package osoc

import (
	"k8s.io/kubernetes/pkg/api/resource"
	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

var (
	_oso_builder        string = "nchellohttp"
	_oso_project        string = "default"
	_oso_ServiceAccount string = "builder"
	_oso_Dockerfile     string = "FROM alpine:3.4\nRUN apk add --update bash netcat-openbsd && rm -rf /var/cache/apk/*\nRUN echo \"<html><body><h1>hello world</h1></body></html>\" >> /tmp/index.html\nEXPOSE 80\nCMD [\"nc\", \"-l\", \"80\", \"</tmp/index.html\"]"
	_oso_dockerPush     string = "172.17.4.50:30005/tangfx/nchellohttp:latest"
	_oso_GitURI         string = "http://172.17.4.50:30080/tangfx/netcat-alpine"
	_oso_timeout        int64  = 900
)

func internalDockerBuildRequestData() *osopb3.DockerBuildRequestData {
	return &osopb3.DockerBuildRequestData{
		Name:        _oso_builder,
		ProjectName: _oso_project,
		Configuration: &osopb3.DockerBuildConfigRequestData{
			Name:        _oso_builder,
			ProjectName: _oso_project,
			Triggers:    []*osopb3.OsoBuildTriggerPolicy{},
			RunPolicy:   osopb3.DockerBuildConfigRequestData_Serial.String(),
			CommonSpec: &osopb3.OsoCommonSpec{
				ServiceAccount: _oso_ServiceAccount,
				Source: &osopb3.BuildSource{
					Type:       osopb3.OsoBuildSourceType_Dockerfile.String(),
					Binary:     (*osopb3.BinaryBuildSource)(nil),
					Dockerfile: _oso_Dockerfile,
					Git: &osopb3.GitBuildSource{
						Uri:        _oso_GitURI,
						Ref:        "master",
						HttpProxy:  "",
						HttpsProxy: "",
					},
					Images:             []*osopb3.ImageSource{},
					ContextDir:         "",
					SourceSecret:       (*kapi.LocalObjectReference)(nil),
					Secrets:            []*osopb3.SecretBuildSource{},
					OsoBuildSourceType: osopb3.OsoBuildSourceType_Dockerfile,
				},
				Revision: &osopb3.SourceRevision{
					Type:            osopb3.OsoBuildSourceType_Dockerfile.String(),
					Git:             (*osopb3.GitSourceRevision)(nil),
					BuildSourceType: osopb3.OsoBuildSourceType_Dockerfile,
				},
				Strategy: &osopb3.BuildStrategy{
					Type: osopb3.BuildStrategy_Docker.String(),
					DockerStrategy: &osopb3.DockerBuildStrategy{
						From:           (*kapi.ObjectReference)(nil),
						PullSecret:     (*kapi.LocalObjectReference)(nil),
						NoCache:        true,
						Env:            []*kapi.EnvVar{},
						ForcePull:      false,
						DockerfilePath: ".",
					},
					SourceStrategy:          (*osopb3.SourceBuildStrategy)(nil),
					CustomStrategy:          (*osopb3.CustomBuildStrategy)(nil),
					JenkinsPipelineStrategy: (*osopb3.JenkinsPipelineBuildStrategy)(nil),
					OsoBuildStrategyType:    osopb3.BuildStrategy_Docker,
				},
				Output: &osopb3.BuildOutput{
					To: &kapi.ObjectReference{
						Kind: "DockerImage",
						Name: _oso_dockerPush,
					},
					PushSecret: &kapi.LocalObjectReference{
						Name: `localdockerconfig`,
					},
				},
				Resources: &kapi.ResourceRequirements{
					Limits:   kapi.ResourceList(map[kapi.ResourceName]resource.Quantity{}),
					Requests: kapi.ResourceList(map[kapi.ResourceName]resource.Quantity{}),
				},
				PostCommit: &osopb3.BuildPostCommitSpec{
					Command: []string{},
					Args:    []string{},
					Script:  "",
				},
				CompletionDeadlineSeconds: _oso_timeout,
			},
			OsoBuildRunPolicy: osopb3.DockerBuildConfigRequestData_Serial,
			Labels:            map[string]string{},
			Annotations:       map[string]string{},
		},
		TriggeredBy: []*osopb3.OsoBuildTriggerCause{
			{
				Message:          "No Trigger",
				GenericWebHook:   (*osopb3.GenericWebHookCause)(nil),
				GithubWebHook:    (*osopb3.GitHubWebHookCause)(nil),
				ImageChangeBuild: (*osopb3.ImageChangeCause)(nil),
			},
		},
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
}
