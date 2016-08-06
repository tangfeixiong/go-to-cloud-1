package origin

import (
	"bytes"
	"fmt"
	"io"
	_ "io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"github.com/helm/helm-classic/codec"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	//"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	_ "github.com/openshift/origin/pkg/cmd/util/tokencmd"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	_ "github.com/openshift/origin/pkg/generate/git"
	projectapi "github.com/openshift/origin/pkg/project/api"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"
	userapi "github.com/openshift/origin/pkg/user/api"
	userapiv1 "github.com/openshift/origin/pkg/user/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	// kapierrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	// clientauth "k8s.io/kubernetes/pkg/client/unversioned/auth"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	// clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	// kubecmdconfig "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"
	_ "k8s.io/kubernetes/pkg/runtime/serializer/json"

	qyapi "github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/api"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

	factory *osclientcmd.Factory

	kubeconfigPath    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/kubeconfig"
	kubeconfigContext string = "openshift-origin-single"

	apiVersion string = "v1"
	oca        string = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	oclientcrt string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	oclientkey string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"

	// token string = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"

	oconfigPath    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
	oconfigContext string = "default/172-17-4-50:30448/system:admin"

	kClientConfig, oClientConfig *clientcmd.ClientConfig
	kClient                      *kclient.Client
	oClient                      *client.Client
	kConfig, oConfig             *restclient.Config

	who *userapi.User

	builderServiceAccount string = "builder"

	overrideDockerfile string = "FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"
	githubURI          string = "https://github.com/tangfeixiong/docker-nc.git"
	githubRef          string = "master"
	githubPath         string = "latest"
	githubSecret       string = "github-qingyuancloud-tangfx"
	dockerPullSecret   string = "tangfeixiong"
	dockerPushSecret   string = "tangfeixiong"
	timeout            int64  = 900
)

type stringValue struct {
	value string
}

func (val stringValue) String() string {
	return val.value
}
func (val stringValue) Set(v string) error {
	val.value = v
	return nil
}
func (val stringValue) Type() string {
	return "string"
}

type intValue struct {
	value int
}

func (val intValue) String() string {
	return strconv.Itoa(val.value)
}
func (val intValue) Set(v string) error {
	var err error
	val.value, err = strconv.Atoi(v)
	return err
}
func (val intValue) Type() string {
	return "int"
}

type boolValue struct {
	value bool
}

func (val boolValue) String() string {
	return strconv.FormatBool(val.value)
}
func (val boolValue) Set(v string) error {
	var err error
	val.value, err = strconv.ParseBool(v)
	return err
}
func (val boolValue) Type() string {
	return "bool"
}

var rootCommand = &cobra.Command{
	Use:   "oc",
	Short: "ociacibuilds",
	Long:  "The openshift image build server.",
	Run: func(c *cobra.Command, args []string) {
	},
}

//func init() {
//	flagtypes.GLog(rootCommand.PersistentFlags())
//}

// User, Group, Identity and UserIdentityMapping

func WhoAmI() (*userapi.User, error) {
	return RetrieveUser("~")
}

func CreateUser(name, fullName string, identities, groups []string) ([]byte, *userapi.User, error) {
	user := new(userapi.User)
	user.Kind = "User"
	user.APIVersion = userapiv1.SchemeGroupVersion.Version
	user.Name = name
	if len(fullName) > 0 {
		user.FullName = fullName
	}
	if len(identities) > 0 {
		user.Identities = identities
	}
	if len(groups) > 0 {
		user.Groups = groups
	}
	return CreateUserWith(user)
}

func CreateUserWith(user *userapi.User) ([]byte, *userapi.User, error) {
	return createUser(nil, user)
}

func CreateUserFromArbitrary(data []byte) ([]byte, *userapi.User, error) {
	return createUser(data, nil)
}

