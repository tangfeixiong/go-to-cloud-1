package main

import (
	"log"
	"os"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	projectapi "github.com/openshift/origin/pkg/project/api"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"
)

func main() {
	clientConfig := &restclient.Config{}
	serverNormalized, err := config.NormalizeServerURL("https://172.17.4.50:30448")
	if err != nil {
		log.Fatal(err)
	}
	clientConfig.Host = serverNormalized
	clientConfig.CAFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	clientConfig.CertFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	clientConfig.KeyFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"
	clientConfig.GroupVersion = &projectapiv1.SchemeGroupVersion
	clientConfig.APIPath = "/oapi"
	//clientConfig.Codec = kapi.Codecs.LegacyCodec(projectapiv1.SchemeGroupVersion)
	projectapi.AddToScheme(kapi.Scheme)
	projectapiv1.AddToScheme(kapi.Scheme)
	clientConfig.Codec = kapi.Codecs.CodecForVersions(runtime.NoopEncoder{Decoder: kapi.Codecs.UniversalDeserializer()}, nil, []unversioned.GroupVersion{projectapiv1.SchemeGroupVersion})

	log.Printf("simple config: %+v\n", clientConfig)

	clientConfig.Username = "tangfeixiong"
	clientConfig.Password = "tangfeixiong"
	token, err := tokencmd.RequestToken(clientConfig, os.Stdin, clientConfig.Username, clientConfig.Password)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("current token: %+v\n", token)
	clientConfig.BearerToken = token

	//clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"
	clientConfig.Username = ""
	clientConfig.Password = ""

	clientK8s, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		log.Fatal(err)
	}

	osClient := &client.Client{clientK8s}

	result, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("result: %+v", result)

	//result = &projectapi.ProjectList{}
	b, err := clientK8s.Get().Resource("projects").VersionedParams(&kapi.ListOptions{}, kapi.ParameterCodec).DoRaw()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("raw: %+v", string(b))

	os.Exit(0)
}
