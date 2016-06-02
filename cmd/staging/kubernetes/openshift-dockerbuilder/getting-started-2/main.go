package main

import (
	"encoding/json"
	_ "errors"
	"flag"
	"io/ioutil"
	"log"
	_ "net/url"
	"os"
	"os/signal"
	_ "strings"
	"syscall"

	"github.com/golang/glog"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildclient "github.com/openshift/origin/pkg/build/client"
	// buildcontroller "github.com/openshift/origin/pkg/build/controller"
	buildcontrollerfactory "github.com/openshift/origin/pkg/build/controller/factory"
	buildstrategy "github.com/openshift/origin/pkg/build/controller/strategy"
	"github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	_ "github.com/openshift/origin/pkg/cmd/server/admin"
	configapi "github.com/openshift/origin/pkg/cmd/server/api"
	"github.com/openshift/origin/pkg/cmd/server/origin"
	"github.com/openshift/origin/pkg/cmd/server/start"
	"github.com/openshift/origin/pkg/cmd/util/variable"
	"github.com/openshift/origin/pkg/controller"

	"k8s.io/kubernetes/pkg/admission"
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/cache"
	_ "k8s.io/kubernetes/pkg/client/record"
	// kclient "k8s.io/kubernetes/pkg/client/unversioned"
	clientadapter "k8s.io/kubernetes/pkg/client/unversioned/adapters/internalclientset"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util"
	_ "k8s.io/kubernetes/pkg/util/sets"
)

var (
	logger *log.Logger = log.New(os.Stdout, "[tangfx] ", log.LstdFlags|log.Lshortfile)

	kubeconfig    string = "/data/src/github.com/openshift/origin/openshift.local.config/master/kubeconfig"
	contextname   string = "openshift-origin-single"
	kClientConfig clientcmd.ClientConfig

	buildPath string = "/data/src/github.com/tangfeixiong/go-to-cloud-1/examples/github101.json"

	masterAddr         string = "https://172.17.4.50:30448"
	etcdAddr           string = "http://10.3.0.16:4001"
	configDir          string = "/data/src/github.com/openshift/origin/openshift.local.config/master"
	kubernetesAddr     string = "https://172.17.4.50"
	clusterNetworkCIDR string = "172.17.0.0/22"
	hostSubnetLength   uint   = 7
	serviceNetworkCIDR string = "10.3.0.0/24"

	dockerImage string = "172.17.4.200:5000/docker.io_openshift_origin-docker-builder:v1.3.0-alpha.0"
)

func withKClientConfig() {
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

	kClientConfig := clientcmd.NewNonInteractiveClientConfig(*conf, contextname, &clientcmd.ConfigOverrides{})
	logger.Printf("rest client config: %+v\n", kClientConfig)
}