func createUser(data []byte, user *userapi.User) ([]byte, *userapi.User, error) {
	if len(data) == 0 && user == nil {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if len(data) == 0 && user != nil {
		result, err := oc.Users().Create(user)
		if err != nil {
			if retry := strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder"); !retry {
				glog.Errorf("Could not access openshift: %s", err)
				return nil, nil, err
			}
		}
		if result == nil {
			glog.V(7).Infoln("Unexpected creation: %+v", user)
			return nil, nil, errUnexpected
		}
		if result != nil {
			if strings.EqualFold("User", result.Kind) && len(result.Name) > 0 {
				b := new(bytes.Buffer)
				if err := codec.JSON.Encode(b).One(result); err != nil {
					glog.Errorf("Could not encode runtime object: %s", err)
					return nil, result, err
				}
				logger.Printf("User: %+v\n", b.String())
				return b.Bytes(), result, nil
			}
		}

		data = make([]byte, 0)
		b := bytes.NewBuffer(data)
		if err := codec.JSON.Encode(b).One(user); err != nil {
			glog.Errorf("Could not serialize object: %+v", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Post().Resource("users").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}

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
	if ok := strings.EqualFold("User", meta.Kind) && len(meta.Name) > 0; !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not know metadata: %+v", meta)
				return raw, nil, err
			}
			return raw, nil, fmt.Errorf("Could not create user: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result := new(userapi.User)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	logger.Printf("User: %+v\n", string(raw))
	return raw, result, nil
}

func RetrieveUsers() error {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return err
	}
	logger.Printf("openshift client: %+v\n", oc)

	result, err := oc.Users().List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func RetrieveUser(name string) (*userapi.User, error) {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return nil, err
	}
	logger.Printf("openshit client: %+v\n", oc)

	result, err := oc.Users().Get(name)
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, err
	}
	if result != nil {
		if strings.EqualFold("User", result.Kind) && len(result.Name) > 0 {
			logger.Printf("User: %+v\n", result)
			return result, nil
		}
		b, err := oc.RESTClient.Get().Resource("users").Name(name).DoRaw()
		if err != nil {
			glog.Errorf("Could not access openshift: %s", err)
			return nil, err
		}

		hco, err := codec.JSON.Decode(b).One()
		if err != nil {
			glog.Errorf("Could not create helm decoder: %s", err)
			return nil, err
		}
		meta, err := hco.Meta()
		if err != nil {
			glog.Errorf("Could not decode into metadata: %s", err)
			return nil, err
		}
		if ok := strings.EqualFold(meta.Kind, "User") && len(meta.Name) > 0; !ok {
			glog.Errorf("Could not know metadata: %+v", meta)
			return nil, errUnexpected
		}
		if err := hco.Object(result); err != nil {
			glog.Errorf("Could not decode runtime object: %s", err)
			return nil, err
		}
		who = result
		logger.Printf("User: %+v\n", result)
	}
	return result, nil
}

// ProjectRequest and Project

func CreateProjectRequest(name, displayName, description string) ([]byte, *projectapi.Project, error) {
	obj := new(projectapi.ProjectRequest)
	obj.Kind = "ProjectRequest"
	obj.APIVersion = projectapiv1.SchemeGroupVersion.Version
	obj.Name = name
	if len(displayName) > 0 {
		obj.DisplayName = displayName
	}
	if len(description) > 0 {
		obj.Description = description
	}
	return CreateProjectRequestWith(obj)
}

func CreateProjectRequestWith(obj *projectapi.ProjectRequest) ([]byte, *projectapi.Project, error) {
	return createProjectRequest(nil, obj)
}

func CreateProjectRequestFromArbitray(data []byte) ([]byte, *projectapi.Project, error) {
	return createProjectRequest(data, nil)
}

