package openshift

import (
	"bytes"
	"errors"
	_ "errors"
	"fmt"
	"log"
	_ "reflect"
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

	Versions           = []unversioned.GroupVersion{{Group: "", Version: "v1"}, {Group: "", Version: "v1beta3"}}
	Version            = unversioned.GroupVersion{Group: "", Version: "v1"}
	SchemeGroupVersion = unversioned.GroupVersion{Group: kapi.GroupName, Version: runtime.APIVersionInternal}

	apiPath               string = "/oapi"
	builderServiceAccount string = "builder"

	overrideDockerfile string = `"FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"`
	githubURI          string = "https://github.com/tangfeixiong/docker-nc.git"
	githubRef          string = "master"
	githubPath         string = "latest"
	githubSecret       string = "github-qingyuancloud-tangfx"
	dockerPullSecret   string = "tangfeixiong"
	dockerPushSecret   string = "tangfeixiong"
	timeout            int64  = 900
)

func NewClientConfig() *restclient.Config {
	clientConfig := &restclient.Config{}

	clientConfig.Host = serverNormalized
	clientConfig.CAFile = caFile
	clientConfig.CertFile = certFile
	clientConfig.KeyFile = keyFile
	clientConfig.GroupVersion = &Version
	clientConfig.APIPath = apiPath
	//clientConfig.NegotiatedSerializer = kapi.Codecs
	clientConfig.Codec = kapi.Codecs.LegacyCodec(Versions...)
	clientConfig.Username = "tangfeixiong"
	clientConfig.Password = "tangfeixiong"
	clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"

	return clientConfig
}

func SignIn(username, password string) (token string, err error) {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = &oauthapiv1.SchemeGroupVersion
	clientConfig.Username = username
	clientConfig.Password = password
	clientConfig.BearerToken = ""
	token, err = tokencmd.RequestToken(clientConfig, nil, clientConfig.Username, clientConfig.Password)
	if err != nil {
		logger.Logger.Printf("Could not get TOKEN: %s\n", err)
	}
	logger.Logger.Printf("Current TOKEN: %s\n", token)
	return
}

func SignOut(token string) error {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = &oauthapiv1.SchemeGroupVersion
	clientConfig.Username = ""
	clientConfig.BearerToken = token

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate REST client: %s\n", err)
		return err
	}

	osClient := &client.Client{restClient}
	if err := osClient.OAuthAccessTokens().Delete(token); err != nil {
		logger.Logger.Printf("Could not access Openshift OAuth service: %s\n", err)
		return err
	}
	return nil
}

type Workspace interface {
	ProjectAppliance() *ProjectAppliance
	DockerImageAppliance() *DockerImageAppliance
}

type Appliance interface {
	Workspace() Workspace
}

type workspace struct {
	clientConfig *restclient.Config
	RESTClient   *restclient.RESTClient
	appliances   []Appliance
}

func EnterWorkspace(username, password string) Workspace {
	clientConfig := NewClientConfig()
	clientConfig.GroupVersion = &oauthapiv1.SchemeGroupVersion
	clientConfig.Username = username
	clientConfig.Password = password
	clientConfig.BearerToken = ""
	token, err := tokencmd.RequestToken(clientConfig, nil, clientConfig.Username, clientConfig.Password)
	if err != nil {
		logger.Logger.Printf("Could not get TOKEN: %s\n", err)
		return nil
	}
	logger.Logger.Printf("Current TOKEN: %s\n", token)
	clientConfig.BearerToken = token

	return &workspace{
		clientConfig: clientConfig,
	}
}

func LeaveWorkspace(ws Workspace) error {
	if ws == nil || ws.(*workspace).clientConfig == nil || len(ws.(*workspace).clientConfig.BearerToken) == 0 {
		return errUnexpected
	}
	clientConfig := ws.(*workspace).clientConfig
	clientConfig.GroupVersion = &oauthapiv1.SchemeGroupVersion
	username := clientConfig.Username
	token := clientConfig.BearerToken
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate REST client: %s\n", err)
		return err
	}

	osClient := &client.Client{restClient}
	if err := osClient.OAuthAccessTokens().Delete(token); err != nil {
		logger.Logger.Printf("Could not access Openshift OAuth service: %s\n", err)
		return err
	}
	return nil
}

