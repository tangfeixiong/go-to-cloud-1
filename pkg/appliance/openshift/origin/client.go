package origin

import (
	"bytes"
	"errors"
	"sync"
	"time"
	//"flag"
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
	//"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	//"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	//cmdutil "github.com/openshift/origin/pkg/cmd/util"
	osclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	// "github.com/openshift/origin/pkg/cmd/util/tokencmd"
	//"github.com/openshift/origin/pkg/generate/git"
	userapi "github.com/openshift/origin/pkg/user/api"
	userapiv1 "github.com/openshift/origin/pkg/user/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/restclient"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
	// clientauth "k8s.io/kubernetes/pkg/client/unversioned/auth"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	// clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	// kubecmdconfig "k8s.io/kubernetes/pkg/kubectl/cmd/config"
	//"k8s.io/kubernetes/pkg/runtime"
	//"k8s.io/kubernetes/pkg/runtime/serializer/json"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/build-builder"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

var (
	rootCommand *cobra.Command = utility.RootCommand

	logger *log.Logger = utility.Logger

	errNotFound       error = errors.New("Not found")
	errNotImplemented error = errors.New("Not implemented")
	errUnexpected     error = errors.New("Unexpected")

	kubeconfigPath    string = "/data/src/github.com/openshift/origin/etc/kubeconfig"
	kubeconfigContext string = "openshift-origin-single"

	apiVersion string = "v1"
	oca        string = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	oclientcrt string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	oclientkey string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"

	// token string = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"

	oconfigPath    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
	oconfigContext string = "default/172-17-4-50:30443/system:admin"

	kClientConfig, oClientConfig *clientcmd.ClientConfig
	kClient                      *kclient.Client
	oClient                      *client.Client
	kConfig, oConfig             *restclient.Config

	factory *osclientcmd.Factory

	who *userapi.User

	builderServiceAccount string = "builder"

	overrideDockerfile string = "FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"
	githubURI          string = "https://github.com/tangfeixiong/docker-nc.git"
	githubRef          string = "master"
	githubPath         string = "latest"
	githubSecret       string = "github-qingyuancloud-tangfx"
	dockerPullSecret   string = "localdockerconfig"
	dockerPushSecret   string = "localdockerconfig"
	timeout            int64  = 900
)

func init() {
	//	flagtypes.GLog(rootCommand.PersistentFlags())
	if v, ok := os.LookupEnv("KUBE_CONFIG"); ok {
		if v != "" {
			kubeconfigPath = v
		}
	}
	if v, ok := os.LookupEnv("KUBE_CONTEXT"); ok {
		if v != "" {
			kubeconfigContext = v
		}
	}
	if v, ok := os.LookupEnv("OSO_CONFIG"); ok {
		if v != "" {
			oconfigPath = v
		}
	}
	if v, ok := os.LookupEnv("OSO_CONTEXT"); ok {
		if v != "" {
			oconfigContext = v
		}
	}
}

func DirectlyRunOriginDockerBuilder(data *buildapiv1.Build) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, DirectlyRunOriginDockerBuilder] ")

	var raw []byte
	var hco *codec.Object
	var err error

	obj := &buildapi.Build{}
	buf := &bytes.Buffer{}
	if err = codec.JSON.Encode(buf).One(data.TypeMeta); err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}
	glog.V(10).Infoln(string(buf.Bytes()))
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}
	obj.TypeMeta = unversioned.TypeMeta{}
	if err = hco.Object(&obj.TypeMeta); err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(data.ObjectMeta); err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}
	glog.V(10).Infoln(string(buf.Bytes()))
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}
	obj.ObjectMeta = kapi.ObjectMeta{}
	if err = hco.Object(&obj.ObjectMeta); err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(data.Spec); err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}
	glog.V(10).Infoln(string(buf.Bytes()))
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}
	obj.Spec = buildapi.BuildSpec{}
	if err = hco.Object(&obj.Spec); err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(data.Status); err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}
	glog.V(10).Infoln(string(buf.Bytes()))
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}
	obj.Status = buildapi.BuildStatus{}
	if err = hco.Object(&obj.Status); err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}

	b := new(bytes.Buffer)
	ccf := util.NewClientCmdFactory()

	obj, err = builder.RunDockerBuild(b, obj, ccf)
	if err != nil {
		logger.Printf("Could not docker build: %+v\n", err)
		return raw, nil, err
	}

	//c := make(chan error, 1)
	go func() {
		var timeout bool
		var offset int = 0
		var m *sync.Mutex = &sync.Mutex{}
		go func() {
			select {
			//case e := <- c:

			case <-time.After(time.Duration(900) * time.Second):
				m.Lock()
				defer m.Unlock()
				timeout = true
			}
		}()
		for false == timeout {
			l := b.Len()
			if l > offset {
				fmt.Print(string(b.Next(l - offset)))
				offset = l
			}
		}
	}()

	time.Sleep(time.Duration(500))

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(obj.TypeMeta); err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}
	data.TypeMeta = unversioned.TypeMeta{}
	if err = hco.Object(&data.TypeMeta); err != nil {
		logger.Printf("Could not validate type meta: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(obj.ObjectMeta); err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}
	data.ObjectMeta = kapiv1.ObjectMeta{}
	if err = hco.Object(&data.ObjectMeta); err != nil {
		logger.Printf("Could not validate object meta: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(obj.Spec); err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}
	data.Spec = buildapiv1.BuildSpec{}
	if err = hco.Object(&data.Spec); err != nil {
		logger.Printf("Could not validate build spec: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(obj.Status); err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}
	hco, err = codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}
	data.Status = buildapiv1.BuildStatus{}
	if err = hco.Object(&data.Status); err != nil {
		logger.Printf("Could not validate build status: %+v\n", err)
		return raw, nil, err
	}

	buf.Reset()
	if err = codec.JSON.Encode(buf).One(data); err != nil {
		logger.Printf("Could not validate object: %+v\n", err)
		return raw, nil, err
	}
	raw = buf.Bytes()

	if b.Len() > 0 {
		//data.Status.Phase = buildapiv1.BuildPhaseRunning
		data.Status.Message += b.String()
	}

	return raw, data, nil
}

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

type stringSliceValue struct {
	value []string
}

func (val stringSliceValue) String() string {
	return strings.Join(val.value, ",")
}
func (val stringSliceValue) Set(v string) error {
	val.value = strings.Split(v, ",")
	return nil
}
func (val stringSliceValue) Type() string {
	return "[]string"
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
	logger = log.New(os.Stdout, "[appliance/openshift/origin, createUser] ", log.LstdFlags|log.Lshortfile)

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
	logger = log.New(os.Stdout, "[appliance/openshift/origin, RetrieveUsers] ", log.LstdFlags|log.Lshortfile)

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
	logger = log.New(os.Stdout, "[appliance/openshift/origin, RetrieveUser] ", log.LstdFlags|log.Lshortfile)

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

func DoBasicAuth() error {
	logger = log.New(os.Stdout, "[appliance/openshift/origin, DoBasicAuth] ", log.LstdFlags|log.Lshortfile)

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
	logger = log.New(os.Stdout, "[appliance/openshift/origin, LoginWithBasicAuth] ", log.LstdFlags|log.Lshortfile)

	//pf := rootCommand.PersistentFlags()
	//pf.StringVarP(&addr, "listen", "l", ":44134", "The address:port to listen on")
	//pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")
	f := factory
	c := cmd.NewCmdLogin("oc", f, os.Stdin, os.Stdout)
	overrideOptions(c)

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
	logger = log.New(os.Stdout, "[appliance/openshift/origin, ProjectWithSimple] ", log.LstdFlags|log.Lshortfile)

	f := osclientcmd.NewFactory(withOClientConfig())
	fullName := "oc"
	c := cmd.NewCmdRequestProject(fullName, "new-project", fullName+" login", fullName+" project", f, out)
	c.SetArgs([]string{name})
	if err := c.Execute(); err != nil {
		logger.Printf("Could not create project: %v\n", err)
	}
}

func overrideStringFlag(c *cobra.Command, name, value, shorthand, usage, noOptDefVal string) {
	logger.SetPrefix("[appliance/origin, overrideStringFlag] ")

	v := c.Flags().Lookup(name)
	if v != nil {
		v.Value.Set(value)
	} else {
		v = c.Flags().VarPF(stringValue{value}, name, shorthand, usage)
	}
	if noOptDefVal != "" {
		v.NoOptDefVal = noOptDefVal
	}
	logger.Printf("override flag %s: %+v\n", name, v)
}

func overrideStringSliceFlag(c *cobra.Command, name string, value []string, shorthand, usage string, noOptDefVal []string) {
	logger.SetPrefix("[appliance/origin, overrideStringSliceFlag] ")

	s := strings.Join(value, ",")
	v := c.Flags().Lookup(name)
	if v != nil {
		v.Value.Set(s)
	} else {
		v = c.Flags().VarPF(stringSliceValue{value}, name, shorthand, usage)
	}
	if len(noOptDefVal) > 0 {
		v.NoOptDefVal = strings.Join(noOptDefVal, ",")
	}
	logger.Printf("override flag %s: %+v\n", name, v)
}

func overrideBoolFlag(c *cobra.Command, name string, value bool, shorthand, usage string, noOptDefVal bool) {
	logger.SetPrefix("[appliance/origin, overrideBoolFlag] ")

	v := c.Flags().Lookup(name)
	if v != nil {
		v.Value.Set(strconv.FormatBool(value))
	} else {
		v = c.Flags().VarPF(boolValue{value}, name, shorthand, usage)
	}
	if noOptDefVal {
		v.NoOptDefVal = strconv.FormatBool(noOptDefVal)
	}
	logger.Printf("override flag %s: %+v\n", name, v)
}

func overrideIntFlag(c *cobra.Command, name string, value int, shorthand, usage string, noOptDefVal int) {
	logger.SetPrefix("[appliance/origin, overrideIntFlag] ")

	v := c.Flags().Lookup(name)
	if v != nil {
		v.Value.Set(strconv.Itoa(value))
	} else {
		v = c.Flags().VarPF(intValue{value}, name, shorthand, usage)
	}
	if noOptDefVal != 0 {
		v.NoOptDefVal = strconv.Itoa(noOptDefVal)
	}
	logger.Printf("override flag %s: %+v\n", name, v)
}

func overrideOptions(c *cobra.Command) {
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

func overrideRootCommandArgs() {
	logger.SetOutput(os.Stdout)
	logger.SetPrefix("[appliance/openshift/origin, overrideRootCommandArgs] ")

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	//f := osclientcmd.New(flags)
	//rootCommand := cli.NewCommandCLI("oc", "oc", os.Stdin, os.Stdout, os.Stderr)
	//rootCommand.Aliases = []string{"oc"}
	//rootCommand.Use = "oc"
	//rootCommand.Short = "openshift origin client"
	flags.VisitAll(func(flag *pflag.Flag) {
		if f := rootCommand.PersistentFlags().Lookup(flag.Name); f == nil {
			glog.V(5).Infof("flag: %v", flag.Name)
			rootCommand.PersistentFlags().AddFlag(flag)
		} else {
			glog.V(5).Infof("already registered flag %s", flag.Name)
		}
	})
	//rootCommand.PersistentFlags().Var(flags.Lookup("config").Value, "config", "Specify a kubeconfig file to define the configuration")
	if val := rootCommand.PersistentFlags().Lookup("config"); val != nil {
		fmt.Println("configed")
		val.Value.Set(oconfigPath)
	} else {
		fmt.Println("setting")
		val = rootCommand.PersistentFlags().VarPF(stringValue{oconfigPath}, "config", "", "")
		logger.Printf("config: %+v\n", val)
	}
	if val := rootCommand.PersistentFlags().Lookup("loglevel"); val != nil {
		fmt.Println("configed")
		val.Value.Set("10")
	} else {
		fmt.Println("setting")
		val = rootCommand.PersistentFlags().VarPF(intValue{5}, "loglevel", "", "")
		logger.Printf("loglevel: %+v\n", val)
	}
	templates.ActsAsRootCommand(rootCommand, []string{"option"})
	rootCommand.AddCommand(cmd.NewCmdOptions(os.Stdout))
}