func createProjectRequest(data []byte, obj *projectapi.ProjectRequest) ([]byte, *projectapi.Project, error) {
	if len(data) == 0 && obj == nil {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if len(data) == 0 && obj != nil {
		result, err := oc.ProjectRequests().Create(obj)
		if err != nil {
			if retry := strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder"); !retry {
				glog.Errorf("Could not access openshift: %s", err)
				return nil, nil, err
			}
			if !strings.EqualFold("no kind is registered for the type api.ProjectRequest", err.Error()) {
				glog.Errorf("Could not access openshift: %s", err)
				return nil, nil, err
			}
		}
		if result == nil {
			glog.V(7).Infoln("Unexpected creation: %+v", obj)
			return nil, nil, errUnexpected
		}
		if result != nil {
			if strings.EqualFold("Project", result.Kind) && len(result.Name) > 0 {
				b := new(bytes.Buffer)
				if err := codec.JSON.Encode(b).One(result); err != nil {
					glog.Errorf("Could not encode runtime object: %s", err)
					return nil, result, err
				}
				logger.Printf("Project: %+v\n", b.String())
				return b.Bytes(), result, nil
			}
		}

		data = make([]byte, 0)
		b := bytes.NewBuffer(data)
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Post().Resource("projectRequests").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}

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
			return raw, nil, fmt.Errorf("Could not create project: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result := new(projectapi.Project)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	logger.Printf("Project: %+v\n", string(raw))
	return raw, result, nil
}

func CreateProject(name string, finalizers ...string) ([]byte, *projectapi.Project, error) {
	obj := new(projectapi.Project)
	obj.Kind = "Project"
	obj.APIVersion = projectapiv1.SchemeGroupVersion.Version
	obj.Name = name
	obj.Spec.Finalizers = []kapi.FinalizerName{projectapi.FinalizerOrigin, kapi.FinalizerKubernetes, qyapi.FinalizerVender}
	for _, v := range finalizers {
		obj.Spec.Finalizers = append(obj.Spec.Finalizers, kapi.FinalizerName(v))
	}
	return CreateProjectWith(obj)
}

func CreateProjectWith(obj *projectapi.Project) ([]byte, *projectapi.Project, error) {
	return createProject(nil, obj)
}

func CreateProjectFromArbitray(data []byte) ([]byte, *projectapi.Project, error) {
	return createProject(data, nil)
}

func createProject(data []byte, obj *projectapi.Project) ([]byte, *projectapi.Project, error) {
	if len(data) == 0 && obj == nil {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if len(data) == 0 && obj != nil {
		result, err := oc.Projects().Create(obj)
		if err != nil {
			if retry := strings.EqualFold("encoding is not allowed for this codec: *recognizer.decoder", err.Error()) || strings.HasPrefix(err.Error(), "no kind is registered for the type api."); !retry {
				glog.Errorf("Could not access openshift: %s", err)
				return nil, nil, err
			}
		}
		if result == nil {
			glog.V(7).Infoln("Unexpected creation: %+v", obj)
			return nil, nil, errUnexpected
		}
		if result != nil {
			if strings.EqualFold("Project", result.Kind) && len(result.Name) > 0 {
				b := new(bytes.Buffer)
				if err := codec.JSON.Encode(b).One(result); err != nil {
					glog.Errorf("Could not encode runtime object: %s", err)
					return nil, result, err
				}
				logger.Printf("Project: %+v\n", b.String())
				return b.Bytes(), result, nil
			}
		}

		//data = make([]byte, 0)
		//b := bytes.NewBuffer(data)
		b := new(bytes.Buffer)
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
		data = b.Bytes()
	}

	raw, err := oc.RESTClient.Post().Resource("projects").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}

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
			return raw, nil, fmt.Errorf("Could not create project: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result := new(projectapi.Project)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	logger.Printf("Project: %+v\n", string(raw))
	return raw, result, nil
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

func DeleteProject(name string) error {
	if len(name) == 0 {
		return errNotFound
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %+v", err)
		return err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if err := oc.Projects().Delete(name); err != nil {
		glog.Errorf("Could not delete project %s: %+v", name, err)
		return err
	}
	return nil
}

// gitRef: branch name, tag name, or commit revision
func CreateBuild(name, projectName string, gitSecret map[string]string, gitURI, gitRef, contextDir string, sourceImages []map[string]interface{}, dockerfile string, buildSecrets []map[string]interface{}, buildStrategy map[string]interface{}) ([]byte, *buildapi.Build, error) {
	obj := &buildapi.Build{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Build",
			APIVersion: buildapiv1.SchemeGroupVersion.Version,
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:              name,
			Namespace:         projectName,
			CreationTimestamp: unversioned.Now(),
			Labels:            map[string]string{buildapi.BuildConfigLabel: "tangfeixiong"},
			Annotations:       map[string]string{buildapi.BuildNumberAnnotation: "1"},
		},
		Spec: buildapi.BuildSpec{
			ServiceAccount: builderServiceAccount,
			Source: buildapi.BuildSource{
				//Binary : &buildapi.BinaryBuildSource {},
				Dockerfile: &dockerfile,
				Git: &buildapi.GitBuildSource{
					URI: gitURI,
					Ref: gitRef,
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
			Strategy: buildapi.BuildStrategy{
				DockerStrategy: &buildapi.DockerBuildStrategy{
					From: &kapi.ObjectReference{
						Kind: "DockerImage",
						Name: "alpine:edge",
					},
					//PullSecret: &kapi.LocalObjectReference{
					//	Name: dockerPullSecret,
					//},
					NoCache: false,
					//Env : []kapi.EnvVar {},
					ForcePull: false,
					//DockerfilePath : ".",
				},
			},
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
	return CreateBuildWith(obj)
}

func CreateBuildWith(obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	return createBuild(nil, obj)
}

func CreateBuildFromArbitray(data []byte) ([]byte, *buildapi.Build, error) {
	return createBuild(data, nil)
}

func createBuild(data []byte, obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	if len(data) == 0 && obj == nil || obj != nil && len(obj.Namespace) == 0 {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %s", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if len(data) == 0 && obj != nil {
		result, err := oc.Builds(obj.Namespace).Create(obj)
		if err != nil {
			if retry := strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder") || strings.HasPrefix(err.Error(), "no kind is registered for the type api."); !retry {
				glog.Errorf("Could not access openshift: %s", err)
				return nil, nil, err
			}
		}
		if result == nil {
			glog.V(7).Infoln("Unexpected creation: %+v", obj)
			return nil, nil, errUnexpected
		}
		if result != nil {
			if strings.EqualFold("Build", result.Kind) && len(result.Name) > 0 {
				b := new(bytes.Buffer)
				if err := codec.JSON.Encode(b).One(result); err != nil {
					glog.Errorf("Could not encode runtime object: %s", err)
					return nil, result, err
				}
				logger.Printf("Build: %+v\n", b.String())
				return b.Bytes(), result, nil
			}
		}

		//data = make([]byte, 0)
		//b := bytes.NewBuffer(data)
		//b := new(bytes.Buffer)
		//if err := codec.JSON.Encode(b).One(obj); err != nil {
		//	glog.Errorf("Could not serialize runtime object: %+v", err)
		//	return nil, nil, err
		//}
		//data = b.Bytes()
		kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, &buildapi.Build{})
		if data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), obj, buildapiv1.SchemeGroupVersion); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
	}

	if obj == nil {
		hco, err := codec.JSON.Decode(data).One()
		if err != nil {
			glog.Errorf("Could not create helm object: %s", err)
			return nil, nil, err
		}
		obj := new(buildapi.Build)
		if err := hco.Object(obj); err != nil {
			glog.Errorf("Could not deserialize into runtime object: %s", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Post().Namespace(obj.Namespace).Resource("builds").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}

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
	if ok := strings.EqualFold("Build", meta.Kind) && len(meta.Name) > 0; !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not know metadata: %+v", meta)
				return raw, nil, err
			}
			glog.Warningf("Could not create build: %+v", status.Message)
			return raw, nil, fmt.Errorf("Could not create build: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result := new(buildapi.Build)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	logger.Printf("Build: %+v\n", string(raw))
	return raw, result, nil
}

func options(c *cobra.Command) {
	if val := c.Flags().Lookup("server"); val != nil {
		val.Value.Set("https://172.17.4.50:30448")
	} else {
		val = c.Flags().VarPF(stringValue{"https://172.17.4.50:30448"}, "server", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag server: %+v\n", val)
	}
	c.Flags().Lookup("server").NoOptDefVal = "https://172.17.4.50:30448"

	if val := c.Flags().Lookup("client-certificate"); val != nil {
		val.Value.Set(oclientcrt)
	} else {
		val = c.Flags().VarPF(stringValue{oclientcrt}, "client-certificate", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag server: %+v\n", val)
	}
	c.Flags().Lookup("client-certificate").NoOptDefVal = oclientcrt
	if val := c.Flags().Lookup("client-key"); val != nil {
		val.Value.Set(oclientkey)
	} else {
		val = c.Flags().VarPF(stringValue{oclientkey}, "client-key", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] client-key: %+v\n", val)
	}
	c.Flags().Lookup("client-key").NoOptDefVal = oclientkey

	if val := c.Flags().Lookup("api-version"); val != nil {
		val.Value.Set(apiVersion)
	} else {
		val = c.Flags().VarPF(stringValue{apiVersion}, "api-version", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] api version: %+v\n", val)
	}
	c.Flags().Lookup("api-version").NoOptDefVal = apiVersion

	if val := c.Flags().Lookup("api-version"); val != nil {
		val.Value.Set(apiVersion)
	} else {
		val = c.Flags().VarPF(stringValue{apiVersion}, "api-version", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] api version: %+v\n", val)
	}
	c.Flags().Lookup("api-version").NoOptDefVal = apiVersion

	if val := c.Flags().Lookup("certificate-authority"); val != nil {
		val.Value.Set(oca)
	} else {
		val = c.Flags().VarPF(stringValue{oca}, "certificate-authority", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] ca: %+v\n", val)
	}
	c.Flags().Lookup("certificate-authority").NoOptDefVal = oca

	if val := c.Flags().Lookup("insecure-skip-tls-verify"); val != nil {
		val.Value.Set("false")
	} else {
		val = c.Flags().VarPF(boolValue{false}, "insecure-skip-tls-verify", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] insecure tls: %+v\n", val)
	}
	c.Flags().Lookup("insecure-skip-tls-verify").NoOptDefVal = "false"

	if val := c.Flags().Lookup("config"); val != nil {
		val.Value.Set(oconfigPath)
	} else {
		val = c.Flags().VarPF(stringValue{oconfigPath}, "config", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] config: %+v\n", val)
	}
	c.Flags().Lookup("config").NoOptDefVal = oconfigPath

}

func DoBasicAuth() error {
	clientConfig := &restclient.Config{}
	serverNormalized, err := config.NormalizeServerURL("https://172.17.4.50:30448")
	if err != nil {
		return err
	}
	clientConfig.Host = serverNormalized
	clientConfig.CAFile = oca
	clientConfig.CertFile = oclientcrt
	clientConfig.KeyFile = oclientkey
	clientConfig.GroupVersion = &unversioned.GroupVersion{Group: "", Version: "v1"}
	clientConfig.APIPath = "/oapi"
	clientConfig.Codec = kapi.Codecs.LegacyCodec(*clientConfig.GroupVersion, kapi.SchemeGroupVersion)
	//clientConfig.Codec = kapi.Codecs.CodecForVersions(json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil), kapi.Codecs.UniversalDeserializer(), []unversioned.GroupVersion{*clientConfig.GroupVersion}, []unversioned.GroupVersion{*clientConfig.GroupVersion})
	//clientConfig.Codec = kapi.Codecs.CodecForVersions(json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil), []unversioned.GroupVersion{*clientConfig.GroupVersion}, []unversioned.GroupVersion{*clientConfig.GroupVersion})
	logger.Printf("simple config: %+v\n", clientConfig)

	//clientConfig.Username = "tangfeixiong"
	//clientConfig.Password = "tangfeixiong"
	//	token, err := tokencmd.RequestToken(clientConfig, os.Stdin, clientConfig.Username, clientConfig.Password)
	//	if err != nil {
	//		return err
	//	}
	//	logger.Printf("current token: %+v\n", token)
	//	clientConfig.BearerToken = token

	clientConfig.BearerToken = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"
	clientConfig.Username = ""
	clientConfig.Password = ""

	clientK8s, err := restclient.RESTClientFor(clientConfig)
	if err != nil {
		return err
	}

	osClient := &client.Client{clientK8s}

	me, err := osClient.Projects().List(kapi.ListOptions{})
	if err != nil {
		return err
	}

	logger.Printf("Result: %+v\n", me)

	return nil
}

func overrideRootCommand() {

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	//f := osclientcmd.New(flags)
	cmds := cli.NewCommandCLI("oc", "oc", os.Stdin, os.Stdout, os.Stderr)
	cmds.Aliases = []string{"oc"}
	cmds.Use = "oc"
	cmds.Short = "openshift client"
	flags.VisitAll(func(flag *pflag.Flag) {
		if f := cmds.PersistentFlags().Lookup(flag.Name); f == nil {
			glog.V(5).Infof("flag: %v", flag.Name)
			cmds.PersistentFlags().AddFlag(flag)
		} else {
			glog.V(5).Infof("already registered flag %s", flag.Name)
		}
	})
	//cmds.PersistentFlags().Var(flags.Lookup("config").Value, "config", "Specify a kubeconfig file to define the configuration")
	if val := cmds.PersistentFlags().Lookup("config"); val != nil {
		fmt.Println("configed")
		val.Value.Set(oconfigPath)
	} else {
		fmt.Println("setting")
		val = cmds.PersistentFlags().VarPF(stringValue{oconfigPath}, "config", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] config: %+v\n", val)
	}
	if val := cmds.PersistentFlags().Lookup("loglevel"); val != nil {
		fmt.Println("configed")
		val.Value.Set("10")
	} else {
		fmt.Println("setting")
		val = cmds.PersistentFlags().VarPF(intValue{5}, "loglevel", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] loglevel: %+v\n", val)
	}
	templates.ActsAsRootCommand(cmds, []string{"option"})
	cmds.AddCommand(cmd.NewCmdOptions(os.Stdout))

	rootCommand = cmds
}

func Whole() (*userapi.UserList, error) {
	pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")

	var whole *userapi.UserList
	var err error
	rootCommand.Run = func(c *cobra.Command, args []string) {
		f := osclientcmd.New(pf)
		oc, _, err := f.Clients()
		if err != nil {
			whole = nil
			return
		}
		whole, err = oc.Users().List(kapi.ListOptions{})
		if err != nil {
			whole = nil
			return
		}
	}

	if err = rootCommand.Execute(); err != nil {
		return nil, err
	}
	return whole, nil
}

func ShowMe() error {
	//rootCommand.SetArgs([]string{"version"})
	if err := rootCommand.Execute(); err != nil {
		glog.V(5).Infof("Failed: %v", err)
		return err
	}
	return nil
}

func fakeWhoAmI() (*userapi.User, error) {
	pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")

	var me *userapi.User = nil
	var err error = nil
	rootCommand.Run = func(c *cobra.Command, args []string) {
		f := osclientcmd.New(pf)
		oc, _, err := f.Clients()
		if err != nil {
			me = nil
			return
		}
		me, err = oc.Users().Get("~")
		if err != nil {
			me = nil
			return
		}
	}

	if err = rootCommand.Execute(); err != nil {
		return nil, err
	}
	return me, err
}

func LoginWithBasicAuth(username, password string) error {
	//pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")
	f := factory
	c := cmd.NewCmdLogin("oc", f, os.Stdin, os.Stdout)
	options(c)

	if val := c.Flags().Lookup("username"); val != nil {
		val.Value.Set(username)
	} else {
		val = c.Flags().VarPF(stringValue{username}, "username", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag username: %+v\n", val)
	}
	c.Flags().Lookup("username").NoOptDefVal = username

	if val := c.Flags().Lookup("password"); val != nil {
		if err := c.Flags().Set("password", password); err != nil {
			fmt.Fprintf(os.Stdout, "[tangfx] flag password err: %+v\n", err)
			return err
		}
	} else {
		val = c.Flags().VarPF(stringValue{password}, "password", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] flag password: %+v\n", f)
	}
	c.Flags().Lookup("password").NoOptDefVal = password

	if val := c.Flags().Lookup("token"); val != nil {
		val.Value.Set("")
	} else {
		val = c.Flags().VarPF(stringValue{""}, "token", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] token: %+v\n", val)
	}
	c.Flags().Lookup("token").NoOptDefVal = ""

	//c.SetArgs([]string{"172.17.4.50:30448"})
	c.SetArgs([]string{})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not login with basic auth: %v\n", err)
		return err
	}
	return nil
}

func ProjectWithSimple(name string, in io.Reader, out, errOut io.Writer) {
	f := osclientcmd.NewFactory(withOClientConfig())
	fullName := "oc"
	c := cmd.NewCmdRequestProject(fullName, "new-project", fullName+" login", fullName+" project", f, out)
	c.SetArgs([]string{name})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not create project: %v\n", err)
	}
}

func BuildWithConfig(name string, in io.Reader, out, errOut io.Writer) {
	f := osclientcmd.NewFactory(withOClientConfig())
	c := NewCmdStartBuild(name, f, in, out)
	c.SetArgs([]string{name})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not start build with config: %v\n", err)
	}
}

// NewCmdNewBuild implements the OpenShift cli new-build command
func NewCmdNewBuild(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	config := newcmd.NewAppConfig()
	config.ExpectToBuild = true
	config.AddEnvironmentToBuild = true
	options := &cmd.NewBuildOptions{Config: config}

	cmd := &cobra.Command{
		Use:        "new-build (IMAGE | IMAGESTREAM | PATH | URL ...)",
		Short:      "Create a new build configuration",
		Long:       fmt.Sprintf(newBuildLong, fullName),
		Example:    fmt.Sprintf(newBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete(fullName, f, c, args, out, in))
			err := options.Run()
			if err == cmdutil.ErrExit {
				os.Exit(1)
			}
			kcmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringSliceVar(&config.SourceRepositories, "code", config.SourceRepositories, "Source code in the build configuration.")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image", "", config.ImageStreams, "Name of an image stream to to use as a builder. (deprecated)")
	cmd.Flags().MarkDeprecated("image", "use --image-stream instead")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image-stream", "i", config.ImageStreams, "Name of an image stream to to use as a builder.")
	cmd.Flags().StringSliceVar(&config.DockerImages, "docker-image", config.DockerImages, "Name of a Docker image to use as a builder.")
	cmd.Flags().StringSliceVar(&config.Secrets, "build-secret", config.Secrets, "Secret and destination to use as an input for the build.")
	cmd.Flags().StringVar(&config.Name, "name", "", "Set name to use for generated build artifacts.")
	cmd.Flags().StringVar(&config.To, "to", "", "Push built images to this image stream tag (or Docker image repository if --to-docker is set).")
	cmd.Flags().BoolVar(&config.OutputDocker, "to-docker", false, "Have the build output push to a Docker repository.")
	cmd.Flags().StringSliceVarP(&config.Environment, "env", "e", config.Environment, "Specify key value pairs of environment variables to set into resulting image.")
	cmd.Flags().StringVar(&config.Strategy, "strategy", "", "Specify the build strategy to use if you don't want to detect (docker|source).")
	cmd.Flags().StringVarP(&config.Dockerfile, "dockerfile", "D", "", "Specify the contents of a Dockerfile to build directly, implies --strategy=docker. Pass '-' to read from STDIN.")
	cmd.Flags().BoolVar(&config.BinaryBuild, "binary", false, "Instead of expecting a source URL, set the build to expect binary contents. Will disable triggers.")
	cmd.Flags().StringP("labels", "l", "", "Label to set in all generated resources.")
	cmd.Flags().BoolVar(&config.AllowMissingImages, "allow-missing-images", false, "If true, indicates that referenced Docker images that cannot be found locally or in a registry should still be used.")
	cmd.Flags().BoolVar(&config.AllowMissingImageStreamTags, "allow-missing-imagestream-tags", false, "If true, indicates that image stream tags that don't exist should still be used.")
	cmd.Flags().StringVar(&config.ContextDir, "context-dir", "", "Context directory to be used for the build.")
	cmd.Flags().BoolVar(&config.DryRun, "dry-run", false, "If true, do not actually create resources.")
	cmd.Flags().BoolVar(&config.NoOutput, "no-output", false, "If true, the build output will not be pushed anywhere.")
	cmd.Flags().StringVar(&config.SourceImage, "source-image", "", "Specify an image to use as source for the build.  You must also specify --source-image-path.")
	cmd.Flags().StringVar(&config.SourceImagePath, "source-image-path", "", "Specify the file or directory to copy from the source image and its destination in the build directory. Format: [source]:[destination-dir].")
	kcmdutil.AddPrinterFlags(cmd)

	return cmd
}

func NewCmdStartBuild(fullName string, f *osclientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	o := &cmd.StartBuildOptions{
		LogLevel:        "5",
		Follow:          true,
		WaitForComplete: false,
	}

	cmd := &cobra.Command{
		Use:        "start-build (BUILDCONFIG | --from-build=BUILD)",
		Short:      "Start a new build",
		Long:       startBuildLong,
		Example:    fmt.Sprintf(startBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, in, out, cmd, args))
			kcmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVar(&o.LogLevel, "build-loglevel", o.LogLevel, "Specify the log level for the build log output")
	cmd.Flags().Lookup("build-loglevel").NoOptDefVal = "5"
	cmd.Flags().StringSliceVarP(&o.Env, "env", "e", o.Env, "Specify key value pairs of environment variables to set for the build container.")

	cmd.Flags().StringVar(&o.FromBuild, "from-build", o.FromBuild, "Specify the name of a build which should be re-run")

	cmd.Flags().BoolVar(&o.Follow, "follow", o.Follow, "Start a build and watch its logs until it completes or fails")
	cmd.Flags().Lookup("follow").NoOptDefVal = "true"
	cmd.Flags().BoolVar(&o.WaitForComplete, "wait", o.WaitForComplete, "Wait for a build to complete and exit with a non-zero return code if the build fails")

	cmd.Flags().StringVar(&o.FromFile, "from-file", o.FromFile, "A file to use as the binary input for the build; example a pom.xml or Dockerfile. Will be the only file in the build source.")
	cmd.Flags().StringVar(&o.FromDir, "from-dir", o.FromDir, "A directory to archive and use as the binary input for a build.")
	cmd.Flags().StringVar(&o.FromRepo, "from-repo", o.FromRepo, "The path to a local source code repository to use as the binary input for a build.")
	cmd.Flags().StringVar(&o.Commit, "commit", o.Commit, "Specify the source code commit identifier the build should use; requires a build based on a Git repository")

	cmd.Flags().StringVar(&o.ListWebhooks, "list-webhooks", o.ListWebhooks, "List the webhooks for the specified build config or build; accepts 'all', 'generic', or 'github'")
	cmd.Flags().StringVar(&o.FromWebhook, "from-webhook", o.FromWebhook, "Specify a webhook URL for an existing build config to trigger")

	cmd.Flags().StringVar(&o.GitPostReceive, "git-post-receive", o.GitPostReceive, "The contents of the post-receive hook to trigger a build")
	cmd.Flags().StringVar(&o.GitRepository, "git-repository", o.GitRepository, "The path to the git repository for post-receive; defaults to the current directory")

	// cmdutil.AddOutputFlagsForMutation(cmd)
	return cmd

}
