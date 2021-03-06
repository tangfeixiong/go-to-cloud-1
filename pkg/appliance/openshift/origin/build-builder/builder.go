package builder

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"

	s2iapi "github.com/openshift/source-to-image/pkg/api"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	"github.com/openshift/origin/pkg/build/api/validation"
	bld "github.com/openshift/origin/pkg/build/builder"
	"github.com/openshift/origin/pkg/build/builder/cmd/scmauth"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	dockerutil "github.com/openshift/origin/pkg/cmd/util/docker"
	"github.com/openshift/origin/pkg/generate/git"
	"github.com/openshift/origin/pkg/version"
)

var (
	openshift_master_version string = "v1.3.0-alpha.2+88b8a33-dirty"
	//openshift_master_version string = "v1.3.0-alpha.1-83-g16d6863"
	//openshift_master_version string = "v1.3.0-alpha.0-52-gbc1ddaa"
	docker_build_name         string
	docker_build_project_name string
	source_build_name         string
	source_build_project_name string
)

type builder interface {
	Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *buildapi.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error
}

type builderConfig struct {
	out             io.Writer
	build           *buildapi.Build
	sourceSecretDir string
	dockerClient    *docker.Client
	dockerEndpoint  string
	buildsClient    client.BuildInterface
}

func newBuilderConfigFromEnvironment(out io.Writer) (*builderConfig, error) {
	cfg := &builderConfig{}
	var err error

	cfg.out = out

	// build (BUILD)
	buildStr := os.Getenv("BUILD")
	glog.V(4).Infof("$BUILD env var is %s \n", buildStr)
	cfg.build = &buildapi.Build{}
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), []byte(buildStr), cfg.build); err != nil {
		return nil, fmt.Errorf("unable to parse build: %v", err)
	}
	if errs := validation.ValidateBuild(cfg.build); len(errs) > 0 {
		return nil, errors.NewInvalid(unversioned.GroupKind{Kind: "Build"}, cfg.build.Name, errs)
	}
	glog.V(4).Infof("Build: %#v", cfg.build)

	masterVersion := os.Getenv(buildapi.OriginVersion)
	thisVersion := version.Get().String()
	if len(masterVersion) != 0 && masterVersion != thisVersion {
		glog.V(3).Infof("warning: OpenShift server version %q differs from this image %q\n", masterVersion, thisVersion)
	} else {
		glog.V(4).Infof("Master version %q, Builder version %q", masterVersion, thisVersion)
	}

	// sourceSecretsDir (SOURCE_SECRET_PATH)
	cfg.sourceSecretDir = os.Getenv("SOURCE_SECRET_PATH")

	// dockerClient and dockerEndpoint (DOCKER_HOST)
	// usually not set, defaults to docker socket
	cfg.dockerClient, cfg.dockerEndpoint, err = dockerutil.NewHelper().GetClient()
	if err != nil {
		return nil, fmt.Errorf("no Docker configuration defined: %v", err)
	}

	// buildsClient (KUBERNETES_SERVICE_HOST, KUBERNETES_SERVICE_PORT)
	clientConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the server: %v", err)
	}
	osClient, err := client.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %v", err)
	}
	cfg.buildsClient = osClient.Builds(cfg.build.Namespace)

	return cfg, nil
}

