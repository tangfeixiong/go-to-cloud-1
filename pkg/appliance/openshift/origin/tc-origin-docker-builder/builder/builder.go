package builder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/crypto"

	s2iapi "github.com/openshift/source-to-image/pkg/api"

	"github.com/openshift/origin/pkg/build/api"
	bld "github.com/openshift/origin/pkg/build/builder"
	"github.com/openshift/origin/pkg/build/builder/cmd/scmauth"
	"github.com/openshift/origin/pkg/client"
	dockerutil "github.com/openshift/origin/pkg/cmd/util/docker"
	"github.com/openshift/origin/pkg/generate/git"
	"github.com/openshift/origin/pkg/version"
)

var (
	//masterVersion string = "v1.3.0-alpha.0-52-gbc1ddaa"
	masterVersion string = "v1.3.0-alpha.1-83-g16d6863"

	exampleBuild string = "/examples/github101-v1.3.json"

	buildName    string = "osobuilds"
	buildProject string = "tangfx"
)

type builder interface {
	Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *api.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error
}

type builderConfig struct {
	build           *api.Build
	sourceSecretDir string
	dockerClient    *docker.Client
	dockerEndpoint  string
	buildsClient    client.BuildInterface
}

func newBuilderFromDockerfileWithJSON(buildStr string) (*builderConfig, error) {
	cfg := &builderConfig{}
	var err error

	// build (BUILD)
	//buildStr := os.Getenv("BUILD")
	glog.V(4).Infof("$BUILD env var is %s \n", buildStr)
	cfg.build = &api.Build{}

	if err = runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), []byte(buildStr), cfg.build); err != nil {
		return nil, fmt.Errorf("unable to parse build: %v", err)
	}
	if cfg.build.Spec.Source.Dockerfile == nil {
		glog.V(9).Infoln("decode from encoding/json")
		if err = json.Unmarshal([]byte(buildStr), cfg.build); err != nil {
			glog.V(9).Infof("decode error: %s", err)
			return nil, err
		}
	}
	glog.V(7).Infof("tangfx > build: %+v, source: %+v, dockerfile: %s", cfg.build, cfg.build.Spec.Source, *cfg.build.Spec.Source.Dockerfile)

	//masterVersion := os.Getenv(api.OriginVersion)
	thisVersion := version.Get().String()
	if len(masterVersion) != 0 && masterVersion != thisVersion {
		glog.Warningf("Master version %q does not match Builder image version %q", masterVersion, thisVersion)
	} else {
		glog.V(2).Infof("Master version %q, Builder version %q", masterVersion, thisVersion)
	}

	// dockerClient and dockerEndpoint (DOCKER_HOST)
	// usually not set, defaults to docker socket
	cfg.dockerClient, cfg.dockerEndpoint, err = dockerutil.NewHelper().GetClient()
	if err != nil {
		return nil, fmt.Errorf("error obtaining docker client: %v", err)
	}

	// buildsClient (KUBERNETES_SERVICE_HOST, KUBERNETES_SERVICE_PORT)
	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		//return nil, fmt.Errorf("failed to get client config: %v", err)
		clientConfig, _ = InClusterConfig()
	}
	osClient, err := client.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error obtaining OpenShift client: %v", err)
	}
	cfg.buildsClient = osClient.Builds(cfg.build.Namespace)

	return cfg, nil
}

// InClusterConfig returns a config object which uses the service account
// kubernetes gives to pods. It's intended for clients that expect to be
// running inside a pod running on kuberenetes. It will return an error if
// called from a process not running in a kubernetes environment.
func InClusterConfig() (*restclient.Config, error) {
	//host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	host, port := "172.17.4.50", "443"
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}

	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/" + kapi.ServiceAccountTokenKey)
	if err != nil {
		//	return nil, err
		glog.Errorf("token: %s", err)
	}
	tlsClientConfig := restclient.TLSClientConfig{}
	rootCAFile := "/var/run/secrets/kubernetes.io/serviceaccount/" + kapi.ServiceAccountRootCAKey
	if _, err := crypto.CertPoolFromFile(rootCAFile); err != nil {
		glog.Errorf("Expected to load root CA config from %s, but got err: %v", rootCAFile, err)
	} else {
		tlsClientConfig.CAFile = rootCAFile
	}

	return &restclient.Config{
		// TODO: switch to using cluster DNS.
		Host:            "https://" + net.JoinHostPort(host, port),
		BearerToken:     string(token),
		TLSClientConfig: tlsClientConfig,
	}, nil
}

