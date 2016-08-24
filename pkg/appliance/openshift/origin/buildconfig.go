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
	"k8s.io/kubernetes/pkg/runtime"
)

func CreateBuildConfigWith(obj *buildapi.BuildConfig) ([]byte, *buildapi.BuildConfig, error) {
	return createBuildConfig(nil, obj)
}

func CreateBuildConfigWithV1(obj *buildapiv1.BuildConfig) {

}

func CreateBuildConfigFromArbitray(data []byte) ([]byte, *buildapi.BuildConfig, error) {
	return createBuildConfig(data, nil)
}

func createBuildConfig(data []byte, obj *buildapi.BuildConfig) ([]byte, *buildapi.BuildConfig, error) {
	logger = log.New(os.Stdout, "[createBuildConifg] ", log.LstdFlags|log.Lshortfile)

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

func RetrieveBuildConfigs(namespace string) error {
	f := NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		return err
	}
	logger.Printf("openshift client: %+v\n", oc)

	result, err := oc.BuildConfigs(namespace).List(kapi.ListOptions{})
	if err != nil {
		return err
	}
	logger.Println(result)
	return nil
}

func RetrieveBuildConfig(namespace, name string) ([]byte, *buildapi.BuildConfig, error) {
	logger = log.New(os.Stdout, "[RetrieveBuildConfig] ", log.LstdFlags|log.Lshortfile)

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

	result, err := oc.BuildConfigs(namespace).Get(name)
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

	raw, err := oc.RESTClient.Get().Resource("buildConfigs").Name(name).DoRaw()
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

func DeleteBuildConfig(namespace, name string) error {
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
	logger = log.New(os.Stdout, "[, InstantiateBuild] ", log.LstdFlags|log.Lshortfile)

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
