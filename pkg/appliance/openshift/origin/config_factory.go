package origin

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"

	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	oclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	clientauth "k8s.io/kubernetes/pkg/client/unversioned/auth"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

type defaultingClientConfig struct {
	*kclientcmd.DirectClientConfig
	nested kclientcmd.ClientConfig
}

func (c defaultingClientConfig) RawConfig() (kclientcmdapi.Config, error) {
	return c.nested.RawConfig()
}

func (c defaultingClientConfig) Namespace() (string, bool, error) {
	namespace, ok, err := c.nested.Namespace()
	if err == nil {
		return namespace, ok, nil
	}
	if !kclientcmd.IsEmptyConfig(err) {
		return "", false, err
	}

	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, true, nil
	}

	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, true, nil
		}
	}

	return api.NamespaceDefault, false, nil
}

func (c defaultingClientConfig) ClientConfig() (*restclient.Config, error) {
	cfg, err := c.nested.ClientConfig()
	if err == nil {
		return cfg, nil
	}

	if !kclientcmd.IsEmptyConfig(err) {
		logger.Printf("Invalid client config from factory: %+v\n", cfg)
		return nil, err
	}

	// TODO: need to expose inClusterConfig upstream and use that
	if icc, err := restclient.InClusterConfig(); err == nil {
		glog.V(4).Infof("Using in-cluster configuration")
		return icc, nil
	}

	return nil, fmt.Errorf(`No configuration file found, please login or point to an existing file:

  1. Via the command-line flag --config
  2. Via the KUBECONFIG environment variable
  3. In your home directory as ~/.kube/config

To view or setup config directly use the 'config' command.`)
}

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
func directKClientConfig() *kclientcmd.DirectClientConfig {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		logger.Printf("kubeconfig not found: %v\n", err)
		return nil
	}
	logger.Printf("kubeconfig: \n%+v\n", string(data))

	conf, err := kclientcmd.Load(data)
	//conf, err := kubectlcmdcfg.NewDefaultPathOptions().GetStartingConfig()
	//conf, err := clientcmdapi.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		logger.Printf("cmd client not configured: %v\n", err)
		return nil
	}
	logger.Printf("cmd client config: %+v\n", conf)

	kClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf,
		kubeconfigContext, &kclientcmd.ConfigOverrides{},
		kclientcmd.NewDefaultClientConfigLoadingRules())
	logger.Printf("rest kclient config: %+v\n", kClientConfig)
	return kClientConfig.(*kclientcmd.DirectClientConfig)
}

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
func directOClientConfig() kclientcmd.ClientConfig {
	conf, err := kclientcmd.LoadFromFile(oconfigPath)
	if err != nil {
		logger.Printf("openshift cmd api client not configured: %v\n", err)
		return nil
	}
	logger.Printf("openshift cmd api cmd client config: %+v\n", conf)

	oClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf,
		oconfigContext,
		&kclientcmd.ConfigOverrides{},
		kclientcmd.NewDefaultClientConfigLoadingRules())

	logger.Printf("rest oclient config: %+v\n", oClientConfig)
	return oClientConfig
}

func NewClientCmdFactory() *oclientcmd.Factory {
	oClientConfig := directOClientConfig()
	kClientConfig := directKClientConfig()
	clientConfig := defaultingClientConfig{DirectClientConfig: kClientConfig, nested: oClientConfig}
	f := oclientcmd.NewFactory(clientConfig)
	return f
}

/*
   k8s.io/kubernetes/pkg/client/unversioned/clientcmd/client_config.go
*/
// inClusterClientConfig makes a config that will work from within a kubernetes cluster container environment.
type inClusterClientConfig struct{}

func (inClusterClientConfig) RawConfig() (kclientcmdapi.Config, error) {
	return kclientcmdapi.Config{}, fmt.Errorf("inCluster environment config doesn't support multiple clusters")
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
func withKClientConfig() kclientcmd.ClientConfig {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		glog.Infof("kubeconfig not found: %v\n", err)
		os.Exit(1)
	}
	glog.Infof("kubeconfig: \n%+v\n", string(data))

	conf, err := kclientcmd.Load(data)
	//conf, err := kubectlcmdcfg.NewDefaultPathOptions().GetStartingConfig()
	//conf, err := clientcmdapi.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		glog.Infof("cmd client not configured: %v\n", err)
		os.Exit(1)
	}
	glog.Infof("cmd client config: %+v\n", conf)

	kClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf,
		kubeconfigContext,
		&kclientcmd.ConfigOverrides{},
		kclientcmd.NewDefaultClientConfigLoadingRules())
	glog.Infof("rest client config: %+v\n", kClientConfig)
	return kClientConfig
}

// k8s.io/kubernetes/pkg/client/unversioned/clientcmd/loader.go
func withOClientConfig() kclientcmd.ClientConfig {
	conf, err := kclientcmd.LoadFromFile(oconfigPath)
	if err != nil {
		logger.Printf("openshift cmd api client not configured: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("openshift cmd api cmd client config: %+v\n", conf)

	oClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf,
		oconfigContext,
		&kclientcmd.ConfigOverrides{},
		kclientcmd.NewDefaultClientConfigLoadingRules())
	logger.Printf("rest client config: %+v\n", oClientConfig)
	return oClientConfig
}

// openshift/origin/pkg/cmd/server/api/helpers.go
func withAdminConfig() {
	var overrides *configapi.ClientConnectionOverrides = &configapi.ClientConnectionOverrides{
		AcceptContentTypes: "application/json",
		ContentType:        "application/json",
		QPS:                2.0,
		Burst:              10,
	}
	//configapi.SetProtobufClientDefaults(overrides)
	if kClient, kConfig, err := configapi.GetKubeClient(kubeconfigPath, overrides); err != nil {
		logger.Printf("Could not get kubernetes admin client: %+v\n", err)
	} else if kClient == nil || kConfig == nil {
		logger.Println("Could not find kubernetes admin client\n")
	} else {
		logger.Printf("Kubernetes admin client %v with config %+v", kClient, kConfig)
	}

	if oClient, oConfig, err := configapi.GetOpenShiftClient(oconfigPath, overrides); err != nil {
		logger.Printf("Could not get openshift admin client: %+v\n", err)
	} else if oClient == nil || oConfig == nil {
		logger.Println("Could not find openshift admin client\n")
	} else {
		logger.Printf("Openshift admin client %v with config %+v", oClient, oConfig)
	}
}