func newBuilderConfigFromEnvironment() (*builderConfig, error) {
	cfg := &builderConfig{}
	var err error

	// build (BUILD)
	buildStr := os.Getenv("BUILD")
	glog.V(4).Infof("$BUILD env var is %s \n", buildStr)
	cfg.build = &api.Build{}
	if err = runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), []byte(buildStr), cfg.build); err != nil {
		return nil, fmt.Errorf("unable to parse build: %v", err)
	}

	masterVersion := os.Getenv(api.OriginVersion)
	thisVersion := version.Get().String()
	if len(masterVersion) != 0 && masterVersion != thisVersion {
		glog.Warningf("Master version %q does not match Builder image version %q", masterVersion, thisVersion)
	} else {
		glog.V(2).Infof("Master version %q, Builder version %q", masterVersion, thisVersion)
	}

	// sourceSecretsDir (SOURCE_SECRET_PATH)
	cfg.sourceSecretDir = os.Getenv("SOURCE_SECRET_PATH")

	// dockerClient and dockerEndpoint (DOCKER_HOST)
	// usually not set, defaults to docker socket
	cfg.dockerClient, cfg.dockerEndpoint, err = dockerutil.NewHelper().GetClient()
	if err != nil {
		return nil, fmt.Errorf("error obtaining docker client: %v", err)
	}

	// buildsClient (KUBERNETES_SERVICE_HOST, KUBERNETES_SERVICE_PORT)
	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config: %v", err)
	}
	osClient, err := client.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("error obtaining OpenShift client: %v", err)
	}
	cfg.buildsClient = osClient.Builds(cfg.build.Namespace)

	return cfg, nil
}

func (c *builderConfig) setupGitEnvironment() ([]string, error) {

	gitSource := c.build.Spec.Source.Git

	// For now, we only handle git. If not specified, we're done
	if gitSource == nil {
		return []string{}, nil
	}

	sourceSecret := c.build.Spec.Source.SourceSecret
	gitEnv := []string{"GIT_ASKPASS=true"}
	// If a source secret is present, set it up and add its environment variables
	if sourceSecret != nil {
		// TODO: this should be refactored to let each source type manage which secrets
		//   it accepts
		sourceURL, err := git.ParseRepository(gitSource.URI)
		if err != nil {
			return nil, fmt.Errorf("cannot parse build URL: %s", gitSource.URI)
		}
		scmAuths := scmauth.GitAuths(sourceURL)

		// TODO: remove when not necessary to fix up the secret dir permission
		sourceSecretDir, err := fixSecretPermissions(c.sourceSecretDir)
		if err != nil {
			return nil, fmt.Errorf("cannot fix source secret permissions: %v", err)
		}

		secretsEnv, overrideURL, err := scmAuths.Setup(sourceSecretDir)
		if err != nil {
			return nil, fmt.Errorf("cannot setup source secret: %v", err)
		}
		if overrideURL != nil {
			c.build.Annotations[bld.OriginalSourceURLAnnotationKey] = gitSource.URI
			gitSource.URI = overrideURL.String()
		}
		gitEnv = append(gitEnv, secretsEnv...)
	}
	if gitSource.HTTPProxy != nil && len(*gitSource.HTTPProxy) > 0 {
		gitEnv = append(gitEnv, fmt.Sprintf("HTTP_PROXY=%s", *gitSource.HTTPProxy))
		gitEnv = append(gitEnv, fmt.Sprintf("http_proxy=%s", *gitSource.HTTPProxy))
	}
	if gitSource.HTTPSProxy != nil && len(*gitSource.HTTPSProxy) > 0 {
		gitEnv = append(gitEnv, fmt.Sprintf("HTTPS_PROXY=%s", *gitSource.HTTPSProxy))
		gitEnv = append(gitEnv, fmt.Sprintf("https_proxy=%s", *gitSource.HTTPSProxy))
	}
	return bld.MergeEnv(os.Environ(), gitEnv), nil
}

