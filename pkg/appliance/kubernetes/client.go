package kubernetes

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/cliconfig"
	"github.com/docker/engine-api/types"
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/docker"
)

var (
	_kubeconfig     string = "/data/src/github.com/openshift/origin/etc/kubeconfig"
	_kube_context   string = "openshift-origin-single"
	_kube_apiserver string = "https://172.17.4.50"
)

type Orchestration struct {
	apiserver string
}

func NewOrchestration() *Orchestration {
	if v, ok := os.LookupEnv("KUBE_CONFIG"); ok && len(v) > 0 {
		_kubeconfig = v
	}
	if v, ok := os.LookupEnv("KUBE_CONTEXT"); ok && len(v) > 0 {
		_kube_context = v
	} else {
		_kube_context = ""
	}
	if v, ok := os.LookupEnv("KUBERNETES_MASTER"); ok && len(v) > 0 {
		_kube_apiserver = v
	} else {
		_kube_apiserver = ""
	}
	return &Orchestration{}
}

func (*Orchestration) VerifyDockerConfigJsonSecretAndServiceAccount(namespace, secret string,
	dac types.AuthConfig, sa string) (*cliconfig.ConfigFile, *kapi.Secret, *kapi.ServiceAccount, error) {
	c, _, err := withKubeconfig(_kubeconfig, _kube_context, _kube_apiserver)
	if err != nil {
		return nil, nil, nil, err
	}
	return verifyDockerConfigJsonSecretAndServiceAccount(c, namespace, secret, dac, sa)
}

func withKubeconfig(kubeconfig, context, server string) (c *kclient.Client, cc kclientcmd.ClientConfig, e error) {
	if len(kubeconfig) == 0 {
		var conf *kclientcmdapi.Config
		var err error
		conf, err = kclientcmd.NewDefaultPathOptions().GetStartingConfig()
		if err != nil {
			glog.Errorf("Failed to access default kubeconfig: %v\n", err)
			return nil, nil, err
		}
		glog.V(10).Infof("kubeconfig: %+v\n", conf)

		if len(server) == 0 {
			cc = kclientcmd.NewNonInteractiveClientConfig(*conf, context,
				&kclientcmd.ConfigOverrides{ClusterInfo: kclientcmd.EnvVarCluster},
				kclientcmd.NewDefaultClientConfigLoadingRules())
		} else {
			cc = kclientcmd.NewNonInteractiveClientConfig(*conf, context,
				&kclientcmd.ConfigOverrides{
					ClusterInfo: kclientcmdapi.Cluster{
						Server: server,
					},
				},
				kclientcmd.NewDefaultClientConfigLoadingRules())
		}
	} else {
		data, err := ioutil.ReadFile(kubeconfig)
		if err != nil {
			glog.Errorf("kubeconfig not accessed: %+v\n", err)
			return nil, nil, err
		}

		conf, err := kclientcmd.Load(data)
		if err != nil {
			glog.Errorf("kubeconfig invalid: %v\n", err)
			return nil, nil, err
		}
		glog.V(10).Infof("kubeconfig: \n%+v\n", string(data))

		cc = kclientcmd.NewNonInteractiveClientConfig(*conf, context,
			&kclientcmd.ConfigOverrides{
				ClusterInfo: kclientcmdapi.Cluster{
					Server: server,
				},
			},
			kclientcmd.NewDefaultClientConfigLoadingRules())
	}

	restconf, err := cc.ClientConfig()
	if err != nil {
		glog.Errorf("Could not validate rest config: %+v", err)
		return nil, cc, err
	}
	c, e = kclient.New(restconf)
	if e != nil {
		glog.Errorf("Could not configure Kubernetes client: %s", e)
		return nil, cc, e
	}
	glog.V(10).Infof("Setup client (%+v) with config (%+v)\n", c, cc)
	return
}

func createSecretForDockerConfigJson(c *kclient.Client,
	namespace, secret string, ac *types.AuthConfig) (*cliconfig.ConfigFile, *kapi.Secret, error) {
	b, dcf, err := docker.SerializeIntoConfigFile(ac)
	if err != nil {
		return nil, nil, err
	}
	obj := &kapiv1.Secret{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: kapiv1.ObjectMeta{
			Name:      secret,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			string(kapi.DockerConfigJsonKey): b,
		},
		Type: kapiv1.SecretType(string(kapi.SecretTypeDockerConfigJson)),
	}

	_, result, _, err := createSecret(c, namespace, obj)
	if err != nil {
		return dcf, nil, err
	}
	return dcf, result, nil
}

