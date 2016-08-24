package origin

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

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
			//Labels:            map[string]string{buildapi.BuildConfigLabel: "tangfx"},
			//Annotations:       map[string]string{buildapi.BuildNumberAnnotation: "1"},
		},
	}
	obj.Spec = buildapi.BuildSpec{
		TriggeredBy: []buildapi.BuildTriggerCause{
			{
				Message: "No message",
				GenericWebHook: &buildapi.GenericWebHookCause{
					Revision: &buildapi.SourceRevision{
						Git: &buildapi.GitSourceRevision{
							Commit: "master",
							Author: buildapi.SourceControlUser{
								Name:  "tangfeixiong",
								Email: "tangfx128@gmail.com",
							},
							Committer: buildapi.SourceControlUser{
								Name:  "tangfeixiong",
								Email: "tangfx128@gmail.com",
							},
							Message: "example",
						},
					},
					Secret: "",
				},
			},
		},
	}
	obj.Spec.CommonSpec = buildapi.CommonSpec{
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
	}
	obj.Status = buildapi.BuildStatus{
		Phase: buildapi.BuildPhaseNew,
	}

	return CreateBuildWith(obj)
}

// gitRef: branch name, tag name, or commit revision
func CreateDockerBuildV1Example(name, projectName string, gitSecret map[string]string, gitURI, gitRef, contextDir string, sourceImages []map[string]interface{}, dockerfile string, buildSecrets []map[string]interface{}, buildStrategy map[string]interface{}) ([]byte, *buildapiv1.Build, error) {
	obj := &buildapiv1.Build{
		TypeMeta: unversioned.TypeMeta{
			Kind:       "Build",
			APIVersion: buildapiv1.SchemeGroupVersion.Version,
		},
		ObjectMeta: kapiv1.ObjectMeta{
			Name:              name,
			Namespace:         projectName,
			CreationTimestamp: unversioned.Now(),
			//Labels:            map[string]string{buildapiv1.BuildConfigLabel: "tangfx"},
			//Annotations:       map[string]string{buildapiv1.BuildNumberAnnotation: "1"},
		},
	}
	obj.Spec = buildapiv1.BuildSpec{
	/*TriggeredBy: []buildapiv1.BuildTriggerCause{
		{
			Message: "No message",
			GenericWebHook: &buildapiv1.GenericWebHookCause{
				Revision: &buildapiv1.SourceRevision{
					Git: &buildapiv1.GitSourceRevision{
						Commit: "master",
						Author: buildapiv1.SourceControlUser{
							Name:  "tangfeixiong",
							Email: "tangfx128@gmail.com",
						},
						Committer: buildapiv1.SourceControlUser{
							Name:  "tangfeixiong",
							Email: "tangfx128@gmail.com",
						},
						Message: "example",
					},
				},
				Secret: "",
			},
		},
	},*/
	}
	obj.Spec.CommonSpec = buildapiv1.CommonSpec{
		ServiceAccount: builderServiceAccount,
		Source: buildapiv1.BuildSource{
			//Binary : &buildapiv1.BinaryBuildSource {},
			//Dockerfile: &dockerfile,
			Git: &buildapiv1.GitBuildSource{
				URI: gitURI,
				Ref: gitRef,
				//HTTPProxy: nil,
				//HTTPSProxy: nil,
			},
			/*Images : []buildapiv1.ImageSource {
			    buildapiv1.ImageSource {
			        From : kapiv1.ObjectReference {
			            Kind : "DockerImage",
			            Name : "alpine:edge",
			        },
			        Paths : []buildapiv1.ImageSourcePath {
			           {
			               SourcePath : "",
			               DestinationDir : "",
			           },
			        },
			        PullSecret : &kapiv1.LocalObjectReference {
			        },
			   },
			},*/
			ContextDir: contextDir,
			//SourceSecret : &kapiv1.LocalObjectReference {
			//    name : githubSecret,
			//},
			//Secrets : []buildapiv1.SecretBuildSource {
			//    Secret : &kapiv1.LocalObjectReference {},
			//    DestinationDir : "/root/.docker/config.json",
			//},
			Type: buildapiv1.BuildSourceGit, // new
		},
		//Revision: &buildapiv1.SourceRevision {},
		Strategy: buildapiv1.BuildStrategy{
			Type: buildapiv1.DockerBuildStrategyType, // new
			DockerStrategy: &buildapiv1.DockerBuildStrategy{
				From: &kapiv1.ObjectReference{
					Kind: "DockerImage",
					Name: "alpine:edge",
				},
				//PullSecret: &kapiv1.LocalObjectReference{
				//	Name: dockerPullSecret,
				//},
				NoCache: false,
				//Env : []kapiv1.EnvVar {},
				ForcePull: false,
				//DockerfilePath : ".",
			},
		},
		Output: buildapiv1.BuildOutput{
			To: &kapiv1.ObjectReference{
				Kind: "DockerImage",
				Name: "docker.io/tangfeixiong/nc-http-dev:latest",
			},
			PushSecret: &kapiv1.LocalObjectReference{
				Name: dockerPushSecret,
			},
		},
		//Resources : kapiv1.ResourceRequirements {},
		//PostCommit : buildapiv1.BuildPostCommitSpec {
		//    Command : []string{},
		//    Args : []string{},
		//    Script: "",
		//},
		//CompletionDeadlineSeconds: &timeout,
	}
	obj.Status = buildapiv1.BuildStatus{
		Phase: buildapiv1.BuildPhaseNew,
	}

	return CreateBuildWithV1(obj)
}

func CreateBuildWith(obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	return createBuild(nil, obj)
}

