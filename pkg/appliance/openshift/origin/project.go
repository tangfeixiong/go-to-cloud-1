package origin

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	projectapi "github.com/openshift/origin/pkg/project/api"
	projectapiv1 "github.com/openshift/origin/pkg/project/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
)

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
	logger := log.New(os.Stdout, "[createBuildConifg] ", log.LstdFlags|log.Lshortfile)

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
	obj := new(projectapiv1.Project)
	obj.Kind = "Project"
	obj.APIVersion = projectapiv1.SchemeGroupVersion.Version
	obj.Name = name
	obj.Spec.Finalizers = []kapiv1.FinalizerName{projectapiv1.FinalizerOrigin, kapiv1.FinalizerKubernetes}
	for _, v := range finalizers {
		obj.Spec.Finalizers = append(obj.Spec.Finalizers, kapiv1.FinalizerName(v))
	}
	//obj.Spec.Finalizers = append(obj.Spec.Finalizers, kapiv1.FinalizerName(FinalizerVender))

	b := new(bytes.Buffer)
	if err := codec.JSON.Encode(b).One(obj); err != nil {
		glog.Errorf("Could not serialize into bytes: %+v", err)
		return nil, nil, err
	}

	return createProject(b.Bytes(), nil)
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
		kapi.Scheme.AddKnownTypes(projectapi.SchemeGroupVersion, &projectapi.Project{})
		data, err := runtime.Encode(kapi.Codecs.LegacyCodec(projectapi.SchemeGroupVersion),
			result)
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
	//kapi.Scheme.AddKnownTypes(projectapi.SchemeGroupVersion, &projectapi.Project{})
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
