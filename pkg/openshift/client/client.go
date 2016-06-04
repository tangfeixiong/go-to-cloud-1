package client

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	_ "github.com/openshift/origin/pkg/api/install"
	//buildapi "github.com/openshift/origin/pkg/build/api"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	_ "github.com/openshift/origin/pkg/cmd/util/tokencmd"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	_ "github.com/openshift/origin/pkg/generate/git"
	userapi "github.com/openshift/origin/pkg/user/api"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	// kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	clientauth "k8s.io/kubernetes/pkg/client/unversioned/auth"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	// kubecmdconfig "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	_ "k8s.io/kubernetes/pkg/runtime"
	_ "k8s.io/kubernetes/pkg/runtime/serializer/json"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

	factory *osclientcmd.Factory

	kubeconfigPath    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/kubeconfig"
	kubeconfigContext string = "openshift-origin-single"

	apiVersion string = "v1"
	oca        string = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	oclientcrt string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	oclientkey string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"

	// token string = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"

	oconfigPath    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
	oconfigContext string = "default/172-17-4-50:30448/system:admin"

	kClientConfig, oClientConfig *clientcmd.ClientConfig
	kClient                      *kclient.Client
	oClient                      *client.Client
	kConfig, oConfig             *restclient.Config
)

type stringValue struct {
	value string
}

func (val stringValue) String() string {
	return val.value
}
func (val stringValue) Set(v string) error {
	val.value = v
	return nil
}
func (val stringValue) Type() string {
	return "string"
}

type intValue struct {
	value int
}

func (val intValue) String() string {
	return strconv.Itoa(val.value)
}
func (val intValue) Set(v string) error {
	var err error
	val.value, err = strconv.Atoi(v)
	return err
}
func (val intValue) Type() string {
	return "int"
}

type boolValue struct {
	value bool
}

func (val boolValue) String() string {
	return strconv.FormatBool(val.value)
}
func (val boolValue) Set(v string) error {
	var err error
	val.value, err = strconv.ParseBool(v)
	return err
}
func (val boolValue) Type() string {
	return "bool"
}

