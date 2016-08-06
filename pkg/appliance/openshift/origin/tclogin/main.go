package main

import (
	_ "bytes"
	"log"
	"os"
	_ "strings"

	"github.com/helm/helm-classic/codec"

	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	userapi "github.com/openshift/origin/pkg/user/api"
	userapiv1 "github.com/openshift/origin/pkg/user/api/v1"

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
	clientConfig.GroupVersion = &userapiv1.SchemeGroupVersion
	clientConfig.APIPath = "/oapi"
	//clientConfig.Codec = kapi.Codecs.LegacyCodec(projectapiv1.SchemeGroupVersion)
	clientConfig.Codec = kapi.Codecs.CodecForVersions(runtime.NoopEncoder{Decoder: kapi.Codecs.UniversalDeserializer()}, nil, []unversioned.GroupVersion{userapiv1.SchemeGroupVersion})

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
	result, err := osClient.Users().Get("~")
	if err != nil {
		logger.Fatal(err)
	}
	if len(result.Name) > 0 && len(result.UID) > 0 {
		logger.Println(result)
		os.Exit(0)
	}

	result = &userapi.User{}
	b, err := osRestClient.Get().Resource("users").Name("~").DoRaw()
	if err != nil {
		logger.Fatal(err)
	}
	log.Printf("raw: %+v\n", string(b))

	if err := osClient.OAuthAccessTokens().Delete(token); err != nil {
		logger.Fatal(err)
	}
	log.Println("sign out")

	result, err = osClient.Users().Get("~")
	if err != nil {
		logger.Fatal(err)
	}
	if len(result.Name) > 0 && len(result.UID) > 0 {
		logger.Println(result)
		os.Exit(0)
	}

	result = &userapi.User{}
	b, err = osRestClient.Get().Resource("users").Name("~").DoRaw()
	if err != nil {
		logger.Fatal(err)
	}
	log.Printf("raw: %+v\n", string(b))

	hobj, err := codec.JSON.Decode(b).One()
	if err != nil {
		logger.Fatalf("Could not set up helm classic codec object: %s\n", err)
	}

	if err := hobj.Object(result); err != nil {
		logger.Fatalf("Could not decode into openshift object: %+v\n", err)
	}

	log.Println(result, "\n", string(b))
}
