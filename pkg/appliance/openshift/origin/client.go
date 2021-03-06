package origin

import (
	"bytes"
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

	"github.com/cloudfoundry/yagnats"
	"github.com/docker/engine-api/types"
	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	oclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/cli/config"
	//"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
	oclientcmd "github.com/openshift/origin/pkg/cmd/util/clientcmd"
	// "github.com/openshift/origin/pkg/cmd/util/tokencmd"
	//"github.com/openshift/origin/pkg/generate/git"
	//"github.com/openshift/origin/pkg/cmd/cli"
	projectapi "github.com/openshift/origin/pkg/project/api"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"
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
	"k8s.io/kubernetes/pkg/runtime"
	//"k8s.io/kubernetes/pkg/runtime/serializer/json"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/etcd"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/kubernetes"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/build-builder"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

var (
	rootCommand *cobra.Command = utility.RootCommand

	kubeconfig_path string = "/data/src/github.com/openshift/origin/etc/kubeconfig"
	kubectl_context string = "openshift-origin-single"

	apiVersion string = "v1"
	oca        string = "/data/src/github.com/openshift/origin/openshift.local.config/master/ca.crt"
	oclientcrt string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.crt"
	oclientkey string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.key"

	// token string = "IqEFJ7eK2_Pls4JHItvMPLBqGcuct5ogPN6NrapH20s"

	osoconfig_path string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
	oc_context     string = "default/172-17-4-50:30443/system:admin"

	kClientConfig, oClientConfig *clientcmd.ClientConfig
	kClient                      *kclient.Client
	oClient                      *oclient.Client
	kConfig, oConfig             *restclient.Config

	factory *oclientcmd.Factory

	who *userapi.User

	builderServiceAccount string = "builder"

	overrideDockerfile string = "FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"
	githubURI          string = "https://github.com/tangfeixiong/docker-nc.git"
	githubRef          string = "master"
	githubPath         string = "latest"
	githubSecret       string = "github-qingyuancloud-tangfx"
	dockerPullSecret   string = "localdockerconfig"
	dockerPushSecret   string = "localdockerconfig"
	timeout            int64  = 1200

	etcdv3_addresses []string = []string{"10.3.0.212:2379"}

	gnatsdAddresses []string = []string{"10.3.0.39:4222"}
	gnatsdUsername  string   = "derek"
	gnatsdPassword  string   = "T0pS3cr3t"
)

func init() {
	//	flagtypes.GLog(rootCommand.PersistentFlags())
	if v, ok := os.LookupEnv("KUBE_CONFIG"); ok {
		kubeconfig_path = v
	}
	if v, ok := os.LookupEnv("KUBE_CONTEXT"); ok {
		kubectl_context = v
	}
	if v, ok := os.LookupEnv("OSO_CONFIG"); ok {
		osoconfig_path = v
	}
	if v, ok := os.LookupEnv("OSO_CONTEXT"); ok {
		oc_context = v
	}
	if v, ok := os.LookupEnv("ETCD_V3_ADDRESSES"); ok {
		etcdv3_addresses = strings.Split(v, ",")
	}
}

type PaaS struct {
	kubeconfigPath  string
	kubectlContext  string
	osoconfigPath   string
	ocContext       string
	ccf             *oclientcmd.Factory
	oc              *oclient.Client
	kc              *kclient.Client
	orchestra       *kubernetes.Orchestration
	err             error
	etcdctl         *etcd.V3ClientContext
	WaitForComplete bool
	Follow          bool
	In              io.Reader
	Out             io.Writer
	ErrOut          io.Writer
}

func NewPaaS() *PaaS {
	return NewPaaSWith(kubeconfig_path, kubectl_context, osoconfig_path, oc_context)
}

func NewPaaSWith(kubeconfigPath, kubectlContext, osoconfigPath, ocContext string) *PaaS {
	p := &PaaS{
		kubeconfigPath: kubeconfigPath,
		kubectlContext: kubectlContext,
		osoconfigPath:  osoconfigPath,
		ocContext:      ocContext,
	}
	p.ccf = util.NewClientCmdFactoryWith(kubeconfigPath, kubectlContext, osoconfigPath, ocContext)
	p.oc, p.kc, p.err = p.ccf.Clients()
	p.orchestra = kubernetes.NewOrchestrationWith(p.kc)
	projectapi.AddToScheme(kapi.Scheme)
	projectapiv1.AddToScheme(kapi.Scheme)
	buildapi.AddToScheme(kapi.Scheme)
	buildapiv1.AddToScheme(kapi.Scheme)
	userapi.AddToScheme(kapi.Scheme)
	userapiv1.AddToScheme(kapi.Scheme)
	return p
}

func (p *PaaS) Factory() *oclientcmd.Factory {
	return p.ccf
}

func (p *PaaS) OC() *oclient.Client {
	return p.oc
}

func (p *PaaS) KC() *kclient.Client {
	return p.kc
}

func (p *PaaS) WithOCctl(kubeconfigPath, kubectlContext, osoconfigPath, ocContext string) *PaaS {
	if p == nil {
		p = &PaaS{}
	}
	p.kubeconfigPath = kubeconfigPath
	p.kubectlContext = kubectlContext
	p.osoconfigPath = osoconfigPath
	p.ocContext = ocContext
	p.occtl()
	return p
}

func (p *PaaS) occtl() {
	p.ccf = util.NewClientCmdFactory()
	mapper, _ := p.ccf.Object(false)
	kapi.RegisterRESTMapper(mapper)
	p.oc, p.kc, p.err = p.ccf.Clients()
}

func (p *PaaS) WithEtcdCtl(addr []string, dialTimeout, requestTimeout time.Duration) *PaaS {
	if p == nil {
		p = &PaaS{}
	}
	p.etcdctl = etcd.NewV3ClientContext(addr, dialTimeout, requestTimeout)
	return p
}

func (p *PaaS) EtcdCtl() *etcd.V3ClientContext {
	return p.etcdctl
}

/*
  github.com/openshift/origin/pkg/cmd/cli/cmd/newbuild.go
*/
func (p *PaaS) OSO_startbuild_NewCmdNewBuild(bc *buildapi.BuildConfig, buildName string, binarySource io.Reader) (messageTarget io.Reader, err error) {
	cmd, o := NewCmdNewBuild("osoc", p.Factory(), os.Stdin, os.Stdout)
	r, w := io.Pipe()
	o.In = binarySource
	o.Out = w
	cmd.SetOutput(o.Out)
	o.ErrOut = cmd.Out()

	//o.NewBuildOptions.Env = []string{}
	//o.NewBuildOptions.WaitForComplete = false
	//o.NewBuildOptions.Follow = true
	//o.NewBuildOptions.Client = p.OC()
	o.NewBuildOptions.Config.OriginNamespace = bc.Namespace
	o.NewBuildOptions.Action.DryRun = false

	o.PaaS = p
	if err = o.complete("osoc", p.Factory(), cmd, []string{}, o.Out, o.In); err != nil {
		glog.Errorf("incorrect new-build settings: %+v", err)
		r.CloseWithError(nil)
		w.CloseWithError(nil)
		messageTarget = nil
		return
	}
	if err = o.runBuildConfig(bc, buildName); err != nil {
		glog.Errorf("failed to invoke new-build: %+v", err)
		r.CloseWithError(nil)
		w.CloseWithError(nil)
		messageTarget = nil
		return
	}
	messageTarget = r
	return
}

func (p *PaaS) OSO_startbuild_NewCmdStartBuild(bc *buildapiv1.BuildConfig, obj *buildapiv1.Build, binarySource io.Reader) (binaryMessage io.Reader, err error) {
	buildconfig := &buildapi.BuildConfig{}
	build := &buildapi.Build{}
	var data []byte
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, bc, obj)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, buildconfig, build)
	data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), bc)
	if err != nil {
		return nil, err
	}
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), data, buildconfig); err != nil {
		return nil, err
	}
	if obj != nil {
		data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), obj)
		if err != nil {
			return nil, err
		}
		if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), data, build); err != nil {
			return nil, err
		}
	}
	return p.oso_startbuild_NewCmdStartBuild(buildconfig, build, binarySource)
}