func retrieveSecretWithDockerConfigJson(c *kclient.Client, namespace, secret string) (*cliconfig.ConfigFile, *kapi.Secret, error) {
	obj, err := retrieveSecret(c, namespace, secret)
	if err != nil {
		return nil, nil, err
	}
	if obj == nil {
		return nil, nil, nil
	}

	val, ok := obj.Data[kapi.DockerConfigJsonKey]
	if !ok {
		return nil, obj, nil
	}

	hco, err := codec.JSON.Decode(val).One()
	if err != nil {
		glog.Errorf("Could not setup helm codec: %+v\n", err)
		return nil, obj, err
	}
	dcf := &cliconfig.ConfigFile{}
	if err := hco.Object(dcf); err != nil {
		glog.Errorf("Could not decode into docker config file: %+v\n", err)
		return nil, obj, err
	}
	return dcf, obj, nil
}

func updateSecretWithDockerConfigJson(c *kclient.Client,
	namespace string, secret *kapi.Secret, ac *types.AuthConfig) (*cliconfig.ConfigFile, *kapi.Secret, error) {
	b, dcf, err := docker.SerializeIntoConfigFile(ac)
	if err != nil {
		return nil, nil, err
	}
	//		if secret.Data == nil {
	//			secret.Data == make(map[string][]byte)
	//		}
	secret.Data[string(kapi.DockerConfigJsonKey)] = b
	result, err := updateSecret(c, namespace, secret)
	if err != nil {
		return dcf, nil, err
	}
	return dcf, result, nil
}