func (c *builderConfig) setupGitEnvironment() ([]string, error) {
	// For now, we only handle git. If not specified, we're done
	gitSource := c.build.Spec.Source.Git
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
	fmt.Fprintf(c.out, "Setup git environment\n")
	gitEnv, err := c.setupGitEnvironment()
	if err != nil {
		return err
	}
	gitClient := git.NewRepositoryWithEnv(gitEnv)

	cgLimits, err := bld.GetCGroupLimits()
	if err != nil {
		return fmt.Errorf("failed to retrieve cgroup limits: %v", err)
	}
	glog.V(4).Infof("Running build with cgroup limits: %#v", *cgLimits)

	fmt.Fprintf(c.out, "Start docker build with cgroup configured: %+v\n", *cgLimits)
	if err := b.Build(c.dockerClient, c.dockerEndpoint, c.buildsClient, c.build, gitClient, cgLimits); err != nil {
		return fmt.Errorf("build error: %v", err)
	}

	if c.build.Spec.Output.To == nil || len(c.build.Spec.Output.To.Name) == 0 {
		fmt.Fprintf(c.out, "Build complete, no image push requested\n")
	}
	fmt.Fprintf(c.out, "Build complete, image pushed\n")

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
func (dockerBuilder) Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *buildapi.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error {
	if build.Name == "" {
		glog.Warningln("Lost build name and project during deserialization")
		build.Name = docker_build_name
		build.Namespace = docker_build_project_name
	}
	if build.Spec.Output.To != nil &&
		build.Spec.Output.To.Kind == "DockerImage" &&
		build.Spec.Output.To.Name != "" &&
		build.Status.OutputDockerImageReference != build.Spec.Output.To.Name {
		glog.Warningf("Force output docker image (%s) with predefined value (%s)",
			build.Status.OutputDockerImageReference, build.Spec.Output.To.Name)
		build.Status.OutputDockerImageReference = build.Spec.Output.To.Name
	}
	return bld.NewDockerBuilder(dockerClient, buildsClient, build, gitClient, cgLimits).Build()
}

type s2iBuilder struct{}

// Build starts an S2I build.
func (s2iBuilder) Build(dockerClient bld.DockerClient, sock string, buildsClient client.BuildInterface, build *buildapi.Build, gitClient bld.GitClient, cgLimits *s2iapi.CGroupLimits) error {
	return bld.NewS2IBuilder(dockerClient, sock, buildsClient, build, gitClient, cgLimits).Build()
}

func runBuild(out io.Writer, builder builder) error {
	cfg, err := newBuilderConfigFromEnvironment(out)
	if err != nil {
		return err
	}
	return cfg.execute(builder)
}

// RunDockerBuild creates a docker builder and runs its build
func RunDockerBuild(out io.Writer, data *buildapi.Build, ccfactory *clientcmd.Factory) (*buildapi.Build, error) {
	//return runBuild(out, dockerBuilder{})

	cfg := &builderConfig{}
	var err error

	cfg.out = out

	// build (BUILD)
	buildStr := os.Getenv("BUILD")
	//glog.V(4).Infof("$BUILD env var is %s \n", buildStr)
	cfg.build = &buildapi.Build{}

	b := &bytes.Buffer{}
	//if b, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion), data); err != nil {
	//	glog.Errorf("Could not feed data: %+v", err)
	//	return nil, fmt.Errorf("unable to parse build: %v", err)
	//}
	if err := codec.JSON.Encode(b).One(data); err != nil {
		glog.Errorf("Could not feed data: %+v", err)
		return nil, fmt.Errorf("Could not feed data: %v", err)
	}
	buildStr = b.String()
	glog.Infof("build object: %s", buildStr)
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), []byte(buildStr), cfg.build); err != nil {
		glog.Errorf("Could not validate raw bytes: %s", err)
		return nil, fmt.Errorf("unable to parse build: %v", err)
	}
	if cfg.build.Name == "" || cfg.build.Namespace == "" ||
		cfg.build.Spec.Output.To == nil ||
		(cfg.build.Spec.Output.To.Kind == "DockerImage" &&
			cfg.build.Spec.Output.To.Name == "") {
		glog.Infoln("Missing key fields: name or project or output to ...")
		return nil, fmt.Errorf("Missing key fields")
	}
	docker_build_name = cfg.build.Name
	docker_build_project_name = cfg.build.Namespace

	if errs := validation.ValidateBuild(cfg.build); len(errs) > 0 {
		glog.Infof("Could not validate object fields: %+v", errs)
		return nil, errors.NewInvalid(unversioned.GroupKind{Kind: "Build"}, cfg.build.Name, errs)
	}
	glog.V(4).Infof("Build: %#v", cfg.build)

	masterVersion := os.Getenv(buildapi.OriginVersion)
	if masterVersion == "" {
		masterVersion = openshift_master_version
	}
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
		glog.Infof("Could not setup docker client: %+v", err)
		return nil, fmt.Errorf("error obtaining docker client: %v", err)
	}

	// buildsClient (KUBERNETES_SERVICE_HOST, KUBERNETES_SERVICE_PORT)
	//clientConfig, err := restclient.InClusterConfig()
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get client config: %v", err)
	//}
	//osClient, err := client.New(clientConfig)
	osClient, _, err := ccfactory.Clients()
	if err != nil {
		glog.Infof("Could not setup openshift client: %+v", err)
		return nil, fmt.Errorf("error obtaining OpenShift client: %v", err)
	}
	cfg.buildsClient = osClient.Builds(cfg.build.Namespace)

	return data, cfg.execute(dockerBuilder{})
}

