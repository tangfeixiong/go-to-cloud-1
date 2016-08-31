package e2e

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	kclientcmd "k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	kclientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

func clientWithKubeconfig(kubeconfig, context, server string) (c *kclient.Client, cc kclientcmd.ClientConfig, e error) {
	logger.SetPrefix("[client/e2e, clientWithKubeconfig] ")

	if len(kubeconfig) == 0 {
		var conf *kclientcmdapi.Config
		var err error
		if v, ok := os.LookupEnv("KUBECONFIG"); ok && len(v) > 0 {
			conf, err = kclientcmd.LoadFromFile(v)
			if err != nil {
				glog.Errorf("kubeconfig not ready: %v\n", err)
				return nil, nil, err
			}
		} else {
			conf, err = kclientcmd.NewDefaultPathOptions().GetStartingConfig()
			if err != nil {
				glog.Errorf("kubeconfig not ready: %v\n", err)
				return nil, nil, err
			}
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
			glog.Errorf("kubeconfig not ready: %+v\n", err)
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
	logger.Printf("Setup client (%+v) with config (%+v)\n", c, cc)
	return
}

func createSecret(c *kclient.Client, namespace string, secret *kapiv1.Secret) (*kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, createSecret] ")

	data, err := c.RESTClient.Verb("POST").Namespace(namespace).Resource("Secrets").Body(secret).DoRaw()
	if err != nil {
		glog.Errorf("Could not access kubernetes: %+v", err)
		return nil, err
	}

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		logger.Printf("Could not setup helm decoder: %+v\n", err)
		return nil, err
	}
	result := &kapi.Secret{}
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into kube secret: %+v\n", err)
		return nil, err
	}

	return result, nil
}

func retrieveSecret(c *kclient.Client, namespace, secret string) (*kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, retrieveSecret] ")

	data, err := c.RESTClient.Verb("GET").Namespace(namespace).Resource("Secrets").Name(secret).DoRaw()
	if err != nil {
		glog.Errorf("Could not access kubernetes: %+v", err)
		return nil, err
	}

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		logger.Printf("Could not setup helm decoder: %+v\n", err)
		return nil, err
	}
	meta := &unversioned.TypeMeta{}
	if err := hco.Object(meta); err != nil {
		logger.Printf("Could not decode into kube TypeMeta: %+v\n", err)
		return nil, err
	}
	if strings.EqualFold("Status", meta.Kind) {
		return nil, nil
	}

	obj := &kapi.Secret{}
	if err := hco.Object(obj); err != nil {
		logger.Printf("Could not decode into kube secret: %+v\n", err)
		return nil, err
	}
	return obj, nil
}

func updateSecret(c *kclient.Client, namespace string, obj *kapi.Secret) (*kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, updateSecret] ")

	result, err := c.Secrets(namespace).Update(obj)
	if err != nil {
		glog.Errorf("Could not update secret: %+v\n", err)
		return nil, err
	}
	logger.Printf("Secret updated: %+v\n", result)
	return result, nil
}

func deleteSecret(c *kclient.Client, namespace, secret string) error {
	logger.SetPrefix("[client/e2e, deleteSecret] ")

	err := c.Secrets(namespace).Delete(secret)
	if err != nil {
		glog.Errorf("Could not delete secret: %+v\n", err)
		return err
	}
	logger.Println("Secret deleted")
	return nil
}

func retrieveServiceAccount(c *kclient.Client, namespace, serviceaccount string) (*kapi.ServiceAccount, error) {
	logger.SetPrefix("[client/e2e, retrieveServiceAccount] ")

	data, err := c.RESTClient.Get().Namespace(namespace).Resource("ServiceAccounts").Name(serviceaccount).DoRaw()
	if err != nil {
		glog.Errorf("Could nout access kube sa: %+v\n", err)
		return nil, err
	}

	hco, err := codec.JSON.Decode(data).One()
	if err != nil {
		logger.Printf("Could not setup helm decoder: %+v\n", err)
		return nil, err
	}
	meta := &unversioned.TypeMeta{}
	if err := hco.Object(meta); err != nil {
		logger.Printf("Could not decode into kube TypeMeta: %+v\n", err)
		return nil, err
	}
	if strings.EqualFold("Status", meta.Kind) {
		return nil, err
	}

	obj := &kapi.ServiceAccount{}
	if err := hco.Object(obj); err != nil {
		logger.Printf("Could not decode into kube serviceaccount: %+v\n", err)
		return nil, err
	}
	return obj, nil
}