func options(c *cobra.Command) {
	if val := c.Flags().Lookup("server"); val != nil {
		val.Value.Set("https://172.17.4.50:30448")
	} else {
		val = c.Flags().VarPF(stringValue{"https://172.17.4.50:30448"}, "server", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag server: %+v\n", val)
	}
	c.Flags().Lookup("server").NoOptDefVal = "https://172.17.4.50:30448"

	if val := c.Flags().Lookup("client-certificate"); val != nil {
		val.Value.Set(oclientcrt)
	} else {
		val = c.Flags().VarPF(stringValue{oclientcrt}, "client-certificate", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag server: %+v\n", val)
	}
	c.Flags().Lookup("client-certificate").NoOptDefVal = oclientcrt
	if val := c.Flags().Lookup("client-key"); val != nil {
		val.Value.Set(oclientkey)
	} else {
		val = c.Flags().VarPF(stringValue{oclientkey}, "client-key", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] client-key: %+v\n", val)
	}
	c.Flags().Lookup("client-key").NoOptDefVal = oclientkey

	if val := c.Flags().Lookup("api-version"); val != nil {
		val.Value.Set(apiVersion)
	} else {
		val = c.Flags().VarPF(stringValue{apiVersion}, "api-version", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] api version: %+v\n", val)
	}
	c.Flags().Lookup("api-version").NoOptDefVal = apiVersion

	if val := c.Flags().Lookup("api-version"); val != nil {
		val.Value.Set(apiVersion)
	} else {
		val = c.Flags().VarPF(stringValue{apiVersion}, "api-version", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] api version: %+v\n", val)
	}
	c.Flags().Lookup("api-version").NoOptDefVal = apiVersion

	if val := c.Flags().Lookup("certificate-authority"); val != nil {
		val.Value.Set(oca)
	} else {
		val = c.Flags().VarPF(stringValue{oca}, "certificate-authority", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] ca: %+v\n", val)
	}
	c.Flags().Lookup("certificate-authority").NoOptDefVal = oca

	if val := c.Flags().Lookup("insecure-skip-tls-verify"); val != nil {
		val.Value.Set("false")
	} else {
		val = c.Flags().VarPF(boolValue{false}, "insecure-skip-tls-verify", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] insecure tls: %+v\n", val)
	}
	c.Flags().Lookup("insecure-skip-tls-verify").NoOptDefVal = "false"

	if val := c.Flags().Lookup("config"); val != nil {
		val.Value.Set(oconfigPath)
	} else {
		val = c.Flags().VarPF(stringValue{oconfigPath}, "config", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] config: %+v\n", val)
	}
	c.Flags().Lookup("config").NoOptDefVal = oconfigPath

}

var rootCommand = &cobra.Command{
	Use:   "oc",
	Short: "ociacibuilds",
	Long:  "The openshift image build server.",
	Run: func(c *cobra.Command, args []string) {
	},
}

func init() {
	flagtypes.GLog(rootCommand.PersistentFlags())

}

func overrideRootCommand() {

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	//f := osclientcmd.New(flags)
	cmds := cli.NewCommandCLI("oc", "oc", os.Stdin, os.Stdout, os.Stderr)
	cmds.Aliases = []string{"oc"}
	cmds.Use = "oc"
	cmds.Short = "openshift client"
	flags.VisitAll(func(flag *pflag.Flag) {
		if f := cmds.PersistentFlags().Lookup(flag.Name); f == nil {
			glog.V(5).Infof("flag: %v", flag.Name)
			cmds.PersistentFlags().AddFlag(flag)
		} else {
			glog.V(5).Infof("already registered flag %s", flag.Name)
		}
	})
	//cmds.PersistentFlags().Var(flags.Lookup("config").Value, "config", "Specify a kubeconfig file to define the configuration")
	if val := cmds.PersistentFlags().Lookup("config"); val != nil {
		fmt.Println("configed")
		val.Value.Set(oconfigPath)
	} else {
		fmt.Println("setting")
		val = cmds.PersistentFlags().VarPF(stringValue{oconfigPath}, "config", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] config: %+v\n", val)
	}
	if val := cmds.PersistentFlags().Lookup("loglevel"); val != nil {
		fmt.Println("configed")
		val.Value.Set("10")
	} else {
		fmt.Println("setting")
		val = cmds.PersistentFlags().VarPF(intValue{5}, "loglevel", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] loglevel: %+v\n", val)
	}
	templates.ActsAsRootCommand(cmds, []string{"option"})
	cmds.AddCommand(cmd.NewCmdOptions(os.Stdout))

	rootCommand = cmds
}

func ShowUsers() error {
	f := NewClientCmdFactory()
	config, err := f.OpenShiftClientConfig.ClientConfig()
	if err != nil {
		return err
	}
	logger.Printf("rest client config: %+v\n", config)
	config.APIPath = "/oapi"
	config.Host = "https://172.17.4.50:30448"

	oc, err := client.New(config)
	if err != nil {
		return err
	}
	logger.Printf("rest client: %+v\n", oc)

	result, err := oc.Users().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func DoBasicAuth() error {
	clientConfig := &restclient.Config{}
	serverNormalized, err := config.NormalizeServerURL("https://172.17.4.50:30448")
	if err != nil {
		return err
	}
	clientConfig.Host = serverNormalized
	clientConfig.CAFile = oca
	clientConfig.CertFile = oclientcrt
	clientConfig.KeyFile = oclientkey
	clientConfig.GroupVersion = &unversioned.GroupVersion{Group: "", Version: "v1"}
	clientConfig.APIPath = "/oapi"
	clientConfig.Codec = kapi.Codecs.LegacyCodec(*clientConfig.GroupVersion, kapi.SchemeGroupVersion)
	//clientConfig.Codec = kapi.Codecs.CodecForVersions(json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil), kapi.Codecs.UniversalDeserializer(), []unversioned.GroupVersion{*clientConfig.GroupVersion}, []unversioned.GroupVersion{*clientConfig.GroupVersion})
	//clientConfig.Codec = kapi.Codecs.CodecForVersions(json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil), []unversioned.GroupVersion{*clientConfig.GroupVersion}, []unversioned.GroupVersion{*clientConfig.GroupVersion})
	logger.Printf("simple config: %+v\n", clientConfig)

	//clientConfig.Username = "tangfeixiong"
	//clientConfig.Password = "tangfeixiong"
	//	token, err := tokencmd.RequestToken(clientConfig, os.Stdin, clientConfig.Username, clientConfig.Password)
	//	if err != nil {
	//		return err
	//	}
	//	logger.Printf("current token: %+v\n", token)
	//	clientConfig.BearerToken = token

	clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"
	clientConfig.Username = ""
	clientConfig.Password = ""

	clientK8s, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		return err
	}

	osClient := &client.Client{clientK8s}

	me, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		return err
	}

	logger.Printf("Result: %+v\n", me)

	return nil
}

func ShowSelf() error {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		return err
	}
	logger.Printf("rest client: %+v\n", oc)

	result, err := oc.Users().Get("~")
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func ShowProjects() error {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		return err
	}
	logger.Printf("rest client: %+v\n", oc)

	result, err := oc.Projects().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func Whole() (*userapi.UserList, error) {
	pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")

	var whole *userapi.UserList
	var err error
	rootCommand.Run = func(c *cobra.Command, args []string) {
		f := osclientcmd.New(pf)
		oc, _, err := f.Clients()
		if err != nil {
			whole = nil
			return
		}
		whole, err = oc.Users().List(kapi.ListOptions{})
		if err != nil {
			whole = nil
			return
		}
	}

	if err = rootCommand.Execute(); err != nil {
		return nil, err
	}
	return whole, nil
}

func ShowMe() error {
	//rootCommand.SetArgs([]string{"version"})
	if err := rootCommand.Execute(); err != nil {
		glog.V(5).Infof("Failed: %v", err)
		return err
	}
	return nil
}

func WhoAmI() (*userapi.User, error) {
	pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")

	var me *userapi.User = nil
	var err error = nil
	rootCommand.Run = func(c *cobra.Command, args []string) {
		f := osclientcmd.New(pf)
		oc, _, err := f.Clients()
		if err != nil {
			me = nil
			return
		}
		me, err = oc.Users().Get("~")
		if err != nil {
			me = nil
			return
		}
	}

	if err = rootCommand.Execute(); err != nil {
		return nil, err
	}
	return me, err
}

func LoginWithBasicAuth(username, password string) error {
	//pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")
	f := factory
	c := cmd.NewCmdLogin("oc", f, os.Stdin, os.Stdout)
	options(c)

	if val := c.Flags().Lookup("username"); val != nil {
		val.Value.Set(username)
	} else {
		val = c.Flags().VarPF(stringValue{username}, "username", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag username: %+v\n", val)
	}
	c.Flags().Lookup("username").NoOptDefVal = username

	if val := c.Flags().Lookup("password"); val != nil {
		if err := c.Flags().Set("password", password); err != nil {
			fmt.Fprintf(os.Stdout, "[tangfx] flag password err: %+v\n", err)
			return err
		}
	} else {
		val = c.Flags().VarPF(stringValue{password}, "password", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag password: %+v\n", f)
	}
	c.Flags().Lookup("password").NoOptDefVal = password

	if val := c.Flags().Lookup("token"); val != nil {
		val.Value.Set("")
	} else {
		val = c.Flags().VarPF(stringValue{""}, "token", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] token: %+v\n", val)
	}
	c.Flags().Lookup("token").NoOptDefVal = ""

	//c.SetArgs([]string{"172.17.4.50:30448"})
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not login with basic auth: %v\n", err)
		return err
	}
	return nil
}

func ProjectWithSimple(name string, in io.Reader, out, errOut io.Writer) {
	f := osclientcmd.NewFactory(withOClientConfig())
	fullName := "oc"
	c := cmd.NewCmdRequestProject(fullName, "new-project", fullName+" login", fullName+" project", f, out)
	c.SetArgs([]string{name})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not create project: %v\n", err)
	}
}

func BuildWithConfig(name string, in io.Reader, out, errOut io.Writer) {
	f := osclientcmd.NewFactory(withOClientConfig())
	c := NewCmdStartBuild(name, f, in, out)
	c.SetArgs([]string{name})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not start build with config: %v\n", err)
	}
}