func (ws *workspace) ProjectAppliance() *ProjectAppliance {
	for _, v := range ws.appliances {
		switch t := v.(type) {
		case *ProjectAppliance:
			return v.(*ProjectAppliance)
		default:
			fmt.Println(t)
		}
	}
	appliance := NewProjectAppliance(ws)
	return appliance
}

func (ws *workspace) DockerImageAppliance() *DockerImageAppliance {
	for _, v := range ws.appliances {
		switch t := v.(type) {
		case *DockerImageAppliance:
			return v.(*DockerImageAppliance)
		default:
			fmt.Println(t)
		}
	}
	appliance := NewDockerImageAppliance(ws)
	return appliance
}

type ProjectAppliance struct {
	workspace *workspace
}

func NewProjectAppliance(ws Workspace) *ProjectAppliance {
	appliance := &ProjectAppliance{
		workspace: ws.(*workspace),
	}
	ws.(*workspace).appliances = append(ws.(*workspace).appliances, appliance)
	return appliance
}

func (app *ProjectAppliance) Workspace() Workspace {
	return app.workspace
}

/*
func (app *ProjectAppliance) CreateProject(name, displayname, description string) (*projectapi.Project, error) {
	app.clientConfig.GroupVersion = &projectapiv1.SchemeGroupVersion
	app.clientConfig.Username = ""

	pr := new(projectapi.ProjectRequest)
	pr.Kind = "ProjectRequest"
	pr.APIVersion = "v1"
	pr.Name = name
	pr.DisplayName = displayname
	pr.Description = description

	return createProject(pr)
}

func (app *ProjectAppliance) CreateProject(pr *projectapi.ProjectRequest) (*projectapi.Project, error) {
    if pr == nil {
        return nil, errUnexpected
    }
	osClient, err := &client.New(app.clientConfig)
    if err != nil {
        logger.Logger.Printf("Could not generate Openshift client: %+v", err)
        return nil, err
    }
	result, err := osClient.ProjectRequests().Create(pr)
	if err != nil {
		if strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder") {
	        app.RESTClient, err = restclient.RESTClientFor(app.clientConfig)
	        if err != nil {
		        logger.Logger.Printf("Could not genterate REST client: %s\n", err)
		        return err
	        }
			var buf bytes.Buffer
			if err := codec.JSON.Encode(&buf).One(pr); err != nil {
				logger.Logger.Printf("Could not set up encoder: %s\n", err)
				return nil, err
			}
			b, err = app.RESTClient.Post().Resource("projectRequests").Body(buf.Bytes()).DoRaw()
			if err != nil {
				logger.Logger.Printf("Bad request to create project: %s\n", err)
				return nil, err
			}
	        logger.Logger.Println(string(b))
            //notice: if project is already exist, return unversioned.Status
			//{"kind":"Status","apiVersion":"v1","metadata":{},"status":"Failure","message":"project \"gogogo\" already exists","reason":"AlreadyExists","details":{"name":"gogogo","kind":"project"},"code":409}
            hobj, err := codec.JSON.decode(b).One()
            if err != nil {
               logger.Logger.Printf("Could not set up decoder: %s\n", err)
				return nil, err
            }
            meta, err := hobj.Meta()
            if strings.EqualFold(meta, "Status") {
               status := unversioned.Status{}
               if err := hobj.Object(&status); err != nil {
                   logger.Logger.Printf("Could not decode into k8s object: %s\n", err)
                    return nil, err
               }
               return nil, errors.New(status.Message)
            }
            val := projectapi.Project{}
            if err := hobj.Object(&val); err != nil {
               logger.Logger.Printf("Could not decode into Openshift object: %s\n", err)
               return nil, err
            }
            return val, nil
		} else {
			logger.Logger.Printf("Could not request to create project: %s\n", err)
            return nil, err
		}
	}

	logger.Logger.Println(result)
	return result, nil
}


func RetrieveProjects() error {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		return err
	}
	logger.Printf("openshift client: %+v\n", oc)

	result, err := oc.Projects().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

*/

