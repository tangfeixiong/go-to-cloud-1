package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"

	oclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
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
	logger.Printf("factory client config: %+v\n", cfg)
	if err == nil {
		return cfg, nil
	}

	if !kclientcmd.IsEmptyConfig(err) {
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

	kClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf, kubeconfigContext, &kclientcmd.ConfigOverrides{})
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

	oClientConfig := kclientcmd.NewNonInteractiveClientConfig(*conf, oconfigContext, &kclientcmd.ConfigOverrides{})

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
