package osoc

import (
	//"encoding/base64"
	"fmt"
	//"io/ioutil"
	//"strings"

	//k8sapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	//"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	//kclient "k8s.io/kubernetes/pkg/client/unversioned"
	//kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	//kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"

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
				Message:          "Manually Trigger",
				GenericWebHook:   (*osopb3.GenericWebHookCause)(nil),
				GithubWebHook:    (*osopb3.GitHubWebHookCause)(nil),
				ImageChangeBuild: (*osopb3.ImageChangeCause)(nil),
			},
		},
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
}

type DockerBuildRequestDataUtility struct {
	//kcc              kclientcmd.ClientConfig
	kubeconfigPath   string
	kubeContext      string
	apiServer        string
	target           *osopb3.DockerBuildRequestData
	dockerfileUsed   bool
	gitUsed          bool
	imgUsed          bool
	outputConfigured bool
}

func NewDockerBuildRequestDataUtility() *DockerBuildRequestDataUtility {
	return &DockerBuildRequestDataUtility{
		target: internalDockerBuildRequestData(),
	}
}

/*func NewDockerBuildRequestDataUtility(kubeconfigPath, kubeContext, apiServer string) *DockerBuildRequestDataUtility {
	return &DockerBuildRequestDataUtility{
		//kcc:    kcc,
		kubeconfigPath: kubeconfigPath,
		kubeContext:    kubeContext,
		apiServer:      apiServer,
		target:         internalDockerBuildRequestData(),
	}
}*/

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
/*func directKClientConfig(kubeconfigPath, kubeContext, apiServer string) (kclientcmd.ClientConfig, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	conf, err := kclientcmd.Load(data)
	if err != nil {
		return nil, err
	}

	return kclientcmd.NewNonInteractiveClientConfig(*conf, kubeContext,
		&kclientcmd.ConfigOverrides{
			ClusterInfo: kclientcmdapi.Cluster{
				Server: apiServer,
			},
		},
		kclientcmd.NewDefaultClientConfigLoadingRules()), nil
}

func (b *DockerBuildRequestDataUtility) RetrieveDockerSecret(project, repo, username, password, email string) (string, error) {
	kcc, err := directKClientConfig(b.kubeconfigPath, b.kubeContext, b.apiServer)
	if err != nil {
		return "", err
	}
	cc, err := kcc.ClientConfig()
	if err != nil {
		return "", err
	}
	client, err := kclient.New(cc)
	if err != nil {
		return "", err
	}

	offset := strings.LastIndex(repo, "/")
	account := repo
	if offset > 0 {
		account = repo[0:offset]
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(account))
	secret, err := client.Secrets(project).Get(sEnc)
	if err != nil {
		return "", err
	}
	if secret != nil && string(secret.Type) != string(kapi.SecretTypeDockercfg) {
		return "", fmt.Errorf("secret is existed as not dockercfg")
	}
	sAuth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	if secret != nil && string(secret.Data[kapi.DockerConfigKey]) == sAuth {
		return sEnc, nil
	}
	if secret != nil {
		secret.Data[kapi.DockerConfigKey] = []byte(sEnc)
		if _, err := client.Secrets(project).Update(secret); err != nil {
			return "", err
		}
	} else {
		secret = &k8sapi.Secret{
			TypeMeta: unversioned.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: k8sapi.ObjectMeta{
				Name:      sEnc,
				Namespace: project,
				Annotations: map[string]string{
					"qingyuanos.com/docker-registry": account,
				},
			},
			Data: map[string][]byte{
				k8sapi.DockerConfigKey: []byte(sAuth),
			},
			Type: k8sapi.SecretTypeDockercfg,
		}
		if _, err := client.Secrets(project).Create(secret); err != nil {
			return "", err
		}
		sa, err := client.ServiceAccounts(project).Get("builder")
		if err != nil {
			return "", err
		}
		sa.Secrets = append(sa.Secrets, k8sapi.ObjectReference{Name: sEnc})
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, k8sapi.LocalObjectReference{Name: sEnc})
		if _, err := client.ServiceAccounts(project).Update(sa); err != nil {
			return "", err
		}
	}
	return sEnc, nil
}

func (b *DockerBuildRequestDataUtility) RetrieveGitSecretBasicAuth(project, repo, username, password string) (string, error) {
	kcc, err := directKClientConfig(b.kubeconfigPath, b.kubeContext, b.apiServer)
	if err != nil {
		return "", err
	}
	cc, err := kcc.ClientConfig()
	if err != nil {
		return "", err
	}
	client, err := kclient.New(cc)
	if err != nil {
		return "", err
	}

	offset := strings.LastIndex(repo, "://")
	account := repo
	if offset > 0 {
		offset += 3
		n := strings.Index(repo[offset:], "/")
		if n > 0 {
			offset += n
			n = strings.Index(repo[offset:], "/")
			if n > 0 {
				offset += n
			}
		}
		account = repo[0:offset]
	}
	sEnc := base64.StdEncoding.EncodeToString([]byte(account))
	secret, err := client.Secrets(project).Get(sEnc)
	if err != nil {
		return "", err
	}
	if string(secret.Type) != string(kapi.SecretTypeOpaque) {
		return "", fmt.Errorf("secret is existed as not opaque")
	}
	sAuth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	if string(secret.Data["BasicAuth"]) != sAuth {
		secret.Data["BasicAuth"] = []byte(sEnc)
		if _, err := client.Secrets(project).Update(secret); err != nil {
			return "", err
		}
	}
	return sEnc, nil
}*/