func updateServiceAccount(c *kclient.Client, namespace string, obj *kapi.ServiceAccount) (*kapi.ServiceAccount, error) {
	logger.SetPrefix("[client/e2e, updateServiceAccount] ")

	result, err := c.ServiceAccounts(namespace).Update(obj)
	if err != nil {
		glog.Errorf("Could not update serviceaccount: %+v\n", err)
		return nil, err
	}
	logger.Printf("ServiceAccount updated: %+v\n", result)
	return result, nil
}

func createSecretWithDockerConfigJson(c *kclient.Client,
	namespace, secret string, auth *DockerAuthConfig) (*DockerConfigFile, *kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, createSecretWithDockerConfigJson] ")

	b, dcf, err := SerializeIntoDockerConfigFile(auth)
	if err != nil {
		logger.Printf("Could not decode docker config json: %+v\n", err)
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

	result, err := createSecret(c, namespace, obj)
	if err != nil {
		return dcf, nil, err
	}
	return dcf, result, nil
}

func retrieveSecretWithDockerConfigJson(c *kclient.Client, namespace, secret string) (*DockerConfigFile, *kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, retrieveSecretWithDockerConfigJson] ")

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
		logger.Printf("Could not setup helm decoder: %+v\n", err)
		return nil, obj, err
	}
	dcf := &DockerConfigFile{}
	if err := hco.Object(dcf); err != nil {
		logger.Printf("Could not decode into docker config file: %+v\n", err)
		return nil, obj, err
	}
	return dcf, obj, nil
}

func updateSecretWithDockerConfigJson(c *kclient.Client,
	namespace string, secret *kapi.Secret, auth *DockerAuthConfig) (*DockerConfigFile, *kapi.Secret, error) {
	logger.SetPrefix("[client/e2e, updateSecretWithDockerConfigJso] ")

	b, dcf, err := SerializeIntoDockerConfigFile(auth)
	if err != nil {
		logger.Printf("Could not decode docker config json: %+v\n", err)
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
	logger.SetPrefix("[client/e2e, verifyServiceAccountSecret] ")

	sa, err := retrieveServiceAccount(c, namespace, serviceaccount)
	if err != nil {
		return nil, err
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

	//	if sa.ImagePullSecrets == nil {
	//		sa.ImagePullSecrets = make([]kapi.LocalObjectReference, 0)
	//	} else {
	//	    for _, ele := range sa.ImagePullSecrets {
	//		    if ele.Name == secret {
	//			    return secret, nil
	//		    }
	//      }
	//	}
	//	sa.ImagePullSecrets = append(sa.ImagePullSecrets, kapi.LocalObjectReference{Name: secret})

	sa, err = updateServiceAccount(c, namespace, sa)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func verifyDockerConfigJsonSecretAndServiceAccount(c *kclient.Client, namespace, secret string,
	dockerAuth DockerAuthConfig, serviceaccount string) (*DockerConfigFile, *kapi.Secret, *kapi.ServiceAccount, error) {
	logger.SetPrefix("[client/e2e, verifyDockerConfigJsonSecret] ")

	dcf, sec, err := retrieveSecretWithDockerConfigJson(c, namespace, secret)
	if err != nil {
		return nil, nil, nil, err
	}

	if sec == nil {
		dcf, sec, err = createSecretWithDockerConfigJson(c, namespace, secret, &dockerAuth)
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
		dcf, sec, err = updateSecretWithDockerConfigJson(c, namespace, sec, &dockerAuth)
		if err != nil {
			return dcf, sec, nil, err
		}
		sa, err := linkSecretWithServiceAccount(c, namespace, serviceaccount, secret)
		if err != nil {
			return dcf, sec, nil, err
		}
		return dcf, sec, sa, nil
	}

	basicauth := fmt.Sprintf("%s:%s", dockerAuth.Username, dockerAuth.Password)
	dockerAuth.Auth = base64.StdEncoding.EncodeToString([]byte(basicauth))
	//make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	//base64.StdEncoding.Encode(bEnc, b)

	val, ok := dcf.AuthConfigs[dockerAuth.ServerAddress]
	if !ok || strings.Compare(val.Auth, dockerAuth.Auth) != 0 {
		dcf.AuthConfigs[dockerAuth.ServerAddress] = DockerAuthConfig{
			Auth: dockerAuth.Auth,
		}
		b, err := SerializeDockerConfigFile(dcf)
		if err != nil {
			logger.Printf("Could not decode docker config json: %+v\n", err)
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