/*
  github.com/openshift/origin/pkg/cmd/cli/cmd/startbuild.go
*/
func (p *PaaS) oso_startbuild_NewCmdStartBuild(bc *buildapi.BuildConfig, obj *buildapi.Build, binarySource io.Reader) (binaryMessage io.Reader, err error) {
	cmd, o := NewCmdStartBuild("osoc", p.Factory(), os.Stdin, os.Stdout)
	r, w := io.Pipe()
	o.In = binarySource
	o.Out = w
	cmd.SetOutput(o.Out)
	o.ErrOut = cmd.Out()
	o.StartBuildOptions.Env = []string{}
	o.StartBuildOptions.WaitForComplete = false
	o.StartBuildOptions.Follow = true
	o.StartBuildOptions.Namespace = bc.Namespace
	o.StartBuildOptions.Client = p.OC()

	o.OP = p
	if err = o.completeBuildConfig(bc.Name, "5"); err != nil {
		glog.Errorf("incorrect start-build settings: %+v", err)
		r.CloseWithError(nil)
		w.CloseWithError(nil)
		binaryMessage = nil
		return
	}
	if err = o.Run(); err != nil {
		glog.Errorf("failed to invoke start-build: %+v", err)
		r.CloseWithError(nil)
		w.CloseWithError(nil)
		binaryMessage = nil
		return
	}

	//	u.Schedulers["DockerBuilder"].WithPaylodHandler(
	//		func() dispatcher.HandleFunc {
	//			glog.Errorf("Schedule docker builder into tracker: %s/%s(%s)\n", obj.Namespace, obj.Name, bc.Name)
	//			return o.TrackWith(ctx, req, resp, op, raw, obj, bc)
	//		}(),
	//	)
	binaryMessage = r
	return
}