func (app *ProjectAppliance) RetrieveProjects() (*projectapi.ProjectList, error) {
	ws := app.workspace
	if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
		return nil, errUnexpected
	}
	clientConfig := ws.clientConfig
	clientConfig.GroupVersion = &projectapiv1.SchemeGroupVersion
	username := clientConfig.Username
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate REST client: %s\n", err)
		return nil, err
	}

	osClient := &client.Client{restClient}
	val, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		logger.Logger.Printf("Could not access Openshift object: %s\n", err)
		return nil, err
	}
	if val != nil && (len(val.Items) == 0 || len(val.Items[0].Name) > 0 && len(val.Items[0].UID) > 0) {
		log.Printf("result: %+v\n", val)
		return val, nil
	}

	result, err := restClient.Get().Resource("projects").VersionedParams(&kapi.ListOptions{}, kapi.ParameterCodec).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not access Openshift: %s\n", err)
		return nil, err
	}

	hobj, err := codec.JSON.Decode(result).One()
	if err != nil {
		log.Printf("raw: %+v\n", string(result))
		logger.Logger.Printf("Could not generate codec: %s\n", err)
		return nil, err
	}

	var obj projectapi.ProjectList
	if err := hobj.Object(&obj); err != nil {
		logger.Logger.Printf("Could not decode into openshift object: %s\n", err)
		return nil, err
	}

	var kobj runtime.Object
	var olist kapi.List
	olist.Kind = "ProjectList"
	olist.APIVersion = Version.Version
	for i := 0; i < len(obj.Items); i += 1 {
		v := &obj.Items[i]
		v.Kind = "Project"
		v.APIVersion = projectapiv1.SchemeGroupVersion.Version

		var buf bytes.Buffer
		if err := codec.JSON.Encode(&buf).One(v); err != nil {
			logger.Logger.Printf("Could not encode openshift object: %s\n", err)
			continue
		}
		kobj = v
		olist.Items = append(olist.Items, kobj)
		log.Printf("runtime object: %+v, %s\n", kobj.(*projectapi.Project), buf.String())
	}

	log.Printf("ProjectList: %+v\n", olist)
	return &obj, nil
}

func (app *ProjectAppliance) RetrieveProjectWithJSON(json []byte) (*projectapi.Project, error) {
	if len(json) == 0 {
		return nil, errUnexpected
	}
	hobj, err := codec.JSON.Decode(json).One()
	if err != nil {
		logger.Logger.Printf("Could not set up codec: %+v", err)
		return nil, err
	}
	obj := &projectapi.Project{}
	if err := hobj.Object(obj); err != nil {
		logger.Logger.Printf("Could not decode into openshift object: %+v", err)
		return nil, err
	}
	obj.Kind = "Project"
	obj.APIVersion = projectapiv1.SchemeGroupVersion.Version
	return app.RetrieveProjectFrom(obj)
}

func (app *ProjectAppliance) RetrieveProjectFrom(obj *projectapi.Project) (*projectapi.Project, error) {
	if obj == nil || len(obj.Name) == 0 {
		return nil, errUnexpected
	}
	return app.RetrieveProject(obj.Name)
}

func (app *ProjectAppliance) RetrieveProject(name string) (*projectapi.Project, error) {
	ws := app.workspace
	if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
		return nil, errUnexpected
	}
	clientConfig := ws.clientConfig
	clientConfig.GroupVersion = &projectapiv1.SchemeGroupVersion
	username := clientConfig.Username
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	restClient, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate REST client: %s\n", err)
		return nil, err
	}

	osClient := &client.Client{restClient}
	val, err := osClient.Projects().Get(name)
	if err != nil {
		logger.Logger.Printf("Could not access Openshift object: %s\n", err)
		return nil, err
	}
	if val != nil && len(val.Name) > 0 && len(val.UID) > 0 {
		log.Printf("result: %+v\n", val)
		return val, nil
	}

	result, err := restClient.Get().Resource("projects").Name(name).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not access Openshift: %s\n", err)
		return nil, err
	}

	hobj, err := codec.JSON.Decode(result).One()
	if err != nil {
		log.Printf("raw: %+v\n", string(result))
		logger.Logger.Printf("Could not generate codec: %s\n", err)
		return nil, err
	}

	val = new(projectapi.Project)
	if err := hobj.Object(val); err != nil {
		logger.Logger.Printf("Could not decode into openshift object: %s\n", err)
		return nil, err
	}
	val.Kind = "Project"
	val.APIVersion = projectapiv1.SchemeGroupVersion.Version
	return val, nil
}