func (b *DockerBuildRequestDataUtility) Result() (*osopb3.DockerBuildRequestData, error) {
	if b.target == nil {
		return nil, fmt.Errorf("not initialized")
	}
	if b.target.ProjectName == "" || b.target.Name == "" ||
		b.target.Configuration.ProjectName == "" || b.target.Configuration.Name == "" {
		return nil, fmt.Errorf("not identified")
	}
	if b.target.Configuration.CommonSpec == nil {
		return nil, fmt.Errorf("not configured")
	}
	if b.target.Configuration.CommonSpec.Source == nil {
		return nil, fmt.Errorf("source not configured")
	}
	if false == (b.dockerfileUsed || b.gitUsed || b.imgUsed) {
		return nil, fmt.Errorf("invalid source")
	}
	if false == b.outputConfigured {
		return nil, fmt.Errorf("output option not provided")
	}

	return b.target, nil
}

func (b *DockerBuildRequestDataUtility) BuilderName(project, name string) *DockerBuildRequestDataUtility {
	if b.target == nil {
		b.target = internalDockerBuildRequestData()
	}
	b.target.ProjectName = project
	b.target.Name = name
	b.target.Configuration.ProjectName = project
	b.target.Configuration.Name = name
	return b
}

func (b *DockerBuildRequestDataUtility) Dockerfile(dockerfile string) *DockerBuildRequestDataUtility {
	if b.target == nil {
		b.target = internalDockerBuildRequestData()
	}
	if b.dockerfileUsed = (dockerfile != ""); b.dockerfileUsed {
		b.target.Configuration.CommonSpec.Source.Dockerfile = dockerfile
		if b.target.Configuration.CommonSpec.Source.Type == osopb3.OsoBuildSourceType_None.String() {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Dockerfile.String()
		}
	} else {
		b.target.Configuration.CommonSpec.Source.Dockerfile = ""
		if b.target.Configuration.CommonSpec.Source.Git != nil {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Git.String()
		} else if len(b.target.Configuration.CommonSpec.Source.Images) > 0 {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Image.String()
		} else {
			b.target.Configuration.CommonSpec.Source.Type = ""
		}
	}
	return b
}

func (b *DockerBuildRequestDataUtility) Git(uri, ref, path string) *DockerBuildRequestDataUtility {
	if b.target == nil {
		b.target = internalDockerBuildRequestData()
	}
	if b.gitUsed = (uri != ""); b.gitUsed {
		if b.target.Configuration.CommonSpec.Source.Git == nil {
			b.target.Configuration.CommonSpec.Source.Git = &osopb3.GitBuildSource{}
		}
		b.target.Configuration.CommonSpec.Source.Git.Uri = uri
		b.target.Configuration.CommonSpec.Source.Git.Ref = ""
		b.target.Configuration.CommonSpec.Source.ContextDir = path
		if b.target.Configuration.CommonSpec.Source.Type == osopb3.OsoBuildSourceType_None.String() {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Dockerfile.String()
		}
	} else {
		b.target.Configuration.CommonSpec.Source.Git = nil
		if len(b.target.Configuration.CommonSpec.Source.Images) > 0 {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Image.String()
		} else if b.target.Configuration.CommonSpec.Source.Dockerfile != "" {
			b.target.Configuration.CommonSpec.Source.Type = osopb3.OsoBuildSourceType_Dockerfile.String()
		} else {
			b.target.Configuration.CommonSpec.Source.Type = ""
		}
	}
	return b
}

func (b *DockerBuildRequestDataUtility) DockerBuildStrategy(overrideBaseImage,
	pullSecret, overrideDockerfilePath string,
	cacheUsed, forcePull bool) *DockerBuildRequestDataUtility {
	if b.target == nil {
		b.target = internalDockerBuildRequestData()
	}
	if b.target.Configuration.CommonSpec.Strategy == nil {
		b.target.Configuration.CommonSpec.Strategy = &osopb3.BuildStrategy{}
	}
	b.target.Configuration.CommonSpec.Strategy.Type = osopb3.BuildStrategy_Docker.String()
	b.target.Configuration.CommonSpec.Strategy.DockerStrategy = &osopb3.DockerBuildStrategy{
		From:           (*kapi.ObjectReference)(nil),
		PullSecret:     (*kapi.LocalObjectReference)(nil),
		NoCache:        !cacheUsed,
		Env:            []*kapi.EnvVar{},
		ForcePull:      forcePull,
		DockerfilePath: overrideDockerfilePath,
	}
	if pullSecret != "" {
		b.target.Configuration.CommonSpec.Strategy.DockerStrategy.PullSecret = &kapi.LocalObjectReference{
			Name: pullSecret,
		}
	}
	if overrideBaseImage != "" {
		st := osopb3.OsoBuildStrategyObjectReferenceType_Strategy_DockerImage.String()
		b.target.Configuration.CommonSpec.Strategy.DockerStrategy.From = &kapi.ObjectReference{
			Kind: st[len("Strategy_"):],
			Name: overrideBaseImage,
		}
	}
	return b
}

func (b *DockerBuildRequestDataUtility) DockerBuildOutputOption(pushRepo,
	pushSecret string) *DockerBuildRequestDataUtility {
	if b.target == nil {
		b.target = internalDockerBuildRequestData()
	}
	if b.target.Configuration.CommonSpec.Output == nil {
		b.target.Configuration.CommonSpec.Output = &osopb3.BuildOutput{}
	}
	if pushSecret != "" {
		b.target.Configuration.CommonSpec.Output.PushSecret = &kapi.LocalObjectReference{
			Name: pushSecret,
		}
	}
	if b.outputConfigured = (pushRepo != ""); b.outputConfigured {
		ot := osopb3.OsoBuildOutputObjectReferenceType_Output_DockerImage.String()
		b.target.Configuration.CommonSpec.Output.To = &kapi.ObjectReference{
			Kind: ot[len("Output_"):],
			Name: pushRepo,
		}
	}
	return b
}