func (p *PaaS) RequestProjectCreation(project string) error {
	ok, err := findProject(p.oc, project)
	if err != nil {
		return err
	}
	if !ok {
		tgt := &projectapiv1.Project{
			TypeMeta: unversioned.TypeMeta{
				Kind:       "Project",
				APIVersion: projectapiv1.SchemeGroupVersion.Version,
			},
			ObjectMeta: kapiv1.ObjectMeta{
				Name: project,
			},
			Spec: projectapiv1.ProjectSpec{
				Finalizers: []kapiv1.FinalizerName{projectapiv1.FinalizerOrigin,
					kapiv1.FinalizerKubernetes},
			},
		}
		_, _, err = createIntoProject(p.oc, nil, tgt)
		return err
	}
	return nil
}

func (p *PaaS) RequestBuilderSecretCreationWithDockerRegistry(project, secret, builder string, dac types.AuthConfig) error {
	_, _, _, err := p.orchestra.VerifyDockerConfigJsonSecretAndServiceAccount(project, secret, dac, builder)
	return err
}

func (p *PaaS) RequestBuildConfigCreation(rawJSON []byte) (data []byte, obj *buildapiv1.BuildConfig, err error) {
	obj, err = reapBuildConfig(rawJSON)
	if err != nil {
		glog.Errorf("Cloud not deserilize into object: %+V", err)
		data = make([]byte, 0)
		obj = nil
		err = fmt.Errorf("%+v: %+v", errUnexpected, err)
		return
	}
	glog.Infof("%+v", obj)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, obj)
	if obj.Spec.Output.PushSecret != nil {
		if obj.Spec.Strategy.DockerStrategy != nil && obj.Spec.Strategy.DockerStrategy.PullSecret != nil {
			if strings.Compare(obj.Spec.Output.PushSecret.Name, obj.Spec.Strategy.DockerStrategy.PullSecret.Name) == 0 {
				obj.Spec.Output.PushSecret = nil
			}
		} else if obj.Spec.Strategy.SourceStrategy != nil && obj.Spec.Strategy.SourceStrategy.PullSecret != nil {
			if strings.Compare(obj.Spec.Output.PushSecret.Name, obj.Spec.Strategy.SourceStrategy.PullSecret.Name) == 0 {
				obj.Spec.Output.PushSecret = nil
			}
		} else if obj.Spec.Strategy.CustomStrategy != nil && obj.Spec.Strategy.CustomStrategy.PullSecret != nil {
			if strings.Compare(obj.Spec.Output.PushSecret.Name, obj.Spec.Strategy.CustomStrategy.PullSecret.Name) == 0 {
				obj.Spec.Output.PushSecret = nil
			}
		}
	}
	rawJSON, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), obj)
	if err != nil {
		return
	}

	if len(obj.Namespace) > 0 {
		d, o, e := p.readProject(obj.Namespace)
		if e != nil {
			data = make([]byte, 0)
			obj = nil
			err = fmt.Errorf("%+v: %+v", errUnexpected, e)
			return
		}
		if len(d) == 0 || o == nil {
			d, o, e = p.createProjectRequest(obj.Namespace)
			if e != nil {
				data = make([]byte, 0)
				obj = nil
				err = fmt.Errorf("%+v: %+v", errUnexpected, e)
				return
			}
		}
	}

	var ok bool
	ok, err = findBuildConfig(p.oc, obj.Namespace, obj.Name)
	if err != nil {
		data = make([]byte, 0)
		obj = nil
		return
	}
	if ok {
		data = make([]byte, 0)
		obj = nil
		err = fmt.Errorf("%+v: build config conflict", errBadRequest)
		return
	}

	data, obj, err = p.createBuildConfigWithJSON(rawJSON, obj.Namespace)
	if err != nil {
		data = make([]byte, 0)
		obj = nil
		err = fmt.Errorf("%+v: %+v", errUnexpected, err)
	}
	return
}