/*

func RetrieveProject(name string) ([]byte, *projectapi.Project, error) {
	if len(name) == 0 {
		return nil, nil, errNotFound
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %+v", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	result, err := oc.Projects().Get(name)
	if err != nil {
		glog.Errorf("Could not delete project %s: %+v", name, err)
		return nil, nil, err
	}
	if result == nil {
		glog.V(7).Infoln("Unexpected retrieve: %s", name)
		return nil, nil, errUnexpected
	}
	if strings.EqualFold("Project", result.Kind) && len(result.Name) > 0 {
		//b := new(bytes.Buffer)
		//if err := codec.JSON.Encode(b).One(result); err != nil {
		//	glog.Errorf("Could not encode runtime object: %s", err)
		//	return nil, result, err
		//}
		//logger.Printf("Build: %+v\n", b.String())
		kapi.Scheme.AddKnownTypes(projectapiv1.SchemeGroupVersion, &projectapiv1.Project{})
		data, err := runtime.Encode(kapi.Codecs.LegacyCodec(projectapiv1.SchemeGroupVersion),
			result, projectapiv1.SchemeGroupVersion)
		if err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, result, err
		}
		return data, result, nil
	}

	raw, err := oc.RESTClient.Get().Resource("projects").Name(name).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	//kapi.Scheme.AddKnownTypes(projectapiv1.SchemeGroupVersion, &projectapiv1.Project{})
	//obj, err := runtime.Decode(kapi.Codecs.UniversalDeserializer(), raw)
	//if err != nil {
	//	glog.Errorf("Could not deserialize raw: %+v", err)
	//	return raw, nil, err
	//}
	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf("Could not create helm object: %s", err)
		return raw, nil, err
	}
	meta, err := hco.Meta()
	if err != nil {
		glog.Errorf("Could not decode into metadata: %s", err)
		return raw, nil, err
	}
	if ok := strings.EqualFold("Project", meta.Kind) && len(meta.Name) > 0; !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not know metadata: %+v", meta)
				return raw, nil, err
			}
			glog.Warningf("Could not find runtime object: %+v", status.Message)
			return raw, nil, fmt.Errorf("Could not find runtime object: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result = new(projectapi.Project)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode raw data: %s", err)
		return raw, nil, err
	}
	logger.Printf("Return runtime object: %s\n", string(raw))
	return raw, result, nil
}

*/

/*
func (app *ProjectAppliance) DeleteProject(name string) error {
	if len(name) == 0 {
		return errUnexpected
	}
	osClient, err := client.New(app.clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate OpenShift client: %s\n", s)
		return err
	}

	if err := osClient.Projects().Delete(name); err != nil {
		logger.Logger.Printf("Could not access OpenShift: %s\n", s)
		return err
	}
	return nil
}
*/

type DockerImageAppliance struct {
	workspace *workspace
}

func NewDockerImageAppliance(ws Workspace) *DockerImageAppliance {
	appliance := &DockerImageAppliance{
		workspace: ws.(*workspace),
	}
	ws.(*workspace).appliances = append(ws.(*workspace).appliances, appliance)
	return appliance
}

func (app *DockerImageAppliance) Workspace() Workspace {
	return app.workspace
}

func (app *DockerImageAppliance) BuildDockerImageIntoRegistryWithJSON(json []byte) ([]byte, *buildapi.Build, error) {
	if len(json) == 0 {
		return nil, nil, errUnexpected
	}
	hobj, err := codec.JSON.Decode(json).One()
	if err != nil {
		logger.Logger.Printf("Could not set up codec: %+v", err)
		return nil, nil, err
	}
	obj := &buildapi.Build{}
	if err := hobj.Object(obj); err != nil {
		logger.Logger.Printf("Could not decode into openshift object: %+v", err)
		return nil, nil, err
	}
	obj.Kind = "Build"
	obj.APIVersion = buildapiv1.SchemeGroupVersion.Version

	return app.buildDockerImageIntoRegistry(json, obj)
}