// NewCmdNewBuild implements the OpenShift cli new-build command
func NewCmdNewBuild(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	config := newcmd.NewAppConfig()
	config.ExpectToBuild = true
	config.AddEnvironmentToBuild = true
	options := &cmd.NewBuildOptions{Config: config}

	cmd := &cobra.Command{
		Use:        "new-build (IMAGE | IMAGESTREAM | PATH | URL ...)",
		Short:      "Create a new build configuration",
		Long:       fmt.Sprintf(newBuildLong, fullName),
		Example:    fmt.Sprintf(newBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete(fullName, f, c, args, out, in))
			err := options.Run()
			if err == cmdutil.ErrExit {
				os.Exit(1)
			}
			kcmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringSliceVar(&config.SourceRepositories, "code", config.SourceRepositories, "Source code in the build configuration.")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image", "", config.ImageStreams, "Name of an image stream to to use as a builder. (deprecated)")
	cmd.Flags().MarkDeprecated("image", "use --image-stream instead")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image-stream", "i", config.ImageStreams, "Name of an image stream to to use as a builder.")
	cmd.Flags().StringSliceVar(&config.DockerImages, "docker-image", config.DockerImages, "Name of a Docker image to use as a builder.")
	cmd.Flags().StringSliceVar(&config.Secrets, "build-secret", config.Secrets, "Secret and destination to use as an input for the build.")
	cmd.Flags().StringVar(&config.Name, "name", "", "Set name to use for generated build artifacts.")
	cmd.Flags().StringVar(&config.To, "to", "", "Push built images to this image stream tag (or Docker image repository if --to-docker is set).")
	cmd.Flags().BoolVar(&config.OutputDocker, "to-docker", false, "Have the build output push to a Docker repository.")
	cmd.Flags().StringSliceVarP(&config.Environment, "env", "e", config.Environment, "Specify key value pairs of environment variables to set into resulting image.")
	cmd.Flags().StringVar(&config.Strategy, "strategy", "", "Specify the build strategy to use if you don't want to detect (docker|source).")
	cmd.Flags().StringVarP(&config.Dockerfile, "dockerfile", "D", "", "Specify the contents of a Dockerfile to build directly, implies --strategy=docker. Pass '-' to read from STDIN.")
	cmd.Flags().BoolVar(&config.BinaryBuild, "binary", false, "Instead of expecting a source URL, set the build to expect binary contents. Will disable triggers.")
	cmd.Flags().StringP("labels", "l", "", "Label to set in all generated resources.")
	cmd.Flags().BoolVar(&config.AllowMissingImages, "allow-missing-images", false, "If true, indicates that referenced Docker images that cannot be found locally or in a registry should still be used.")
	cmd.Flags().BoolVar(&config.AllowMissingImageStreamTags, "allow-missing-imagestream-tags", false, "If true, indicates that image stream tags that don't exist should still be used.")
	cmd.Flags().StringVar(&config.ContextDir, "context-dir", "", "Context directory to be used for the build.")
	cmd.Flags().BoolVar(&config.DryRun, "dry-run", false, "If true, do not actually create resources.")
	cmd.Flags().BoolVar(&config.NoOutput, "no-output", false, "If true, the build output will not be pushed anywhere.")
	cmd.Flags().StringVar(&config.SourceImage, "source-image", "", "Specify an image to use as source for the build.  You must also specify --source-image-path.")
	cmd.Flags().StringVar(&config.SourceImagePath, "source-image-path", "", "Specify the file or directory to copy from the source image and its destination in the build directory. Format: [source]:[destination-dir].")
	kcmdutil.AddPrinterFlags(cmd)

	return cmd
}

func NewCmdStartBuild(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	o := &cmd.StartBuildOptions{
		LogLevel:        "5",
		Follow:          true,
		WaitForComplete: false,
	}

	cmd := &cobra.Command{
		Use:        "start-build (BUILDCONFIG | --from-build=BUILD)",
		Short:      "Start a new build",
		Long:       startBuildLong,
		Example:    fmt.Sprintf(startBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, in, out, cmd, args))
			kcmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVar(&o.LogLevel, "build-loglevel", o.LogLevel, "Specify the log level for the build log output")
	cmd.Flags().Lookup("build-loglevel").NoOptDefVal = "5"
	cmd.Flags().StringSliceVarP(&o.Env, "env", "e", o.Env, "Specify key value pairs of environment variables to set for the build container.")

	cmd.Flags().StringVar(&o.FromBuild, "from-build", o.FromBuild, "Specify the name of a build which should be re-run")

	cmd.Flags().BoolVar(&o.Follow, "follow", o.Follow, "Start a build and watch its logs until it completes or fails")
	cmd.Flags().Lookup("follow").NoOptDefVal = "true"
	cmd.Flags().BoolVar(&o.WaitForComplete, "wait", o.WaitForComplete, "Wait for a build to complete and exit with a non-zero return code if the build fails")

	cmd.Flags().StringVar(&o.FromFile, "from-file", o.FromFile, "A file to use as the binary input for the build; example a pom.xml or Dockerfile. Will be the only file in the build source.")
	cmd.Flags().StringVar(&o.FromDir, "from-dir", o.FromDir, "A directory to archive and use as the binary input for a build.")
	cmd.Flags().StringVar(&o.FromRepo, "from-repo", o.FromRepo, "The path to a local source code repository to use as the binary input for a build.")
	cmd.Flags().StringVar(&o.Commit, "commit", o.Commit, "Specify the source code commit identifier the build should use; requires a build based on a Git repository")

	cmd.Flags().StringVar(&o.ListWebhooks, "list-webhooks", o.ListWebhooks, "List the webhooks for the specified build config or build; accepts 'all', 'generic', or 'github'")
	cmd.Flags().StringVar(&o.FromWebhook, "from-webhook", o.FromWebhook, "Specify a webhook URL for an existing build config to trigger")

	cmd.Flags().StringVar(&o.GitPostReceive, "git-post-receive", o.GitPostReceive, "The contents of the post-receive hook to trigger a build")
	cmd.Flags().StringVar(&o.GitRepository, "git-repository", o.GitRepository, "The path to the git repository for post-receive; defaults to the current directory")

	// cmdutil.AddOutputFlagsForMutation(cmd)
	return cmd

}

/*
   k8s.io/kubernetes/pkg/client/unversioned/clientcmd/client_config.go
*/
// inClusterClientConfig makes a config that will work from within a kubernetes cluster container environment.
type inClusterClientConfig struct{}

func (inClusterClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, fmt.Errorf("inCluster environment config doesn't support multiple clusters")
}

func (inClusterClientConfig) ClientConfig() (*restclient.Config, error) {
	return restclient.InClusterConfig()
}

func (inClusterClientConfig) Namespace() (string, error) {
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
	}

	return "default", nil
}

