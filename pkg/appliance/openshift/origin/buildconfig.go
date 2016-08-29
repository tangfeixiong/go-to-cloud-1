package origin

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	oclient "github.com/openshift/origin/pkg/client"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
)

func createIntoBuildConfig(oc *oclient.Client,
	data []byte, obj *buildapiv1.BuildConfig) ([]byte, *buildapiv1.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, createIntoBuildConfig] ")
	if oc == nil {
		f := util.NewClientCmdFactory()
		var err error
		oc, _, err = f.Clients()
		if err != nil {
			logger.Printf("Could not create openshift origin client: %+v", err)
			return nil, nil, err
		}
	}
	if len(data) == 0 {
		b := &bytes.Buffer{}
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			logger.Printf("Could not serialize: %s\n", err)
			return nil, nil, err
		}
		data = b.Bytes()
	}
	raw, err := oc.RESTClient.Verb("POST").Namespace(obj.Namespace).Resource("buildConfigs").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	if len(raw) == 0 {
		logger.Println("Nothing deserialized")
		return nil, nil, errUnexpected
	}

	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf("Could not create helm object: %s", err)
		return raw, nil, err
	}
	meta := new(unversioned.TypeMeta)
	if err := hco.Object(meta); err != nil {
		glog.Errorf("Could not decode into typemeta: %s\nReturn: %+v\n", err, string(raw))
		return raw, nil, err
	}

	if !strings.EqualFold("BuildConfig", meta.Kind) {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Warningf("Could not inspect metadata: %+v", meta)
				return raw, nil, err
			}
			glog.Warningf("Status inspected: %+v", status.Message)
			return raw, nil, nil
		}
		glog.Errorf("Unexpected result: %+v", string(raw))
		return raw, nil, errUnexpected
	}

	//meta, err := hco.Meta()
	result := new(buildapiv1.BuildConfig)
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	glog.V(10).Infof("BuildConfig result: %+v\n", string(raw))
	return raw, result, nil
}

func retrieveIntoBuildConfig(oc *oclient.Client, project, name string) ([]byte, *buildapiv1.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, retrieveIntoBuildConfig] ")

	if oc == nil {
		f := util.NewClientCmdFactory()
		var err error
		oc, _, err = f.Clients()
		if err != nil {
			logger.Printf("Could not create openshift client: %+v", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Verb("GET").Namespace(project).Resource("buildConfigs").Name(name).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	if len(raw) == 0 {
		logger.Println("Nothing deserialized")
		return nil, nil, errUnexpected
	}

	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		glog.Errorf("Could not create helm object: %s", err)
		return raw, nil, err
	}
	meta := new(unversioned.TypeMeta)
	if err := hco.Object(meta); err != nil {
		glog.Errorf("Could not decode into typemeta: %s\nReturn: %+v\n", err, string(raw))
		return raw, nil, err
	}

	if !strings.EqualFold("BuildConfig", meta.Kind) {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Warningf("Could not inspect unknown metadata: %+v", meta)
				return raw, nil, nil
			}
			glog.Warningf("Status inspected: %+v", status.Message)
			return nil, nil, nil
		}
		glog.Errorf("Unexpected result: %+v", string(raw))
		return raw, nil, errUnexpected
	}

	//meta, err := hco.Meta()
	result := new(buildapiv1.BuildConfig)
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	glog.V(10).Infof("result: %+v\n", string(raw))
	return raw, result, nil
}

func findBuildConfig(oc *oclient.Client, project, name string) (bool, error) {
	_, obj, err := retrieveIntoBuildConfig(oc, project, name)
	if err != nil {
		return false, err
	}
	return obj != nil, nil
}

