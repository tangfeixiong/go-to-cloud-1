package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"time"

	_ "github.com/ghodss/yaml"

	"k8s.io/helm/pkg/kube"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	// clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	// kubectlcmdcfg "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	"k8s.io/kubernetes/pkg/runtime/serializer/json"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

	kubeconfig  string = "/data/src/github.com/openshift/origin/openshift.local.config/master/kubeconfig"
	contextname string = "openshift-origin-single"
	namespace   string = "default"
)

func main() {
	data, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		logger.Printf("kubeconfig not found: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("kubeconfig: \n%+v\n", string(data))

	conf, err := clientcmd.Load(data)
	//conf, err := kubectlcmdcfg.NewDefaultPathOptions().GetStartingConfig()
	//conf, err := clientcmdapi.NewDefaultPathOptions().GetStartingConfig()
	if err != nil {
		logger.Printf("cmd client not configured: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("cmd client config: %+v\n", conf)

	clientconfig, err := clientcmd.NewDefaultClientConfig(*conf, &clientcmd.ConfigOverrides{}).ClientConfig()

	if err != nil {
		logger.Printf("restclient not configured: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("restclient: %+v\n", clientconfig)

	logger.Println("Creating with helm")

	cc := clientcmd.NewNonInteractiveClientConfig(*conf, contextname, &clientcmd.ConfigOverrides{})
	helmClient := kube.New(cc)

	if err := helmClient.Create(namespace, bytes.NewBufferString(`
apiVersion: v1
kind: Pod
metadata:
  annotations:
    developer: |
      {"author":{"name":"tangfeixiong","email":"tangfx128@gmail.com","description":"kubernetes+helm+openshift"}}
  creationTimestamp: 2016-05-22T10:27:29Z
  labels:
    app: netcat-httpserver-hello
    name: netcat-httpserver-hello
    run: netcat-httpserver-hello
  name: netcat-httpserver-hello
  namespace: default
spec:
  containers:
  - image: quay.io/tangfeixiong/netcat-http-server-simple
    imagePullPolicy: IfNotPresent
    name: netcat-httpserver-hello
    ports:
    - containerPort: 80
      protocol: TCP
  restartPolicy: Always
status: {}
    `)); err != nil {
		logger.Printf("Could not deploy pod: %v\n", err)
		os.Exit(1)
	}

	logger.Println("Going to get after 5 seconds")
	time.Sleep(5 * time.Second)

	restcc, err := cc.ClientConfig()
	if err != nil {
		logger.Printf("restclient not configured: %v\n", err)
		os.Exit(1)
	}

	restclient, err := unversioned.New(restcc)
	if err != nil {
		logger.Printf("restclient not created: %v\n", err)
		os.Exit(1)
	}

	obj, err := restclient.Pods(namespace).Get("netcat-httpserver-hello")
	if err != nil {
		logger.Printf("Could not get pod: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("pod: %+v\n", obj)

	logger.Println("Deleting with helm")
	buf := &bytes.Buffer{}
	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, true)
	encoder := api.Codecs.EncoderForVersion(serializer, api.Unversioned)
	if err := encoder.EncodeToStream(obj, buf, api.Unversioned); err != nil {
		logger.Printf("Could not serialize object: %v\n", err)
		os.Exit(1)
	}

	//output, err := yaml.Marshal(obj)
	//if err := helmClient.Delete(namespace, bytes.NewBuffer(output)); err != nil {
	if err := helmClient.Delete(namespace, buf); err != nil {
		logger.Printf("Could not undeploy pod: %v\n", err)
		os.Exit(1)
	}

	logger.Println("Show others")

	pods, err := restclient.Pods(namespace).List(api.ListOptions{})
	if err != nil {
		logger.Printf("Could not list pods: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("pods: %+v\n", pods)
}