// execute is responsible for running a build
func (c *builderConfig) execute(b builder) error {

	gitEnv, err := c.setupGitEnvironment()
	if err != nil {
		return err
	}
	gitClient := git.NewRepositoryWithEnv(gitEnv)

	cgLimits, err := bld.GetCGroupLimits()
	if err != nil {
		return fmt.Errorf("failed to retrieve cgroup limits: %v", err)
	}
	glog.V(2).Infof("Running build with cgroup limits: %#v", *cgLimits)
	glog.Infof("git: %+v, %+v, cgroup: %+v", gitEnv, gitClient, cgLimits)

	if err := b.Build(c.dockerClient, c.dockerEndpoint, c.buildsClient, c.build, gitClient, cgLimits); err != nil {
		return fmt.Errorf("build error: %v", err)
	}

	glog.Infof("Build: %+v", c.build)
	if c.build.Spec.Output.To == nil || len(c.build.Spec.Output.To.Name) == 0 {
		glog.Warning("Build does not have an Output defined, no output image was pushed to a registry.")
	}

	return nil
}

// fixSecretPermissions loweres access permissions to very low acceptable level
// TODO: this method should be removed as soon as secrets permissions are fixed upstream
// Kubernetes issue: https://github.com/kubernetes/kubernetes/issues/4789
func fixSecretPermissions(secretsDir string) (string, error) {
	secretTmpDir, err := ioutil.TempDir("", "tmpsecret")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("cp", "-R", ".", secretTmpDir)
	cmd.Dir = secretsDir
	if err := cmd.Run(); err != nil {
		return "", err
	}
	secretFiles, err := ioutil.ReadDir(secretTmpDir)
	if err != nil {
		return "", err
	}
	for _, file := range secretFiles {
		if err := os.Chmod(filepath.Join(secretTmpDir, file.Name()), 0600); err != nil {
			return "", err
		}
	}
	return secretTmpDir, nil
}

type dockerBuilder struct{}

// Build starts a Docker build.
func (dockerBuilder) Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *api.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error {
	glog.Infof("Go to docker build %+v with %+v", build.Spec.Output.To, build)
	if build.Name == "" {
		build.Name = buildName
		build.Namespace = buildProject
	}
	if build.Spec.Output.To != nil &&
		build.Spec.Output.To.Kind == "DockerImage" &&
		build.Spec.Output.To.Name != "" {
		build.Status.OutputDockerImageReference = build.Spec.Output.To.Name
	}
	return bld.NewDockerBuilder(dockerClient, buildsClient, build, gitClient, cgLimits).Build()
}

type s2iBuilder struct{}

// Build starts an S2I build.
func (s2iBuilder) Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *api.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error {
	return bld.NewS2IBuilder(dockerClient, sock, buildsClient, build, gitClient, cgLimits).Build()
}

func runBuildDifferently(builder builder) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	//f := wd + "/examples/build101.json"
	f := wd + exampleBuild
	glog.V(9).Infof("path: %s", f)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		panic(err)
	}
	cfg, err := newBuilderFromDockerfileWithJSON(string(b))

	//cfg, err := newBuilderConfigFromEnvironment()
	if err != nil {
		glog.Fatalf("Cannot setup builder configuration: %v", err)
	}
	err = cfg.execute(builder)
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
}

func runBuild(builder builder) {
	cfg, err := newBuilderConfigFromEnvironment()
	if err != nil {
		glog.Fatalf("Cannot setup builder configuration: %v", err)
	}
	err = cfg.execute(builder)
	if err != nil {
		glog.Fatalf("Error: %v", err)
	}
}

// RunDockerBuild creates a docker builder and runs its build
func RunDockerBuild() {
	//runBuild(dockerBuilder{})
	runBuildDifferently(dockerBuilder{})
}

// RunSTIBuild creates a STI builder and runs its build
func RunSTIBuild() {
	runBuild(s2iBuilder{})
}
