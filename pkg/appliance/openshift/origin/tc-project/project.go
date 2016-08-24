package main

import (
	"bytes"
	"log"
	"os"

	"github.com/helm/helm-classic/codec"

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
	logger := log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

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
	//clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"
	token, err := tokencmd.RequestToken(clientConfig, os.Stdin, clientConfig.Username, clientConfig.Password)
	if err != nil {
		logger.Fatal(err)
	}
	log.Printf("current token: %+v\n", token)

	clientConfig.BearerToken = token
	//clientConfig.Password = ""
	clientConfig.Username = ""
	osRestClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Fatal(err)
	}

	osClient := &client.Client{osRestClient}
	result, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		logger.Fatal(err)
	}
	log.Printf("result: %+v\n", result)

	//result = &projectapi.ProjectList{}
	b, err := osRestClient.Get().Resource("projects").VersionedParams(&kapi.ListOptions{}, kapi.ParameterCodec).DoRaw()
	if err != nil {
		logger.Fatal(err)
	}
	log.Printf("raw: %+v\n", string(b))

	hobj, err := codec.JSON.Decode(b).One()
	if err != nil {
		logger.Fatalf("Could not set up helm classic codec object: %s\n", err)
	}

	var obj projectapi.ProjectList
	if err := hobj.Object(&obj); err != nil {
		logger.Fatalf("Could not decode into openshift object: %s\n", err)
	}

	var olist kapi.List
	olist.Kind = "ProjectList"
	olist.APIVersion = "v1"

	var kobj runtime.Object
	for i := 0; i < len(obj.Items); i += 1 {
		v := &obj.Items[i]
		v.Kind = "Project"
		v.APIVersion = "v1"

		var buf bytes.Buffer
		if err := codec.JSON.Encode(&buf).One(v); err != nil {
			logger.Fatalf("Could not encode with openshift object: %s\n", err)
		}
		log.Println(buf.String())

		kobj = v
		olist.Items = append(olist.Items, kobj)
		log.Println(kobj.(*projectapi.Project))
	}

	log.Printf("ProjectList: %+v\n%+v\n", olist, obj)
	os.Exit(0)
}