func CreateIntoBuildConfigWithV1(obj *buildapiv1.BuildConfig) ([]byte, *buildapiv1.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, CreateIntoBuildConfigWithV1] ")

	f := util.NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		logger.Printf("Could not create openshift origin client: %+v", err)
		return nil, nil, err
	}
	glog.V(10).Infof("Create openshift origin client: %+v", oc)

	b := &bytes.Buffer{}
	if err := codec.JSON.Encode(b).One(obj); err != nil {
		logger.Printf("Could not serialize: %+v", err)
		return nil, nil, err
	}
	data := b.Bytes()

	return createIntoBuildConfig(oc, data, obj)
}

func RetrieveBuildConfig(project, name string) ([]byte, *buildapi.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, RetrieveBuildConfig] ")

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

	result, err := oc.BuildConfigs(project).Get(name)
	if err != nil {
		if result == nil {
			glog.Errorf("Could not delete build config %s: %+v", name, err)
			return nil, nil, err
		}
		logger.Printf("BuildConfig:\n%+v\n", result)
	}
	if result == nil {
		glog.V(7).Infoln("Unexpected retrieve: %s", name)
		return nil, nil, errUnexpected
	}
	if strings.EqualFold("BuildConfig", result.Kind) && len(result.Name) > 0 {
		//b := new(bytes.Buffer)
		//if err := codec.JSON.Encode(b).One(result); err != nil {
		//	glog.Errorf("Could not encode runtime object: %s", err)
		//	return nil, result, err
		//}
		//logger.Printf("Build Config: %+v\n", b.String())
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
		data, err := runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			result)
		if err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, result, err
		}
		return data, result, nil
	}

	raw, err := oc.RESTClient.Get().Namespace(project).Resource("buildConfigs").Name(name).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift: %s", err)
		return nil, nil, err
	}
	logger.Printf("raw:\n%s\n", string(raw))
	if bytes.IndexAny(raw, "404:") == 0 {
		return nil, nil, nil
	}
	//kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
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
	//meta, err := hco.Meta()
	meta := unversioned.TypeMeta{}
	if err := hco.Object(&meta); err != nil {
		glog.Errorf("Could not decode into metadata: %s", err)
		return raw, nil, err
	}
	if ok := strings.EqualFold("BuildConfig", meta.Kind); !ok {
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
	result = new(buildapi.BuildConfig)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode raw data: %s", err)
		return raw, nil, err
	}
	logger.Printf("Return runtime object: %s\n", string(raw))
	return raw, result, nil
}