func linkSecretWithServiceAccount(c *kclient.Client, namespace, serviceaccount, secret string) (*kapi.ServiceAccount, error) {
	_, sa, v1, err := retrieveServiceAccount(c, namespace, serviceaccount)
	if err != nil {
		return nil, err
	}
	if v1 == nil {
		v1 = &kapiv1.ServiceAccount{
			TypeMeta: unversioned.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: kapiv1.ObjectMeta{
				Name:      serviceaccount,
				Namespace: namespace,
			},
			Secrets: []kapiv1.ObjectReference{
				{Name: secret},
			},
			ImagePullSecrets: []kapiv1.LocalObjectReference{
				{secret},
			},
		}
		_, sa, v1, err = createServiceAccount(c, namespace, v1)
		if err != nil {
			return nil, err
		}
		return sa, nil
	}
	if sa.Secrets == nil {
		sa.Secrets = make([]kapi.ObjectReference, 0)
	} else {
		for _, ele := range sa.Secrets {
			if ele.Name == secret {
				return sa, nil
			}
		}
	}
	sa.Secrets = append(sa.Secrets, kapi.ObjectReference{Name: secret})

	var obj *kapi.LocalObjectReference
	if sa.ImagePullSecrets == nil {
		sa.ImagePullSecrets = make([]kapi.LocalObjectReference, 0)
	} else {
		for _, ele := range sa.ImagePullSecrets {
			if ele.Name == secret {
				obj = &ele
				break
			}
		}
	}
	if obj == nil {
		sa.ImagePullSecrets = append(sa.ImagePullSecrets, kapi.LocalObjectReference{Name: secret})

	}
	sa, err = updateServiceAccount(c, namespace, sa)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func verifyDockerConfigJsonSecretAndServiceAccount(c *kclient.Client, namespace, secret string,
	dac types.AuthConfig, serviceaccount string) (*cliconfig.ConfigFile, *kapi.Secret, *kapi.ServiceAccount, error) {
	dcf, sec, err := retrieveSecretWithDockerConfigJson(c, namespace, secret)
	if err != nil {
		return nil, nil, nil, err
	}

	if sec == nil {
		dcf, sec, err = createSecretForDockerConfigJson(c, namespace, secret, &dac)
		if err != nil {
			return dcf, nil, nil, err
		}
		sa, err := linkSecretWithServiceAccount(c, namespace, serviceaccount, secret)
		if err != nil {
			return dcf, sec, nil, err
		}
		return dcf, sec, sa, nil
	}

	if dcf == nil {
		dcf, sec, err = updateSecretWithDockerConfigJson(c, namespace, sec, &dac)
		if err != nil {
			return dcf, sec, nil, err
		}
		sa, err := linkSecretWithServiceAccount(c, namespace, serviceaccount, secret)
		if err != nil {
			return dcf, sec, nil, err
		}
		return dcf, sec, sa, nil
	}

	basicauth := fmt.Sprintf("%s:%s", dac.Username, dac.Password)
	auth := base64.StdEncoding.EncodeToString([]byte(basicauth))
	//make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	//base64.StdEncoding.Encode(bEnc, b)
	if len(dac.Auth) == 0 || strings.Compare(dac.Auth, auth) != 0 {
		dac.Auth = auth
	}

	val, ok := dcf.AuthConfigs[dac.ServerAddress]
	if !ok || strings.Compare(val.Auth, dac.Auth) != 0 {
		dcf.AuthConfigs[dac.ServerAddress] = cliconfig.AuthConfig{
			Auth: dac.Auth,
		}
		b, err := docker.SerializeConfigFile(dcf)
		if err != nil {
			return dcf, sec, nil, err
		}
		sec.Data[string(kapi.DockerConfigJsonKey)] = b
		sec, err = updateSecret(c, namespace, sec)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	sa, err := linkSecretWithServiceAccount(c, namespace, serviceaccount, secret)
	if err != nil {
		return dcf, sec, nil, err
	}
	return dcf, sec, sa, nil
}

// PrintObject prints an api object given command line flags to modify the output format
func PrintObject(f *util.Factory, mapper meta.RESTMapper, obj runtime.Object, out io.Writer,
	outputversion, template, sortby string, noheaders, wide, showall, showlabels, iswatch bool,
	labelcolumns []string) error {
	gvks, _, err := kapi.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}

	mapping, err := mapper.RESTMapping(gvks[0].GroupKind())
	if err != nil {
		return err
	}

	printer, err := PrinterForMapping(f, mapping, false, outputversion, template, sortby, noheaders, wide, showall, showlabels, iswatch, labelcolumns)
	if err != nil {
		return err
	}
	return printer.PrintObj(obj, out)
}

// PrinterForMapping returns a printer suitable for displaying the provided resource type.
// Requires that printer flags have been added to cmd (see AddPrinterFlags).
func PrinterForMapping(f *util.Factory, mapping *meta.RESTMapping, withNamespace bool,
	outputversion, template, sortby string, noheaders, wide, showall, showlabels, iswatch bool,
	labelcolumns []string) (kubectl.ResourcePrinter, error) {
	printer, ok, err := PrinterForCommand(outputversion, template, sortby)
	if err != nil {
		return nil, err
	}
	if ok {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return nil, err
		}

		version, err := OutputVersion(outputversion, clientConfig.GroupVersion)
		if err != nil {
			return nil, err
		}
		if version.IsEmpty() && mapping != nil {
			version = mapping.GroupVersionKind.GroupVersion()
		}
		if version.IsEmpty() {
			return nil, fmt.Errorf("you must specify an output-version when using this output format")
		}

		if mapping != nil {
			printer = kubectl.NewVersionedPrinter(printer, mapping.ObjectConvertor, version, mapping.GroupVersionKind.GroupVersion())
		}

	} else {
		// Some callers do not have "label-columns" so we can't use the GetFlagStringSlice() helper
		/*columnLabel, err := cmd.Flags().GetStringSlice("label-columns")
		if err != nil {
			columnLabel = []string{}
		}*/

		printer, err = f.Printer(mapping, &kubectl.PrintOptions{
			NoHeaders:          noheaders, /*GetFlagBool(cmd, "no-headers")*/
			WithNamespace:      withNamespace,
			Wide:               wide,         /*GetWideFlag(cmd)*/
			ShowAll:            showall,      /*GetFlagBool(cmd, "show-all")*/
			ShowLabels:         showlabels,   /*GetFlagBool(cmd, "show-labels")*/
			AbsoluteTimestamps: iswatch,      /*isWatch(cmd)*/
			ColumnLabels:       labelcolumns, /*columnLabel*/
		})
		if err != nil {
			return nil, err
		}
		printer = MaybeWrapSortingPrinter(sortby, printer)
	}

	return printer, nil
}