func (app *DockerImageAppliance) BuildDockerImageIntoRegistryFrom(name, projectName string, gitSecret map[string]string, gitURI, branchTagCommit, contextDir string, sourceImages []map[string]interface{}, dockerfile string, buildSecrets []map[string]interface{}, buildStrategy map[string]interface{}) ([]byte, *buildapi.Build, error) {
	obj := &buildapi.Build{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Build",
			APIVersion: buildapiv1.SchemeGroupVersion.Version,
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:              name,
			Namespace:         projectName,
			CreationTimestamp: unversioned.Now(),
		},
		Spec: buildapi.BuildSpec{
			ServiceAccount: builderServiceAccount,
			Source: buildapi.BuildSource{
				//Binary : &buildapi.BinaryBuildSource {},
				Dockerfile: &dockerfile,
				Git: &buildapi.GitBuildSource{
					URI: gitURI,
					Ref: branchTagCommit,
					//HTTPProxy: nil,
					//HTTPSProxy: nil,
				},
				/*Images : []buildapi.ImageSource {
				    buildapi.ImageSource {
				        From : kapi.ObjectReference {
				            Kind : "DockerImage",
				            Name : "alpine:edge",
				        },
				        Paths : []buildapi.ImageSourcePath {
				           {
				               SourcePath : "",
				               DestinationDir : "",
				           },
				        },
				        PullSecret : &kapi.LocalObjectReference {
				        },
				   },
				},*/
				ContextDir: contextDir,
				//SourceSecret : &kapi.LocalObjectReference {
				//    name : githubSecret,
				//},
				//Secrets : []buildapi.SecretBuildSource {
				//    Secret : &kapi.LocalObjectReference {},
				//    DestinationDir : "/root/.docker/config.json",
				//},
			},
			//Revision: &buildapi.SourceRevision {},
			/*Strategy: buildapi.BuildStrategy{
				DockerStrategy: &buildapi.DockerBuildStrategy{
					From: &kapi.ObjectReference{
						Kind: "DockerImage",
						Name: "alpine:edge",
					},
					PullSecret: &kapi.LocalObjectReference{
						Name: dockerPullSecret,
					},
					NoCache: false,
					//Env : []kapi.EnvVar {},
					ForcePull: false,
					//DockerfilePath : ".",
				},
			},*/
			Output: buildapi.BuildOutput{
				To: &kapi.ObjectReference{
					Kind: "DockerImage",
					Name: "docker.io/tangfeixiong/nc-http-dev:latest",
				},
				PushSecret: &kapi.LocalObjectReference{
					Name: dockerPushSecret,
				},
			},
			//Resources : kapi.ResourceRequirements {},
			//PostCommit : buildapi.BuildPostCommitSpec {
			//    Command : []string{},
			//    Args : []string{},
			//    Script: "",
			//},
			CompletionDeadlineSeconds: &timeout,
		},
		Status: buildapi.BuildStatus{
			Phase: buildapi.BuildPhaseNew,
		},
	}

	buf := bytes.Buffer{}
	if err := codec.JSON.Encode(&buf).One(obj); err != nil {
		logger.Logger.Printf("Could not encode openshift object: %+v", err)
		return nil, nil, err
	}
	return app.BuildDockerImageIntoRegistry(buf.Bytes(), obj)
}