func RetrieveBuildConfigs(project string) error {
	logger.SetPrefix("[appliance/openshift/origin, RetrieveBuildConfigs] ")
	f := util.NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		logger.Printf("Could not create openshift client: %+v\n", err)
		return err
	}

	result, err := oc.BuildConfigs(project).List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func convertFromV1(data []byte, obj *buildapiv1.BuildConfig) ([]byte, *buildapi.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, convertFromV1] ")

	var hco *codec.Object
	var err error
	if len(data) > 0 && obj == nil {
		if hco, err = codec.JSON.Decode(data).One(); err != nil {
			logger.Printf("Could not setup decoder (BuildConfig): %+v", err)
			return nil, nil, err
		}
		obj = new(buildapiv1.BuildConfig)
		if err := hco.Object(obj); err != nil {
			logger.Printf("Could not decode into build config: %+v", err)
			return nil, nil, err
		}
	}
	if obj == nil {
		return nil, nil, errBadRequest
	}
	tgt := new(buildapi.BuildConfig)

	b := &bytes.Buffer{}
	if err = codec.JSON.Encode(b).One(&obj.TypeMeta); err != nil {
		logger.Printf("Could not serialize build config (TypeMeta): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (TypeMeta): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.TypeMeta); err != nil {
		logger.Printf("Could not decode into TypeMeta: %+v", err)
		return nil, nil, err
	}
	if !strings.EqualFold(tgt.Kind, "Build") || !strings.EqualFold(tgt.APIVersion, "v1") {
		glog.Errorf("Invalid destination type from meta: %s, %s", tgt.Kind, tgt.APIVersion)
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.ObjectMeta); err != nil {
		logger.Printf("Could not serialize build config (ObjectMeta): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (ObjectMeta): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.ObjectMeta); err != nil {
		logger.Printf("Could not decode into ObjectMeta: %+v", err)
		return nil, nil, err
	}
	if tgt.Namespace == "" || tgt.Name == "" {
		glog.Errorln("Invalid destination object from meta")
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Spec); err != nil {
		logger.Printf("Could not serialize build config (BuildSpec): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (BuildSpec): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.Spec); err != nil {
		logger.Printf("Could not decode into BuildSpec: %+v", err)
		return nil, nil, err
	}
	if tgt.Spec.Source.Dockerfile == nil && tgt.Spec.Source.Git == nil &&
		len(tgt.Spec.Source.Images) == 0 && tgt.Spec.Source.Binary == nil {
		glog.Errorln("Invalid destination from BuildSpec")
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Status); err != nil {
		logger.Printf("Could not serialize build config (BuildStatus): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (BuildStatus): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.Status); err != nil {
		logger.Printf("Could not decode into BuildStatus: %+v", err)
		return nil, nil, err
	}

	b.Reset()
	if err = codec.JSON.Encode(b).One(tgt); err != nil {
		logger.Printf("Could not encode into bytes: %+v", err)
		return nil, nil, err
	}
	glog.V(10).Infof("Destination object: \n%s", b.String())
	return b.Bytes(), tgt, nil
}

func convertIntoV1(data []byte, obj *buildapi.BuildConfig) ([]byte, *buildapiv1.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, convertIntoV1] ")

	var hco *codec.Object
	var err error
	if len(data) > 0 && obj == nil {
		if hco, err = codec.JSON.Decode(data).One(); err != nil {
			logger.Printf("Could not setup decoder (BuildConfig): %+v", err)
			return nil, nil, err
		}
		obj = new(buildapi.BuildConfig)
		if err := hco.Object(obj); err != nil {
			logger.Printf("Could not decode into build config: %+v", err)
			return nil, nil, err
		}
	}
	if obj == nil {
		return nil, nil, errBadRequest
	}
	tgt := new(buildapiv1.BuildConfig)

	b := &bytes.Buffer{}
	if err = codec.JSON.Encode(b).One(&obj.TypeMeta); err != nil {
		logger.Printf("Could not serialize build config (TypeMeta): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (TypeMeta): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.TypeMeta); err != nil {
		logger.Printf("Could not decode into TypeMeta: %+v", err)
		return nil, nil, err
	}
	if !strings.EqualFold(tgt.Kind, "Build") || !strings.EqualFold(tgt.APIVersion, "v1") {
		glog.Errorf("Invalid destination type from meta: %s, %s", tgt.Kind, tgt.APIVersion)
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.ObjectMeta); err != nil {
		logger.Printf("Could not serialize build config (ObjectMeta): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (ObjectMeta): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.ObjectMeta); err != nil {
		logger.Printf("Could not decode into ObjectMeta: %+v", err)
		return nil, nil, err
	}
	if tgt.Namespace == "" || tgt.Name == "" {
		glog.Errorln("Invalid destination object from meta")
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Spec); err != nil {
		logger.Printf("Could not serialize build config (BuildSpec): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (BuildSpec): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.Spec); err != nil {
		logger.Printf("Could not decode into BuildSpec: %+v", err)
		return nil, nil, err
	}
	if tgt.Spec.Source.Dockerfile == nil && tgt.Spec.Source.Git == nil &&
		len(tgt.Spec.Source.Images) == 0 && tgt.Spec.Source.Binary == nil {
		glog.Errorln("Invalid destination from BuildSpec")
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Status); err != nil {
		logger.Printf("Could not serialize build config (BuildStatus): %+v", err)
		return nil, nil, err
	}
	if hco, err = codec.JSON.Decode(b.Bytes()).One(); err != nil {
		logger.Printf("Could not setup decoder (BuildStatus): %+v", err)
		return nil, nil, err
	}
	if err = hco.Object(&tgt.Status); err != nil {
		logger.Printf("Could not decode into BuildStatus: %+v", err)
		return nil, nil, err
	}

	b.Reset()
	if err = codec.JSON.Encode(b).One(tgt); err != nil {
		logger.Printf("Could not encode into bytes: %+v", err)
		return nil, nil, err
	}
	glog.V(10).Infof("Destination object: \n%s", b.String())
	return b.Bytes(), tgt, nil
}

func CreateBuildConfigWith(obj *buildapi.BuildConfig) ([]byte, *buildapi.BuildConfig, error) {
	return createBuildConfig(nil, obj)
}

func createBuildConfig(data []byte, obj *buildapi.BuildConfig) ([]byte, *buildapi.BuildConfig, error) {
	logger.SetPrefix("[appliance/openshift/origin, createBuildConifg] ")

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
		result, err := oc.BuildConfigs(obj.Namespace).Create(obj)
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
			if strings.EqualFold("BuildConfig", result.Kind) && len(result.Name) > 0 {
				b := new(bytes.Buffer)
				if err := codec.JSON.Encode(b).One(result); err != nil {
					glog.Errorf("Could not encode runtime object: %s", err)
					return nil, result, err
				}
				logger.Printf("BuildConfig: %+v\n", b.String())
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
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
		if data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			obj); err != nil {
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
		obj = new(buildapi.BuildConfig)
		if err := hco.Object(obj); err != nil {
			glog.Errorf("Could not deserialize into runtime object: %s", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Post().Namespace(obj.Namespace).Resource("buildConfigs").Body(data).DoRaw()
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
	if ok := strings.EqualFold("BuildConfig", meta.Kind) && len(meta.Name) > 0; !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not know metadata: %+v", meta)
				return raw, nil, err
			}
			glog.Warningf("Could not create buildconfig: %+v", status.Message)
			return raw, nil, fmt.Errorf("Could not create buildconfig: %+v", status.Message)
		}
		glog.Errorf("Could not know metadata: %+v", string(raw))
		return raw, nil, errUnexpected
	}
	result := new(buildapi.BuildConfig)
	if err := hco.Object(result); err != nil {
		glog.Errorf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	logger.Printf("BuildConfig: %+v\n", string(raw))
	return raw, result, nil
}

func DeleteBuildConfig(namespace, name string) error {
	logger.SetPrefix("[appliance/openshift/origin, DeleteBuildConfig] ")
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

	if err := oc.BuildConfigs(namespace).Delete(name); err != nil {
		glog.Errorf("Could not delete build config %s: %+v", name, err)
		return err
	}
	return nil
}

func InstantiateBuild(data []byte, obj *buildapi.BuildRequest) ([]byte, *buildapi.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, InstantiateBuild] ")

	if len(data) == 0 && obj == nil || obj != nil && len(obj.Namespace) == 0 {
		return nil, nil, errUnexpected
	}
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not instantiate openshift client: %s", err)
		return nil, nil, err
	}
	logger.Printf("openshift client: %+v\n", oc)

	/*if len(data) == 0 && obj != nil {
		result, err := oc.BuildConfigs(obj.Namespace).Instantiate(obj)
		if err != nil {
			if retry := strings.EqualFold(err.Error(), "encoding is not allowed for this codec: *recognizer.decoder") ||
				strings.HasPrefix(err.Error(), "no kind is registered for the type api."); !retry {
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
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
		if data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
	}*/

	if obj == nil {
		hco, err := codec.JSON.Decode(data).One()
		if err != nil {
			glog.Errorf("Could not create helm object: %s", err)
			return nil, nil, err
		}
		obj = new(buildapi.BuildRequest)
		if err := hco.Object(obj); err != nil {
			glog.Errorf("Could not deserialize into runtime object: %s", err)
			return nil, nil, err
		}
	}

	raw, err := oc.RESTClient.Post().Namespace(obj.Namespace).
		Resource("buildConfigs").Name(obj.Name).
		SubResource("instantiate").Body(data).DoRaw()
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