func main() {
	flag.Parse()
	flag.Lookup("v").Value.Set("10")
	withKClientConfig()

	buildData, err := ioutil.ReadFile(buildPath)
	if err != nil {
		logger.Printf("build data not ready: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("build data: \n%+v\n", string(buildData))

	buildObj := new(buildapi.Build)
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), buildData, buildObj); err != nil {
		logger.Printf("unable to create build object: %v\n", err)
		os.Exit(1)
	}
	if buildObj.Spec.Source.Dockerfile == nil {
		if err := json.Unmarshal(buildData, buildObj); err != nil {
			logger.Printf("could not decode into build object: %v\n", err)
			os.Exit(1)
		}
	}
	buildObj.Status.Phase = buildapi.BuildPhaseNew

	options, err := buildSerializeableMasterConfig()
	if err != nil {
		logger.Printf("could not create master config: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("master config created: %v\n", options)
	/*
	   /Users/fanhongling/Downloads/github.com/openshift/origin/pkg/cmd/server/origin/master_config.go#BuildMasterConfig
	*/
	c, err := origin.BuildMasterConfig(*options)
	if err != nil {
		logger.Printf("could not know origin config: %v\n", err)
		os.Exit(1)
	}
	logger.Printf("origin config created: %v\n", c)
	factory := createBuildControllerFactory(c)
	logger.Printf("build factory created: %v\n", c)
	/*
	   /Users/fanhongling/Downloads/github.com/openshift/origin/pkg/build/controller/factory/factory.go#Create
	*/
	/*
		eventBroadcaster := record.NewBroadcaster()
		eventBroadcaster.StartRecordingToSink(factory.KubeClient.Events(""))

		client := buildcontrollerfactory.ControllerClient{factory.KubeClient, factory.OSClient}
		buildController := &buildcontroller.BuildController{
			BuildUpdater:      factory.BuildUpdater,
			ImageStreamClient: client,
			PodManager:        client,
			BuildStrategy: &typeBasedFactoryStrategy{
				DockerBuildStrategy: factory.DockerBuildStrategy,
				SourceBuildStrategy: factory.SourceBuildStrategy,
				CustomBuildStrategy: factory.CustomBuildStrategy,
			},
			Recorder: eventBroadcaster.NewRecorder(kapi.EventSource{Component: "build-controller"}),
		}

		fakeRetryController := struct{ Handle func(obj interface{}) error }{
			Handle: func(obj interface{}) error {
				build := obj.(*buildapi.Build)
				err := buildController.HandleBuild(build)
				if err != nil {
					// Update the build status message only if it changed.
					if msg := errors.ErrorToSentence(err); build.Status.Message != msg {
						// Set default Reason.
						if len(build.Status.Reason) == 0 {
							build.Status.Reason = buildapi.StatusReasonError
						}
						build.Status.Message = msg
						if err := buildController.BuildUpdater.Update(build.Namespace, build); err != nil {
							glog.V(2).Infof("Failed to update status message of Build %s/%s: %v", build.Namespace, build.Name, err)
						}
						buildController.Recorder.Eventf(build, kapi.EventTypeWarning, "HandleBuildError", "Build has error: %v", err)
					}
				}
				return err
			},
		}
	*/

	glog.V(10).Infoln("build")
	fakeRetryController := factory.Create().(*controller.RetryController)
	queue := fakeRetryController.Queue.(*cache.FIFO)
	queue.AddIfNotPresent(buildObj)
	if err := fakeRetryController.Handle(buildObj); err != nil {
		logger.Printf("could not build docker image: %v\n", err)
		//os.Exit(1)
		var fatal buildstrategy.FatalError = "Could not build docker image"
		fakeRetryController.RetryManager.Retry(buildObj, fatal)

		fakeRetryController.Run()
	}

	// Catch the Ctrl-C and SIGTERM from kill command

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		signalType := <-ch
		signal.Stop(ch)
		glog.V(10).Infoln("Exit command received. Exiting...")

		// this is a good place to flush everything to disk
		// before terminating.
		glog.V(10).Infof("Signal type : %v", signalType)

		os.Exit(0)

	}()
}

/*
   /Users/fanhongling/Downloads/github.com/openshift/origin/pkg/cmd/server/origin/run_components.go#RunBuildController
*/
func createBuildControllerFactory(c *origin.MasterConfig) *buildcontrollerfactory.BuildControllerFactory {
	// initialize build controller
	dockerImage := c.ImageFor("docker-builder")
	stiImage := c.ImageFor("sti-builder")

	storageVersion := c.Options.EtcdStorageConfig.OpenShiftStorageVersion
	groupVersion := unversioned.GroupVersion{Group: "", Version: storageVersion}
	codec := kapi.Codecs.LegacyCodec(groupVersion)

	admissionControl := admission.NewFromPlugins(clientadapter.FromUnversionedClient(c.PrivilegedLoopbackKubernetesClient), []string{"SecurityContextConstraint"}, "")

	osclient, kclient := c.BuildControllerClients()
	//_, osclient, kclient, err := c.GetServiceAccountClients("default")
	//if err != nil {
	//	logger.Printf("could not get service account: %v\n", err)
	//	os.Exit(1)
	//}
	factory := buildcontrollerfactory.BuildControllerFactory{
		OSClient:     osclient,
		KubeClient:   kclient,
		BuildUpdater: buildclient.NewOSClientBuildClient(osclient),
		DockerBuildStrategy: &buildstrategy.DockerBuildStrategy{
			Image: dockerImage,
			// TODO: this will be set to --storage-version (the internal schema we use)
			Codec: codec,
		},
		SourceBuildStrategy: &buildstrategy.SourceBuildStrategy{
			Image: stiImage,
			// TODO: this will be set to --storage-version (the internal schema we use)
			Codec:            codec,
			AdmissionControl: admissionControl,
		},
		CustomBuildStrategy: &buildstrategy.CustomBuildStrategy{
			// TODO: this will be set to --storage-version (the internal schema we use)
			Codec: codec,
		},
	}
	return &factory
}

/*
  /Users/fanhongling/Downloads/github.com/openshift/origin/pkg/cmd/server/start/master_args.go#BuildSerializeableMasterConfig
*/
func buildSerializeableMasterConfig() (*configapi.MasterConfig, error) {
	args := &start.MasterArgs{
		MasterAddr:       flagtypes.Addr{Value: masterAddr, DefaultScheme: "https", DefaultPort: 30448, AllowPrefix: true}.Default(),
		EtcdAddr:         flagtypes.Addr{Value: etcdAddr, DefaultScheme: "http", DefaultPort: 4001}.Default(),
		MasterPublicAddr: flagtypes.Addr{Value: masterAddr, DefaultScheme: "https", DefaultPort: 30448, AllowPrefix: true}.Default(),
		DNSBindAddr:      flagtypes.Addr{Value: "0.0.0.0:53", DefaultScheme: "tcp", DefaultPort: 53, AllowPrefix: true}.Default(),

		ConfigDir: &util.StringFlag{},

		ListenArg: &start.ListenArg{
			ListenAddr: flagtypes.Addr{Value: "0.0.0.0:8443", DefaultScheme: "https", DefaultPort: 8443, AllowPrefix: true}.Default(),
		},
		ImageFormatArgs: &start.ImageFormatArgs{
			ImageTemplate: variable.NewDefaultImageTemplate(),
		},
		KubeConnectionArgs: &start.KubeConnectionArgs{
			KubernetesAddr: flagtypes.Addr{Value: kubernetesAddr, DefaultScheme: "https", DefaultPort: 443, AllowPrefix: true}.Default(),
			ClientConfig:   kClientConfig,
			ClientConfigLoadingRules: clientcmd.ClientConfigLoadingRules{
				ExplicitPath: kubeconfig,
			},
		},
		NetworkArgs: &start.NetworkArgs{
			NetworkPluginName:  "",
			ClusterNetworkCIDR: clusterNetworkCIDR,
			HostSubnetLength:   hostSubnetLength,
			ServiceNetworkCIDR: serviceNetworkCIDR,
		},
	}
	args.MasterAddr.Provided = true
	args.EtcdAddr.Provided = true
	args.MasterPublicAddr.Provided = true
	args.DNSBindAddr.Provided = true
	args.ConfigDir.Set(configDir)
	args.ListenArg.ListenAddr.Provided = true
	args.KubeConnectionArgs.KubernetesAddr.Provided = true

	return args.BuildSerializeableMasterConfig()
}