func (app *DockerImageAppliance) BuildDockerImageIntoRegistry(raw []byte, build *buildapi.Build) ([]byte, *buildapi.Build, error) {
	ws := app.workspace
	if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
		return nil, nil, errUnexpected
	}
	clientConfig := ws.clientConfig
	clientConfig.GroupVersion = &buildapiv1.SchemeGroupVersion
	username := clientConfig.Username
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	osClient, err := client.New(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate OpenShift client: %s\n", err)
		return nil, nil, err
	}

	if len(raw) > 0 {
		if build == nil || len(build.Name) == 0 || len(build.UID) == 0 {
			hobj, err := codec.JSON.Decode(raw).One()
			if err != nil {
				logger.Logger.Printf("Could not set up helm classic codec: %s\n", err)
				return nil, nil, err
			}
			build = new(buildapi.Build)
			if err := hobj.Object(build); err != nil {
				logger.Logger.Printf("Could not decode into openshift object: %+v", err)
				return nil, nil, err
			}
			return app.buildDockerImageIntoRegistry(raw, build)
		}
	}

	val, err := osClient.Builds(build.Namespace).Create(build)
	if err != nil {
		if !strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder") {
			logger.Logger.Printf("Could not build Docker image into registry: %s\n", err)
			return nil, nil, err
		}
	}
	if val != nil && len(val.Name) > 0 && len(val.UID) > 0 {
		buf := bytes.Buffer{}
		if err := codec.JSON.Encode(&buf).One(val); err != nil {
			logger.Logger.Printf("Could not encode openshift object: %s\n", err)
			return nil, nil, err
		}
		return buf.Bytes(), val, nil
	}

	if len(raw) == 0 {
		buf := bytes.Buffer{}
		if err := codec.JSON.Encode(&buf).One(build); err != nil {
			logger.Logger.Printf("Could not encode openshift object: %+v", err)
			return nil, nil, err
		}
		raw = buf.Bytes()
	}

	app.workspace.RESTClient = osClient.RESTClient

	return app.buildDockerImageIntoRegistry(raw, build)

}

func (app *DockerImageAppliance) buildDockerImageIntoRegistry(raw []byte, build *buildapi.Build) ([]byte, *buildapi.Build, error) {
	ws := app.workspace
	if ws.RESTClient == nil {
		if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
			return nil, nil, errUnexpected
		}
		clientConfig := ws.clientConfig
		clientConfig.GroupVersion = &buildapiv1.SchemeGroupVersion
		username := clientConfig.Username
		clientConfig.Username = ""
		defer func() {
			clientConfig.Username = username
		}()

		var err error
		if ws.RESTClient, err = restclient.RESTClientFor(clientConfig); err != nil {
			logger.Logger.Printf("Could not generate REST client: %s\n", err)
			return nil, nil, err
		}
	}
	val, err := ws.RESTClient.Post().Namespace(build.Namespace).Resource("builds").Body(raw).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not access Openshift: %s\n", err)
		return nil, nil, err
	}

	hobj, err := codec.JSON.Decode(val).One()
	if err != nil {
		logger.Logger.Printf("Could not set up helm codec: %s\n", err)
		return val, nil, err
	}

	meta, err := hobj.Meta()
	if err != nil {
		logger.Logger.Printf("Could not set up helm codec: %s\n", err)
		return val, nil, err
	}
	if strings.EqualFold(meta.Kind, "Status") {
		log.Printf("Return: %s", string(val))
		var status unversioned.Status
		if err := hobj.Object(&status); err != nil {
			logger.Logger.Printf("Could not encode into openshift object: %+v", err)
			return val, nil, err
		}
		return nil, nil, errors.New(status.Message)
	}

	//var build buildapi.Build
	if err := hobj.Object(build); err != nil {
		logger.Logger.Printf("Could not encode into openshift object: %+v", err)
		return val, nil, err
	}
	build.Kind = "Build"
	build.APIVersion = "v1"
	return val, build, nil
}

