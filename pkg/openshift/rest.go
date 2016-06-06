package openshift

import (
	"bytes"
	"strings"

	"github.com/helm/helm-classic/codec"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/util/tokencmd"
	//oauthapi "github.com/openshift/origin/pkg/oauth/api"
	oauthapiv1 "github.com/openshift/origin/pkg/oauth/api/v1"
	projectapi "github.com/openshift/origin/pkg/project/api"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/logger"
)

var (
	serverNormalized string = "https://172.17.4.50:30448"
	caFile           string = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	certFile         string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	keyFile          string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"

	Versions = []unversioned.GroupVersion{{Group: "", Version: "v1"}, {Group: "", Version: "v1beta3"}}
	Version  = unversioned.GroupVersion{Group: "", Version: "v1"}

	SchemeGroupVersion = unversioned.GroupVersion{Group: GroupName, Version: runtime.APIVersionInternal}
	apiPath            = "/oapi"
)

func NewClientConfig() *restclient.Config {
	clientConfig := &restclient.Config{}

	clientConfig.Host = serverNormalized
	clientConfig.CAFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	clientConfig.CertFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	clientConfig.KeyFile = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"
	clientConfig.GroupVersion = Version
	clientConfig.APIPath = "/oapi"
	clientConfig.NegotiatedSerializer = kapi.Codecs
	clientConfig.Codec = kapi.Codecs.LegacyCodec(Versions)
	clientConfig.Username = "tangfeixiong"
	clientConfig.Password = "tangfeixiong"
	clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"
}

func EnterWorkspace(username, password string) (token string, err error) {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = oauthapiv1.SchemeGroupVersion
	clientConfig.Username = username
	clientConfig.Password = password
	clientConfig.BearerToken = ""
	token, err = tokencmd.RequestToken(clientConfig, nil, clientConfig.Username, clientConfig.Password)
	if err != nil {
		logger.Logger.Printf("Could not get TOKEN: %s\n", err)
	}
	logger.Logger.Printf("current token: %s\n", token)
	return
}

func LeaveWorkspace(token string) error {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = oauthapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not provide REST client: %s\n", err)
	}
	osClient := &client.Client{restClient}
	if err := osClient.OAuthAccessTokens().Delete(token); err != nil {
		logger.Logger.Println(err)
		return err
	}
	return nil
}

func CreateProject(token string, projectName, displayName, description string) error {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = projectapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not provide REST client: %s\n", err)
		return err
	}

	pr := new(projectapi.ProjectRequest)
	pr.Kind = "ProjectRequest"
	pr.APIVersion = "v1"
	pr.Name = projectName
	if len(strings.TrimSpace(displayName)) == 0 {
		pr.DisplayName = projectName
	} else {
		pr.DisplayName = displayName
	}
	if len(strings.TrimSpace(description)) == 0 {
		pr.Description = projectName
	} else {
		pr.Description = description
	}

	return createProject(restClient, pr)
}

func createProject(restClient *restclient.RESTClient, pr *projectapi.ProjectRequest) error {
	osClient := &client.Client{restClient}
	result1, err := osClient.ProjectRequests().Create(pr)
	if err != nil {
		if strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder") {
			var buf bytes.Buffer
			if err := codec.JSON.Encode(&buf).One(pr); err != nil {
				logger.Logger.Printf("Could not set up encoder: %s\n", err)
				return err
			}
			//so if failure (already exist), return unversioned.Status
			//{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Failure","message":"project \"gogogo\" already exists","reason":"AlreadyExists","details":{"name":"gogogo","kind":"project"},"code":409}
			b, err = osRestClient.Post().Resource("projectRequests").Body(buf.Bytes()).DoRaw()
			if err != nil {
				logger.Logger.Printf("Bad request to create project: %s\n", err)
				return err
			}
		} else {
			logger.Logger.Printf("Could not request to create project: %s\n", err)
		}
	}

	log.Println(result1, "\n", string(b))
	return nil
}

func RetrieveProjects(token string) (*projectapi.ProjectList, error) {
	clientConfig := NewClientConfig()()
	clientConfig.GroupVersion = projectapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	osClient := &client.Client{restClient}
	val, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		logger.Logger.Fatal(err)
	}
	logger.Logger.Printf("result: %+v\n", result)

	if val != nil && len(val.Items) > 0 && len(val[0].Name) > 0 && len(val[1].UID) > 0 {
		return val, nil
	}

	result, err := restClient.Get().Resource("projects").VersionedParams(&kapi.ListOptions{}, kapi.ParameterCodec).DoRaw()
	if err != nil {
		logger.Logger.Fatal(err)
	}
	logger.Logger.Printf("raw: %+v\n", string(b))

	hobj, err := codec.JSON.Decode(b).One()
	if err != nil {
		logger.Logger.Fatalf("Could not set up helm classic codec object: %s\n", err)
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
	return &obj, nil
}

func RetrieveProject(token string, name string) (*projectapi.Project, error) {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = projectapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not provide REST client: %s\n", err)
		return nil, err
	}

	osClient := &client.Client{restClient}
	val, err := osClient.Projects().Get(name)
	if err != nil {
		logger.Logger.Printf("Could not get openshift object", err)
		return nil, err
	}
	if val != nil && len(val.Name) > 0 && len(val.UID) > 0 {
		return val, nil
	}

	result, err := restClient.Get().Resource("projects").Name(name).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not get object data", err)
	}
	logger.Logger.Printf("raw: %+v\n", string(b))

	hobj, err := codec.JSON.Decode(b).One()
	if err != nil {
		logger.Logger.Printf("Could not set up helm classic codec object: %s\n", err)
		return nil, err
	}

	val = new(projectapi.Project)
	if err := hobj.Object(val); err != nil {
		logger.Logger.Printf("Could not decode into openshift object: %s\n", err)
		return nil, err
	}
	val.Kind = "Project"
	val.APIVersion = "v1"
	return val, nil
}

func DeleteProject(token string, name string) error {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = projectapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not provide REST client: %s\n", err)
	}
	osClient := &client.Client{restClient}
	if err := osClient.Projects().Delete(name); err != nil {
		logger.Logger.Println(err)
		return err
	}
	return nil
}

func BuildSingleDockerfileIntoRegistry(token string, project string, dockerfileText string, repository string, registryBasicAuth string) {

}

func BuildGithubDockerfileIntoRegistry(token string, project string, uri string, ref string, contextPath string, repository string, registryBasicAuth string) {

}

func BuildDockerfileContextArchiveIntoRegistry(token string, project string, contextArchive, contextPath string, repository string, registryBasicAuth string) {

}

func BuildDockerImageIntoRegistry(token string, build buildapi.Build) {

}

func StartOrRebuildDockerImageIntoRegistry(token string, buildConfig buildapi.BuildConfig) {

}

func CreateDockerImageBuildingConfiguration(token string, buildConfig buildapi.BuildConfig) {

}

func RetrieveDockerImageBuildingConfigurations(token string) buildapi.BuildConfigList {

}

func RetrieveDockerImageBuildingConfiguration(token string, buildName, projectName string) {

}

func DeleteDockerImageBuildingConfiguration(token string, buildName, projectName string) {

}