// RunS2IBuild creates a S2I builder and runs its build
func RunS2IBuild(out io.Writer, data *buildapiv1.Build, ccfactory *clientcmd.Factory) error {
	//return runBuild(out, s2iBuilder{})

	if data == nil {
		glog.Errorf("data required")
		return fmt.Errorf("data required")
	}
	if ccfactory == nil {
		glog.Errorf("occtl config required")
		return fmt.Errorf("occtl required")
	} else {
		mapper, _ := ccfactory.Object(false)
		kapi.RegisterRESTMapper(mapper)
		buildapi.AddToScheme(kapi.Scheme)
		buildapiv1.AddToScheme(kapi.Scheme)

	}

	cfg := &builderConfig{}
	var err error

	cfg.out = out

	// build (BUILD)
	buildStr := os.Getenv("BUILD")
	//glog.V(4).Infof("$BUILD env var is %s \n", buildStr)
	cfg.build = &buildapi.Build{}

	var buf *bytes.Buffer
	b, err := runtime.Encode(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), data)
	if err != nil {
		glog.Errorf("Could not transform data: %+v", err)
		return fmt.Errorf("unable to parse build: %v", err)
	}
	buf = bytes.NewBuffer(b)
	//	if err := codec.JSON.Encode(b).One(data); err != nil {
	//		glog.Errorf("Could not transform data with helm codec: %+v", err)
	//		return fmt.Errorf("Could not transform data: %v", err)
	//	}
	buildStr = buf.String()
	glog.Infof("$BUILD env var is %s \n", buildStr)
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), buf.Bytes(), cfg.build); err != nil {
		glog.Errorf("Could not decode raw bytes: %s", err)
		return fmt.Errorf("unable to parse build: %v", err)
	}
	if cfg.build.Name == "" || cfg.build.Namespace == "" ||
		cfg.build.Spec.Output.To == nil ||
		(len(cfg.build.Spec.Output.To.Kind) == 0 && len(cfg.build.Spec.Output.To.Name) == 0) {
		glog.Errorln("Missing key fields: name or project or output to ...")
		return fmt.Errorf("Missing key fields")
	}
	//docker_build_name = cfg.build.Name
	//docker_build_project_name = cfg.build.Namespace
	if len(cfg.build.Status.OutputDockerImageReference) == 0 {
		glog.Warningf("Force output docker image reference with predefined name(%s)",
			data.Spec.Output.To.Name)
		cfg.build.Status.OutputDockerImageReference = data.Spec.Output.To.Name
	}
	if errs := validation.ValidateBuild(cfg.build); len(errs) > 0 {
		glog.Infof("Could not validate object fields: %+v", errs)
		return errors.NewInvalid(unversioned.GroupKind{Kind: "Build"}, cfg.build.Name, errs)
	}
	glog.V(4).Infof("Build: %#v", cfg.build)

	masterVersion := os.Getenv(buildapi.OriginVersion)
	if len(masterVersion) == 0 {
		masterVersion = openshift_master_version
	}
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
		glog.Infof("Could not setup docker client: %+v", err)
		return fmt.Errorf("error obtaining docker client: %v", err)
	}

	// buildsClient (KUBERNETES_SERVICE_HOST, KUBERNETES_SERVICE_PORT)
	//clientConfig, err := restclient.InClusterConfig()
	//if err != nil {
	//	return fmt.Errorf("failed to get client config: %v", err)
	//}
	//osClient, err := client.New(clientConfig)
	osClient, _, err := ccfactory.Clients()
	if err != nil {
		glog.Infof("Could not setup openshift client: %+v", err)
		return fmt.Errorf("error obtaining OpenShift client: %v", err)
	}
	cfg.buildsClient = osClient.Builds(cfg.build.Namespace)

	return cfg.execute(s2iBuilder{})
}