func (p *PaaS) RequestBuildCreation(name, message string, conf *buildapiv1.BuildConfig) ([]byte, *buildapiv1.Build, *buildapiv1.BuildConfig, error) {
	var bc *buildapiv1.BuildConfig
	var ok bool
	var err error
	ok, err = findBuildConfig(p.oc, conf.Namespace, conf.Name)
	if err != nil {
		return nil, nil, nil, err
	}
	if !ok {
		_, bc, err = createIntoBuildConfig(p.oc, nil, conf)
		if err != nil {
			return nil, nil, nil, err
		}
	} else {
		bc = conf
	}
	buildRequestCauses := []buildapiv1.BuildTriggerCause{}
	obj := &buildapiv1.Build{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
			Kind:       "Build",
		},
		ObjectMeta: kapiv1.ObjectMeta{
			Name:      name,
			Namespace: bc.Namespace,
		},
		Spec: buildapiv1.BuildSpec{
			CommonSpec: bc.Spec.CommonSpec,
			TriggeredBy: append(buildRequestCauses,
				buildapiv1.BuildTriggerCause{
					Message: message,
				},
			),
		},
	}

	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[buildapi.BuildConfigAnnotation] = bc.Name //"openshift.io/build-config.name"
	obj.Annotations[buildapi.BuildNumberAnnotation] = "1"     //"openshift.io/build.number"
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[buildapi.BuildConfigAnnotation] = bc.Name
	obj.Labels[buildapi.BuildRunPolicyLabel] = "Serial" //"openshift.io/build.start-policy"
	obj.Status.Config = &kapiv1.ObjectReference{
		Kind:      bc.Kind,
		Name:      bc.Name,
		Namespace: bc.Namespace,
	}

	var raw []byte
	var result *buildapiv1.Build
	raw, result, err = createIntoBuild(p.oc, nil, obj)
	if err != nil {
		return nil, nil, nil, err
	}
	return raw, result, bc, nil
}