func CreateBuildWithV1(obj *buildapiv1.Build) ([]byte, *buildapiv1.Build, error) {
	logger = log.New(os.Stdout, "[appliance/openshift/origin, CreateBuildWithV1] ", log.LstdFlags|log.Lshortfile)

	b := new(bytes.Buffer)
	if err := codec.JSON.Encode(b).One(obj); err != nil {
		logger.Printf("Could not encode into openshift origin: %s", err)
		return nil, nil, err
	}

	raw, _, err := createBuild(b.Bytes(), nil)
	if err != nil {
		return nil, nil, err
	}
	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		logger.Printf("Could not setup runtime object decoder: %s", err)
		return nil, nil, err
	}
	result := new(buildapiv1.Build)
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into runtime object: %s", err)
		return nil, nil, err
	}
	return raw, result, err
}

func CreateBuildFromArbitray(data []byte) ([]byte, *buildapi.Build, error) {
	return createBuild(data, nil)
}

func createBuild(data []byte, obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	logger = log.New(os.Stdout, "[appliance/openshift/origin, createBuild] ", log.LstdFlags|log.Lshortfile)

	if len(data) == 0 && obj == nil || obj != nil && len(obj.Namespace) == 0 {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %+v", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	if len(data) == 0 {
		/*result, err := oc.Builds(obj.Namespace).Create(obj)
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
		b := new(bytes.Buffer)
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
		data = b.Bytes()*/
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
		if data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
	}
	logger.Printf("Build: %s\n", string(data))
	if obj == nil {
		hco, err := codec.JSON.Decode(data).One()
		if err != nil {
			glog.Errorf("Could not setup openshift origin codec: %s", err)
			return nil, nil, err
		}
		obj = new(buildapi.Build)
		if err := hco.Object(obj); err != nil {
			glog.Errorf("Could not codec with openshift origin: %s", err)
			return nil, nil, err
		}
		if obj.Name == "" || obj.Namespace == "" {
			kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
			kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, &buildapiv1.Build{})
			val := new(buildapiv1.Build)
			if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(),
				data, val); err != nil {
				glog.Errorf("Could not serialize runtime object: %+v", err)
				return nil, nil, err
			}
			if val == nil {
				glog.V(6).Infoln("Nothing deserialized")
				return nil, nil, errUnexpected
			}
			obj.Name = val.Name
			obj.Namespace = val.Namespace
		}
	}
	logger.Printf("Build: %+v\n", obj)
	raw, err := oc.RESTClient.Post().Namespace(obj.Namespace).
		Resource("builds").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	if len(raw) == 0 {
		glog.V(6).Infoln("Nothing deserialized")
		return nil, nil, errUnexpected
	}

	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf("Could not create helm object: %s", err)
		return raw, nil, err
	}
	meta := new(unversioned.TypeMeta)
	//meta, err := hco.Meta()
	if err := hco.Object(meta); err != nil {
		glog.Errorf("Could not decode into metadata: %s\nReturn: %+v\n", err, string(raw))
		return raw, nil, err
	}
	if ok := strings.EqualFold("Build", meta.Kind); !ok {
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

func RetrieveBuild(namespace, name string) ([]byte, *buildapiv1.Build, error) {
	logger = log.New(os.Stdout, "[RetrieveBuild] ", log.LstdFlags|log.Lshortfile)

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

	/*result, err := oc.Builds(namespace).Get(name)
	if err != nil {
		if result == nil {
			glog.Errorf("Could not get build %s: %+v", name, err)
			return nil, nil, err
		}
		logger.Printf("Result:\n%+v\n", result)
	}
	if result == nil {
		glog.V(7).Infoln("Unexpected retrieve: %s", name)
		return nil, nil, errUnexpected
	}
	if strings.EqualFold("Build", result.Kind) && len(result.Name) > 0 {
		//b := new(bytes.Buffer)
		//if err := codec.JSON.Encode(b).One(result); err != nil {
		//	glog.Errorf("Could not encode runtime object: %s", err)
		//	return nil, result, err
		//}
		//logger.Printf("Build: %+v\n", b.String())
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
		data, err := runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			result)
		if err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, result, err
		}
		return data, result, nil
	}*/

	raw, err := oc.RESTClient.Get().Resource("builds").Namespace(namespace).Name(name).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	if len(raw) == 0 {
		return nil, nil, nil
	}
	logger.Printf("Result:\n%s\n", string(raw))
	if bytes.IndexAny(raw, "404:") == 0 {
		return nil, nil, nil
	}
	//kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
	//obj, err := runtime.Decode(kapi.Codecs.UniversalDeserializer(), raw)
	//if err != nil {
	//	glog.Errorf("Could not deserialize raw: %+v", err)
	//	return raw, nil, err
	//}
	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf("Could not create helm object: %s", err)
		return nil, nil, err
	}
	//meta, err := hco.Meta()
	meta := unversioned.TypeMeta{}
	if err := hco.Object(&meta); err != nil {
		glog.Errorf("Could not decode into metadata: %s", err)
		return nil, nil, err
	}
	logger.Printf("Meta: %+v", meta)
	if ok := strings.EqualFold("Build", meta.Kind); !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not know metadata: %+v", meta)
				return nil, nil, err
			}
			glog.Warningf("Status message: %+v", status.Message)
			return nil, nil, nil
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return nil, nil, errUnexpected
	}
	logger.Printf("Helm Object: %+v", hco)
	out := new(buildapiv1.Build)
	if err := hco.Object(out); err != nil {
		glog.Errorf("Could not decode raw data: %s", err)
		return nil, nil, err
	}
	return raw, out, nil
}

func DeleteBuild(namespace, name string) error {
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

	if err := oc.Builds(namespace).Delete(name); err != nil {
		glog.Errorf("Could not delete build config %s: %+v", name, err)
		return err
	}
	return nil
}
