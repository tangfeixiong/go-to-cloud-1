package origin

import (
	buildapi "github.com/openshift/origin/pkg/build/api"
	kapi "k8s.io/kubernetes/pkg/api"
)

var (
	FinalizerVender string = "qingyuanos.io/paas"
)

type DockerImageBuild struct {
	DockerfileContent    string
	GitSourcedDockerfile *GitSourcedDockerfile
	PullCredential       BasicAuth
	PullSecretName       string
	PullAlways           bool
	CacheLayers          bool
	Env                  []kapi.EnvVar
	Repo                 string
	PushCredential       BasicAuth
	PushSecretName       string
	Trigger              *buildapi.BuildTriggerPolicy
}

type GitSourcedDockerfile struct {
	URL                   string
	Ref                   string
	DockerfileContextPath string
	BasicAuth             BasicAuth
	GitConfig             string
	SecretName            string
}

type HttpSourcedDockerfile struct {
	URL string
}

type FtpSourcedDockerfile struct {
	URL string
}

type SwiftSourcedDockerfile struct {
	URL string
}

type BasicAuth struct {
	Username string
	Password string
	Token    string
	Certs    string
}