// Possible returns true if loading an inside-kubernetes-cluster is possible.
func (inClusterClientConfig) Possible() bool {
	fi, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
		os.Getenv("KUBERNETES_SERVICE_PORT") != "" &&
		err == nil && !fi.IsDir()
}

// makeUserIdentificationFieldsConfig returns a client.Config capable of being merged using mergo for only user identification information
func makeUserIdentificationConfig(info clientauth.Info) *restclient.Config {
	config := &restclient.Config{}
	config.Username = info.User
	config.Password = info.Password
	config.CertFile = info.CertFile
	config.KeyFile = info.KeyFile
	config.BearerToken = info.BearerToken
	return config
}

// makeUserIdentificationFieldsConfig returns a client.Config capable of being merged using mergo for only server identification information
func makeServerIdentificationConfig(info clientauth.Info) restclient.Config {
	config := restclient.Config{}
	config.CAFile = info.CAFile
	if info.Insecure != nil {
		config.Insecure = *info.Insecure
	}
	return config
}

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
func withKClientConfig() clientcmd.ClientConfig {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		glog.Infof("kubeconfig not found: %v\n", err)
		os.Exit(1)
	}
	glog.Infof("kubeconfig: \n%+v\n", string(data))

	conf, err := clientcmd.Load(data)
	//conf, err := kubectlcmdcfg.NewDefaultPathOptions().GetStartingConfig()
	//conf, err := clientcmdapi.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		glog.Infof("cmd client not configured: %v\n", err)
		os.Exit(1)
	}
	glog.Infof("cmd client config: %+v\n", conf)

	kClientConfig := clientcmd.NewNonInteractiveClientConfig(*conf, kubeconfigContext, &clientcmd.ConfigOverrides{})
	glog.Infof("rest client config: %+v\n", kClientConfig)
	return kClientConfig
}

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
func withOClientConfig() clientcmd.ClientConfig {
	conf, err := clientcmd.LoadFromFile(oconfigPath)
	if err != nil {
		logger.Printf("openshift cmd api client not configured: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("openshift cmd api cmd client config: %+v\n", conf)

	oClientConfig := clientcmd.NewNonInteractiveClientConfig(*conf, oconfigContext, &clientcmd.ConfigOverrides{})
	logger.Printf("rest client config: %+v\n", oClientConfig)
	return oClientConfig
}

// openshift/origin/pkg/cmd/server/api/helpers.go
func withAdminConfig() {
	if kClient, kConfig, err := configapi.GetKubeClient(kubeconfigPath); err != nil {
		logger.Printf("Could not get kubernetes admin client: %+v\n", err)
	} else if kClient == nil || kConfig == nil {
		logger.Println("Could not find kubernetes admin client\n")
	} else {
		logger.Printf("Kubernetes admin client %v with config %+v", kClient, kConfig)
	}

	if oClient, oConfig, err := configapi.GetOpenShiftClient(oconfigPath); err != nil {
		logger.Printf("Could not get openshift admin client: %+v\n", err)
	} else if oClient == nil || oConfig == nil {
		logger.Println("Could not find openshift admin client\n")
	} else {
		logger.Printf("Openshift admin client %v with config %+v", oClient, oConfig)
	}
}