func (p *PaaS) VerifyProject(project string) error {
	if p.oc == nil {
		p.occtl()
	}
	if p.err != nil {
		glog.Errorf("Could not config openshift origin: %+v\n", p.err)
		return p.err
	}

	ok, err := findProject(p.oc, project)
	if err != nil {
		return err
	}
	if !ok {
		tgt := &projectapiv1.Project{
			TypeMeta: unversioned.TypeMeta{
				Kind:       "Project",
				APIVersion: projectapiv1.SchemeGroupVersion.Version,
			},
			ObjectMeta: kapiv1.ObjectMeta{
				Name: project,
			},
			Spec: projectapiv1.ProjectSpec{
				Finalizers: []kapiv1.FinalizerName{projectapiv1.FinalizerOrigin,
					kapiv1.FinalizerKubernetes},
			},
		}
		_, _, err = createIntoProject(p.oc, nil, tgt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PaaS) CreateNewBuild(obj *buildapiv1.Build, conf *buildapiv1.BuildConfig) ([]byte, *buildapiv1.Build, *buildapiv1.BuildConfig, error) {
	if p.oc == nil {
		p.occtl()
	}
	if p.err != nil {
		glog.Errorf("Could not config openshift origin: %+v\n", p.err)
		return nil, nil, nil, p.err
	}

	var ok bool
	var err error
	ok, err = findBuildConfig(p.oc, conf.Namespace, conf.Name)
	if err != nil {
		return nil, nil, nil, err
	}
	var bc *buildapiv1.BuildConfig
	if !ok {
		_, bc, err = createIntoBuildConfig(p.oc, nil, conf)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[buildapi.BuildConfigAnnotation] = conf.Name //"openshift.io/build-config.name"
	obj.Annotations[buildapi.BuildNumberAnnotation] = "1"       //"openshift.io/build.number"
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[buildapi.BuildConfigAnnotation] = conf.Name
	obj.Labels[buildapi.BuildRunPolicyLabel] = "Serial" //"openshift.io/build.start-policy"
	obj.Status.Config = &kapiv1.ObjectReference{
		Kind:      conf.Kind,
		Name:      conf.Name,
		Namespace: conf.Namespace,
	}
	var raw []byte
	var result *buildapiv1.Build
	raw, result, err = createIntoBuild(p.oc, nil, obj)
	if err != nil {
		return nil, nil, nil, err
	}
	return raw, result, bc, nil
}

const (
	Openshift_origin_api_error_formatter = "Could not access openshift: %+v"
	Helm_classic_setup_error_formatter   = "Faild to setup helm decode: %+v"
	Helm_classic_decode_error_formatter  = "Could not decode metadata: %+v; JSON: %+v"
	Kubernetes_deserialize_err_formatter = "Coude not deserialize into runtime object: %+v; JSON: %+v"
)

func validateRuntimeJSON(raw []byte, kind string) error {
	if len(raw) == 0 {
		glog.Warningln("Nothing should be to deserialize")
		return nil
	}
	js := string(raw)
	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf(Helm_classic_setup_error_formatter, err)
		return fmt.Errorf("%+v; %+v", err, js)
	}

	meta := new(unversioned.TypeMeta)
	if err := hco.Object(meta); err != nil {
		glog.Errorf(Helm_classic_decode_error_formatter, err, js)
		return fmt.Errorf("%+v; %+v", err, js)
	}

	switch {
	case strings.EqualFold(kind, meta.Kind):
		return nil
	case strings.EqualFold("Status", meta.Kind):
		status := new(unversioned.Status)
		if err := hco.Object(status); err != nil {
			glog.Warningf("Failed to decode into status: %+v", err)
		} else {
			glog.Warningf("Status message: %+v", status.Message)
		}
		err = fmt.Errorf("%+v; %+v", errBadRequest, js)
	default:
		glog.Errorf("Unexpected data: %+v", js)
		err = fmt.Errorf("%+v; %+v", errUnexpected, js)
	}
	return err
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

	clientnats := yagnats.NewClient()
	if err := clientnats.Connect(&yagnats.ConnectionInfo{
		Addr:     gnatsdAddresses[0],
		Username: gnatsdUsername,
		Password: gnatsdPassword,
	}); err != nil {
		return nil, nil, err
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

			case <-time.After(time.Duration(1800) * time.Second):
				m.Lock()
				defer m.Unlock()
				timeout = true
			}
		}()
		subject := obj.Namespace + "/" + obj.Name
		for false == timeout {
			time.Sleep(1000 * time.Millisecond)
			l := b.Len()
			if l > offset {
				//fmt.Print(string(b.Next(l - offset)))
				clientnats.Publish(subject, b.Next(l-offset))
				offset = l
			}
		}
	}()

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
	return createUser(nil, user)
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

	osClient := &oclient.Client{clientK8s}

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
		f := oclientcmd.New(pf)
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
		f := oclientcmd.New(pf)
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

func ProjectCreation(name string, in io.Reader, out, errOut io.Writer) {
	logger = log.New(os.Stdout, "[appliance/openshift/origin, ProjectWithSimple] ", log.LstdFlags|log.Lshortfile)

	f := oclientcmd.NewFactory(withOClientConfig())
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
		val.Value.Set(osoconfig_path)
	} else {
		val = c.Flags().VarPF(stringValue{osoconfig_path}, "config", "", "")
		fmt.Fprintf(os.Stdout, "[tangfx] config: %+v\n", val)
	}
	c.Flags().Lookup("config").NoOptDefVal = osoconfig_path
}

func overrideRootCommandArgs() {
	logger.SetOutput(os.Stdout)
	logger.SetPrefix("[appliance/openshift/origin, overrideRootCommandArgs] ")

	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	//f := oclientcmd.New(flags)
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
		val.Value.Set(osoconfig_path)
	} else {
		fmt.Println("setting")
		val = rootCommand.PersistentFlags().VarPF(stringValue{osoconfig_path}, "config", "", "")
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