func (app *DockerImageAppliance) RetrieveDockerImageBuilders(namespace string) ([]byte, *buildapi.BuildList, error) {
	if len(namespace) == 0 {
		return nil, nil, errUnexpected
	}
	ws := app.workspace
	if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
		return nil, nil, errUnexpected
	}
	clientConfig := ws.clientConfig
	clientConfig.GroupVersion = &buildapiv1.SchemeGroupVersion
	username := clientConfig.Username
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	osClient, err := client.New(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate OpenShift client: %s\n", err)
		return nil, nil, err
	}

	result, err := osClient.Builds(namespace).List(kapi.ListOptions{})
	if err != nil {
		logger.Logger.Printf("Could not access openshift object: %s\n", err)
		return nil, nil, err
	}
	if result != nil && (len(result.Items) == 0 || len(result.Items[0].Name) > 0 && len(result.Items[0].UID) > 0) {
		buf := bytes.Buffer{}
		if err := codec.JSON.Encode(&buf).One(result); err != nil {
			logger.Logger.Printf("Could not encode openshift object: %s\n", err)
			return nil, nil, err
		}
		return buf.Bytes(), result, nil
	}

	ws.RESTClient = osClient.RESTClient
	val, err := ws.RESTClient.Get().Namespace(namespace).Resource("builds").VersionedParams(&kapi.ListOptions{}, kapi.ParameterCodec).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not access Openshift: %s\n", err)
		return nil, nil, err
	}

	hobj, err := codec.JSON.Decode(val).One()
	if err != nil {
		logger.Logger.Printf("Could not set up helm codec: %s\n", err)
		return val, nil, err
	}

	obj := new(buildapi.BuildList)
	if err := hobj.Object(obj); err != nil {
		logger.Logger.Printf("Could not encode openshift object: %s\n", err)
		return val, nil, err
	}
	obj.Kind = "BuildList"
	obj.APIVersion = buildapiv1.SchemeGroupVersion.Version

	for _, v := range obj.Items {
		v.Kind = "Build"
		v.APIVersion = buildapiv1.SchemeGroupVersion.Version
	}
	return val, obj, nil
}

func (app *DockerImageAppliance) RetrieveDockerImageBuilder(namespace, name string) ([]byte, *buildapi.Build, error) {
	if len(namespace) == 0 || len(name) == 0 {
		return nil, nil, errUnexpected
	}
	ws := app.workspace
	if ws == nil || ws.clientConfig == nil || len(ws.clientConfig.BearerToken) == 0 {
		return nil, nil, errUnexpected
	}
	clientConfig := ws.clientConfig
	clientConfig.GroupVersion = &buildapiv1.SchemeGroupVersion
	username := clientConfig.Username
	clientConfig.Username = ""
	defer func() {
		clientConfig.Username = username
	}()

	osClient, err := client.New(clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate OpenShift client: %s\n", err)
		return nil, nil, err
	}

	build, err := osClient.Builds(namespace).Get(name)
	if err != nil {
		logger.Logger.Printf("Could not access openshift object: %s\n", err)
		return nil, nil, err
	}
	if build != nil && len(build.Name) > 0 && len(build.UID) > 0 {
		buf := &bytes.Buffer{}
		if err := codec.JSON.Encode(buf).One(build); err != nil {
			logger.Logger.Printf("Could not encode openshift object: %s\n", err)
			return nil, nil, err
		}
		return buf.Bytes(), build, nil
	}

	ws.RESTClient = osClient.RESTClient
	val, err := ws.RESTClient.Get().Namespace(namespace).Resource("builds").Body(name).DoRaw()
	if err != nil {
		logger.Logger.Printf("Could not access Openshift: %s\n", err)
		return nil, nil, err
	}
	hobj, err := codec.JSON.Decode(val).One()
	if err != nil {
		logger.Logger.Printf("Could not set up helm codec: %s\n", err)
		return val, nil, err
	}
	//var build buildapi.Build
	build = new(buildapi.Build)
	if err := hobj.Object(build); err != nil {
		logger.Logger.Printf("Could not encode openshift object: %s\n", err)
		return val, nil, err
	}
	build.Kind = "Build"
	build.APIVersion = buildapiv1.SchemeGroupVersion.Version
	return val, build, nil
}

/*
func (app *DockerImageBuildAppliance) RebuildDockerImageIntoRegistry(namespace, name string) ([]byte, *buildapi.Build, err) {
	raw, build, err := app.GetDockerImageBuild(namespace, name)
	if err != nil {
		return nil, err
	}
	return app.buildDockerImageIntoRegistry(raw, build)
}

func (app *DockerImageBuildAppliance) DeleteProject(namespace, name string) error {
	if len(namespace) == 0 || len(name) == 0 {
		return errUnexpected
	}
	osClient, err := client.New(app.clientConfig)
	if err != nil {
		logger.Logger.Printf("Could not generate OpenShift client: %s\n", s)
		return err
	}

	if err := osClient.Builds(namespace).Delete(name); err != nil {
		logger.Logger.Printf("Could not access OpenShift: %s\n", s)
		return err
	}
	return nil
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
*/
