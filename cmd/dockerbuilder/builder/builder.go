
package builder


import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"

	s2iapi "github.com/openshift/source-to-image/pkg/api"

	"github.com/openshift/origin/pkg/build/api"
	bld "github.com/openshift/origin/pkg/build/builder"
	"github.com/openshift/origin/pkg/build/builder/cmd/scmauth"
	"github.com/openshift/origin/pkg/client"
	dockerutil "github.com/openshift/origin/pkg/cmd/util/docker"
	"github.com/openshift/origin/pkg/generate/git"
	"github.com/openshift/origin/pkg/version"
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

func newBuilderConfigWithDockerfile(dockerBuildRoot string) (*builderConfig, error) {
	cfg := &builderConfig{}

	masterVersion := os.Getenv(api.OriginVersion)
	thisVersion := version.Get().String()
	if len(masterVersion) != 0 && masterVersion != thisVersion {
		glog.Warningf("Master version %q does not match Builder image version %q", masterVersion, thisVersion)
	} else {
		glog.V(2).Infof("Master version %q, Builder version %q", masterVersion, thisVersion)
	}

    var err error
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

	if err := b.Build(c.dockerClient, c.dockerEndpoint, c.buildsClient, c.build, gitClient, cgLimits); err != nil {
		return fmt.Errorf("build error: %v", err)
	}

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
	return bld.NewDockerBuilder(dockerClient, buildsClient, build, gitClient, cgLimits).Build()
}

type s2iBuilder struct{}

// Build starts an S2I build.
func (s2iBuilder) Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *api.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error {
	return bld.NewS2IBuilder(dockerClient, sock, buildsClient, build, gitClient, cgLimits).Build()
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
	runBuild(dockerBuilder{})
}

// RunSTIBuild creates a STI builder and runs its build
func RunSTIBuild() {
	runBuild(s2iBuilder{})
}
