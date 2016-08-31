package origin

import (
	"bytes"
	"strings"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"
	oclient "github.com/openshift/origin/pkg/client"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/runtime"
	//"k8s.io/kubernetes/pkg/types"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
)

func createIntoBuild(oc *oclient.Client, data []byte, obj *buildapiv1.Build) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, createIntoBuild] ")
	if oc == nil {
		f := util.NewClientCmdFactory()
		var err error
		oc, _, err = f.Clients()
		if err != nil {
			glog.Errorf("Could not setup openshift origin client: %+v", err)
			return nil, nil, err
		}
	}
	if len(data) == 0 {
		b := &bytes.Buffer{}
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			logger.Printf("Could not serialize: %+v\n", err)
			return nil, nil, err
		}
		data = b.Bytes()
	}

	raw, err := oc.RESTClient.Verb("POST").Namespace(obj.Namespace).Resource("builds").Body(data).DoRaw()
	if err != nil {
		glog.Errorf("Could not access openshift origin: %+v", err)
		return nil, nil, err
	}
	if len(raw) == 0 {
		logger.Println("Nothing responsed")
		return nil, nil, errUnexpected
	}

	hco, err := codec.JSON.Decode(raw).One()
	if err != nil {
		logger.Printf("Could not setup helm object: %s\n", err)
		return raw, nil, err
	}
	meta := new(unversioned.TypeMeta)
	if err := hco.Object(meta); err != nil {
		logger.Printf("Could not inspect typemeta: %s\nReturn: %+v\n", err, string(raw))
		return raw, nil, err
	}

	if !strings.EqualFold("Build", meta.Kind) {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				logger.Printf("Could not inspect metadata: %+v\n", meta)
				return raw, nil, err
			}
			logger.Printf("Status inspected: %+v\n", status.Message)
			return raw, nil, errUnexpected
		}
		glog.Errorf("Unexpected result: %+v", string(raw))
		return raw, nil, errUnexpected
	}

	//meta, err := hco.Meta()
	result := new(buildapiv1.Build)
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into runtime object: %s\n", err)
		return raw, nil, err
	}
	glog.V(10).Infof("Build result: %+v\n", string(raw))
	return raw, result, nil
}

/*
	修改：     pkg/cmd/cli/cmd/newbuild.go
	修改：     pkg/generate/app/cmd/newapp.go
	修改：     pkg/generate/app/cmd/resolve.go
	修改：     pkg/generate/app/componentresolvers.go
	修改：     pkg/generate/app/dockerimagelookup.go
*/

func retrieveIntoBuild(oc *oclient.Client, project, name string) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, retrieveIntoBuild] ")
	if oc == nil {
		f := util.NewClientCmdFactory()
		var err error
		oc, _, err = f.Clients()
		if err != nil {
			logger.Printf("Could not create openshift client: %+v", err)
			return nil, nil, err
		}
	}
	raw, err := oc.RESTClient.Verb("GET").Namespace(project).Resource("builds").Name(name).DoRaw()
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

	if !strings.EqualFold("Build", meta.Kind) {
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
	result := new(buildapiv1.Build)
	if err := hco.Object(result); err != nil {
		logger.Printf("Could not decode into runtime object: %s", err)
		return raw, nil, err
	}
	glog.V(10).Infof("Build result: %+v\n", string(raw))
	return raw, result, nil
}

func V1ToBuild(obj *buildapiv1.Build) *buildapi.Build {
	tgt := &buildapi.Build{
		TypeMeta: unversioned.TypeMeta{
			Kind:       obj.Kind,
			APIVersion: obj.APIVersion,
		},
		ObjectMeta: kapi.ObjectMeta{
			Name:                       obj.Name,
			GenerateName:               obj.GenerateName,
			Namespace:                  obj.Namespace,
			SelfLink:                   obj.SelfLink,
			UID:                        obj.UID,
			ResourceVersion:            obj.ResourceVersion,
			Generation:                 obj.Generation,
			CreationTimestamp:          obj.CreationTimestamp,
			DeletionTimestamp:          obj.DeletionTimestamp,
			DeletionGracePeriodSeconds: obj.DeletionGracePeriodSeconds,
			Labels:          make(map[string]string),
			Annotations:     make(map[string]string),
			OwnerReferences: make([]kapi.OwnerReference, 0),
			Finalizers:      make([]string, 0),
		},
		Spec: buildapi.BuildSpec{
			//CommonSpec: buildapi.CommonSpec{}
			TriggeredBy: make([]buildapi.BuildTriggerCause, 0),
		},
		Status: buildapi.BuildStatus{
			Phase:                      buildapi.BuildPhase(string(obj.Status.Phase)),
			Cancelled:                  obj.Status.Cancelled,
			Reason:                     buildapi.StatusReason(string(obj.Status.Reason)),
			Message:                    obj.Status.Message,
			StartTimestamp:             obj.Status.StartTimestamp,
			CompletionTimestamp:        obj.Status.CompletionTimestamp,
			Duration:                   obj.Status.Duration,
			OutputDockerImageReference: obj.Status.OutputDockerImageReference,
			/*Config: &kapi.ObjectReference{
				Kind:            obj.Status.Config.Kind,
				Namespace:       obj.Status.Config.Namespace,
				Name:            obj.Status.Config.Name,
				UID:             obj.Status.Config.UID,
				APIVersion:      obj.Status.Config.APIVersion,
				ResourceVersion: obj.Status.Config.ResourceVersion,
				FieldPath:       obj.Status.Config.FieldPath,
			},*/
		},
	}

	for k, v := range obj.Labels {
		tgt.Labels[k] = v
	}
	for k, v := range obj.Annotations {
		tgt.Annotations[k] = v
	}
	for _, ele := range obj.OwnerReferences {
		val := kapi.OwnerReference{
			APIVersion: ele.APIVersion,
			Kind:       ele.Kind,
			Name:       ele.Name,
			UID:        ele.UID,
			Controller: ele.Controller,
		}
		tgt.OwnerReferences = append(tgt.OwnerReferences, val)
	}
	for _, ele := range obj.Finalizers {
		tgt.Finalizers = append(tgt.Finalizers, ele)
	}

	//tgt.Spec.CommonSpec.ServiceAccount
	tgt.Spec.ServiceAccount = obj.Spec.ServiceAccount
	tgt.Spec.Source = buildapi.BuildSource{
		Dockerfile: obj.Spec.Source.Dockerfile,
		Images:     make([]buildapi.ImageSource, 0),
		ContextDir: obj.Spec.Source.ContextDir,
		Secrets:    make([]buildapi.SecretBuildSource, 0),
	}
	if obj.Spec.Source.Git != nil {
		tgt.Spec.Source.Git = &buildapi.GitBuildSource{
			URI:        obj.Spec.Source.Git.URI,
			Ref:        obj.Spec.Source.Git.Ref,
			HTTPProxy:  obj.Spec.Source.Git.HTTPProxy,
			HTTPSProxy: obj.Spec.Source.Git.HTTPSProxy,
		}
	}
	if obj.Spec.Source.SourceSecret != nil {
		tgt.Spec.Source.SourceSecret = &kapi.LocalObjectReference{
			Name: obj.Spec.Source.SourceSecret.Name,
		}
	}
	for _, ele := range obj.Spec.Source.Images {
		val := buildapi.ImageSource{
			From: kapi.ObjectReference{
				Kind:            ele.From.Kind,
				Namespace:       ele.From.Namespace,
				Name:            ele.From.Name,
				UID:             ele.From.UID,
				APIVersion:      ele.From.APIVersion,
				ResourceVersion: ele.From.ResourceVersion,
				FieldPath:       ele.From.FieldPath,
			},
			Paths: make([]buildapi.ImageSourcePath, 0),
		}
		for _, e := range ele.Paths {
			v := buildapi.ImageSourcePath{
				SourcePath:     e.SourcePath,
				DestinationDir: e.DestinationDir,
			}
			val.Paths = append(val.Paths, v)
		}
		if ele.PullSecret != nil {
			val.PullSecret = &kapi.LocalObjectReference{
				Name: ele.PullSecret.Name,
			}
		}
		tgt.Spec.Source.Images = append(tgt.Spec.Source.Images, val)
	}
	for _, ele := range obj.Spec.Source.Secrets {
		val := buildapi.SecretBuildSource{
			Secret: kapi.LocalObjectReference{
				Name: ele.Secret.Name,
			},
			DestinationDir: ele.DestinationDir,
		}
		tgt.Spec.Source.Secrets = append(tgt.Spec.Source.Secrets, val)
	}
	if obj.Spec.Source.Binary != nil {
		tgt.Spec.Source.Binary = &buildapi.BinaryBuildSource{
			AsFile: obj.Spec.Source.Binary.AsFile,
		}
	}

	if obj.Spec.Revision != nil {
		tgt.Spec.Revision = new(buildapi.SourceRevision)
		if obj.Spec.Revision.Git != nil {
			tgt.Spec.Revision.Git = &buildapi.GitSourceRevision{
				Commit: obj.Spec.Revision.Git.Commit,
				Author: buildapi.SourceControlUser{
					Name:  obj.Spec.Revision.Git.Author.Name,
					Email: obj.Spec.Revision.Git.Author.Email,
				},
				Committer: buildapi.SourceControlUser{
					Name:  obj.Spec.Revision.Git.Committer.Name,
					Email: obj.Spec.Revision.Git.Committer.Email},
				Message: obj.Spec.Revision.Git.Message,
			}
		}
	}
	//tgt.Spec.Strategy = buildapi.BuildStrategy{}
	if obj.Spec.Strategy.DockerStrategy != nil {
		tgt.Spec.Strategy.DockerStrategy = &buildapi.DockerBuildStrategy{
			NoCache:        obj.Spec.Strategy.DockerStrategy.NoCache,
			Env:            make([]kapi.EnvVar, 0),
			ForcePull:      obj.Spec.Strategy.DockerStrategy.ForcePull,
			DockerfilePath: obj.Spec.Strategy.DockerStrategy.DockerfilePath,
		}
		if obj.Spec.Strategy.DockerStrategy.From != nil {
			tgt.Spec.Strategy.DockerStrategy.From = &kapi.ObjectReference{
				Kind:            obj.Spec.Strategy.DockerStrategy.From.Kind,
				Namespace:       obj.Spec.Strategy.DockerStrategy.From.Namespace,
				Name:            obj.Spec.Strategy.DockerStrategy.From.Name,
				UID:             obj.Spec.Strategy.DockerStrategy.From.UID,
				APIVersion:      obj.Spec.Strategy.DockerStrategy.From.APIVersion,
				ResourceVersion: obj.Spec.Strategy.DockerStrategy.From.ResourceVersion,
				FieldPath:       obj.Spec.Strategy.DockerStrategy.From.FieldPath,
			}
		}
		if obj.Spec.Strategy.DockerStrategy.PullSecret != nil {
			tgt.Spec.Strategy.DockerStrategy.PullSecret = &kapi.LocalObjectReference{
				Name: obj.Spec.Strategy.DockerStrategy.PullSecret.Name,
			}
		}
		for _, ele := range obj.Spec.Strategy.DockerStrategy.Env {
			val := kapi.EnvVar{
				Name:  ele.Name,
				Value: ele.Value,
			}
			if ele.ValueFrom != nil {
				val.ValueFrom = &kapi.EnvVarSource{}
				if ele.ValueFrom.FieldRef != nil {
					val.ValueFrom.FieldRef = &kapi.ObjectFieldSelector{
						APIVersion: ele.ValueFrom.FieldRef.APIVersion,
						FieldPath:  ele.ValueFrom.FieldRef.FieldPath,
					}
				}
				if ele.ValueFrom.ResourceFieldRef != nil {
					val.ValueFrom.ResourceFieldRef = &kapi.ResourceFieldSelector{
						ContainerName: ele.ValueFrom.ResourceFieldRef.ContainerName,
						Resource:      ele.ValueFrom.ResourceFieldRef.Resource,
						Divisor:       ele.ValueFrom.ResourceFieldRef.Divisor,
					}
				}
				if ele.ValueFrom.ConfigMapKeyRef != nil {
					val.ValueFrom.ConfigMapKeyRef = &kapi.ConfigMapKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
				if ele.ValueFrom.SecretKeyRef != nil {
					val.ValueFrom.SecretKeyRef = &kapi.SecretKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.SecretKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.SecretKeyRef.Key,
					}
				}
			}
			tgt.Spec.Strategy.DockerStrategy.Env = append(tgt.Spec.Strategy.DockerStrategy.Env, val)
		}
	}
	if obj.Spec.Strategy.SourceStrategy != nil {
		tgt.Spec.Strategy.SourceStrategy = &buildapi.SourceBuildStrategy{
			From: kapi.ObjectReference{
				Kind:            obj.Spec.Strategy.SourceStrategy.From.Kind,
				Namespace:       obj.Spec.Strategy.SourceStrategy.From.Namespace,
				Name:            obj.Spec.Strategy.SourceStrategy.From.Name,
				UID:             obj.Spec.Strategy.SourceStrategy.From.UID,
				APIVersion:      obj.Spec.Strategy.SourceStrategy.From.APIVersion,
				ResourceVersion: obj.Spec.Strategy.SourceStrategy.From.ResourceVersion,
				FieldPath:       obj.Spec.Strategy.SourceStrategy.From.FieldPath,
			},
			Env:              make([]kapi.EnvVar, 0),
			Scripts:          obj.Spec.Strategy.SourceStrategy.Scripts,
			Incremental:      obj.Spec.Strategy.SourceStrategy.Incremental,
			ForcePull:        obj.Spec.Strategy.SourceStrategy.ForcePull,
			RuntimeArtifacts: make([]buildapi.ImageSourcePath, 0),
		}
		if obj.Spec.Strategy.SourceStrategy.PullSecret != nil {
			tgt.Spec.Strategy.SourceStrategy.PullSecret = &kapi.LocalObjectReference{
				Name: obj.Spec.Strategy.SourceStrategy.PullSecret.Name,
			}
		}
		if obj.Spec.Strategy.SourceStrategy.RuntimeImage != nil {
			tgt.Spec.Strategy.SourceStrategy.RuntimeImage = &kapi.ObjectReference{
				Kind:            obj.Spec.Strategy.SourceStrategy.RuntimeImage.Kind,
				Namespace:       obj.Spec.Strategy.SourceStrategy.RuntimeImage.Namespace,
				Name:            obj.Spec.Strategy.SourceStrategy.RuntimeImage.Name,
				UID:             obj.Spec.Strategy.SourceStrategy.RuntimeImage.UID,
				APIVersion:      obj.Spec.Strategy.SourceStrategy.RuntimeImage.APIVersion,
				ResourceVersion: obj.Spec.Strategy.SourceStrategy.RuntimeImage.ResourceVersion,
				FieldPath:       obj.Spec.Strategy.SourceStrategy.RuntimeImage.FieldPath,
			}
		}
		for _, ele := range obj.Spec.Strategy.SourceStrategy.Env {
			val := kapi.EnvVar{
				Name:  ele.Name,
				Value: ele.Value,
			}
			if ele.ValueFrom != nil {
				val.ValueFrom = &kapi.EnvVarSource{}
				if ele.ValueFrom.FieldRef != nil {
					val.ValueFrom.FieldRef = &kapi.ObjectFieldSelector{
						APIVersion: ele.ValueFrom.FieldRef.APIVersion,
						FieldPath:  ele.ValueFrom.FieldRef.FieldPath,
					}
				}
				if ele.ValueFrom.ResourceFieldRef != nil {
					val.ValueFrom.ResourceFieldRef = &kapi.ResourceFieldSelector{
						ContainerName: ele.ValueFrom.ResourceFieldRef.ContainerName,
						Resource:      ele.ValueFrom.ResourceFieldRef.Resource,
						Divisor:       ele.ValueFrom.ResourceFieldRef.Divisor,
					}
				}
				if ele.ValueFrom.ConfigMapKeyRef != nil {
					val.ValueFrom.ConfigMapKeyRef = &kapi.ConfigMapKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
				if ele.ValueFrom.SecretKeyRef != nil {
					val.ValueFrom.SecretKeyRef = &kapi.SecretKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.SecretKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.SecretKeyRef.Key,
					}
				}
			}
			tgt.Spec.Strategy.SourceStrategy.Env = append(tgt.Spec.Strategy.SourceStrategy.Env, val)
		}
		for _, ele := range obj.Spec.Strategy.SourceStrategy.RuntimeArtifacts {
			val := buildapi.ImageSourcePath{
				SourcePath:     ele.SourcePath,
				DestinationDir: ele.DestinationDir,
			}
			tgt.Spec.Strategy.SourceStrategy.RuntimeArtifacts = append(tgt.Spec.Strategy.SourceStrategy.RuntimeArtifacts, val)
		}
	}
	if obj.Spec.Strategy.CustomStrategy != nil {
		tgt.Spec.Strategy.CustomStrategy = &buildapi.CustomBuildStrategy{
			From: kapi.ObjectReference{
				Kind:            obj.Spec.Strategy.CustomStrategy.From.Kind,
				Namespace:       obj.Spec.Strategy.CustomStrategy.From.Namespace,
				Name:            obj.Spec.Strategy.CustomStrategy.From.Name,
				UID:             obj.Spec.Strategy.CustomStrategy.From.UID,
				APIVersion:      obj.Spec.Strategy.CustomStrategy.From.APIVersion,
				ResourceVersion: obj.Spec.Strategy.CustomStrategy.From.ResourceVersion,
				FieldPath:       obj.Spec.Strategy.CustomStrategy.From.FieldPath,
			},
			Env:                make([]kapi.EnvVar, 0),
			ExposeDockerSocket: obj.Spec.Strategy.CustomStrategy.ExposeDockerSocket,
			ForcePull:          obj.Spec.Strategy.CustomStrategy.ForcePull,
			Secrets:            make([]buildapi.SecretSpec, 0),
			BuildAPIVersion:    obj.Spec.Strategy.CustomStrategy.BuildAPIVersion,
		}
		if obj.Spec.Strategy.CustomStrategy.PullSecret != nil {
			tgt.Spec.Strategy.CustomStrategy.PullSecret = &kapi.LocalObjectReference{
				Name: obj.Spec.Strategy.CustomStrategy.PullSecret.Name,
			}
		}
		for _, ele := range obj.Spec.Strategy.CustomStrategy.Env {
			val := kapi.EnvVar{
				Name:  ele.Name,
				Value: ele.Value,
			}
			if ele.ValueFrom != nil {
				val.ValueFrom = &kapi.EnvVarSource{}
				if ele.ValueFrom.FieldRef != nil {
					val.ValueFrom.FieldRef = &kapi.ObjectFieldSelector{
						APIVersion: ele.ValueFrom.FieldRef.APIVersion,
						FieldPath:  ele.ValueFrom.FieldRef.FieldPath,
					}
				}
				if ele.ValueFrom.ResourceFieldRef != nil {
					val.ValueFrom.ResourceFieldRef = &kapi.ResourceFieldSelector{
						ContainerName: ele.ValueFrom.ResourceFieldRef.ContainerName,
						Resource:      ele.ValueFrom.ResourceFieldRef.Resource,
						Divisor:       ele.ValueFrom.ResourceFieldRef.Divisor,
					}
				}
				if ele.ValueFrom.ConfigMapKeyRef != nil {
					val.ValueFrom.ConfigMapKeyRef = &kapi.ConfigMapKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.ConfigMapKeyRef.Key,
					}
				}
				if ele.ValueFrom.SecretKeyRef != nil {
					val.ValueFrom.SecretKeyRef = &kapi.SecretKeySelector{
						LocalObjectReference: kapi.LocalObjectReference{
							Name: ele.ValueFrom.SecretKeyRef.LocalObjectReference.Name,
						},
						Key: ele.ValueFrom.SecretKeyRef.Key,
					}
				}
			}
			tgt.Spec.Strategy.CustomStrategy.Env = append(tgt.Spec.Strategy.CustomStrategy.Env, val)
		}
		for _, ele := range obj.Spec.Strategy.CustomStrategy.Secrets {
			val := buildapi.SecretSpec{
				SecretSource: kapi.LocalObjectReference{
					Name: ele.SecretSource.Name,
				},
				MountPath: ele.MountPath,
			}
			tgt.Spec.Strategy.CustomStrategy.Secrets = append(tgt.Spec.Strategy.CustomStrategy.Secrets, val)
		}
	}
	if obj.Spec.Strategy.JenkinsPipelineStrategy != nil {
		tgt.Spec.Strategy.JenkinsPipelineStrategy = &buildapi.JenkinsPipelineBuildStrategy{
			Jenkinsfile:     obj.Spec.Strategy.JenkinsPipelineStrategy.Jenkinsfile,
			JenkinsfilePath: obj.Spec.Strategy.JenkinsPipelineStrategy.JenkinsfilePath,
		}
	}
	//tgt.Spec.Output = buildapi.BuildOutput{}
	if obj.Spec.Output.To != nil {
		tgt.Spec.Output.To = &kapi.ObjectReference{
			Kind:            obj.Spec.Output.To.Kind,
			Namespace:       obj.Spec.Output.To.Namespace,
			Name:            obj.Spec.Output.To.Name,
			UID:             obj.Spec.Output.To.UID,
			APIVersion:      obj.Spec.Output.To.APIVersion,
			ResourceVersion: obj.Spec.Output.To.ResourceVersion,
			FieldPath:       obj.Spec.Output.To.FieldPath,
		}
	}
	if obj.Spec.Output.PushSecret != nil {
		tgt.Spec.Output.PushSecret = &kapi.LocalObjectReference{
			Name: obj.Spec.Output.PushSecret.Name,
		}
	}

	tgt.Spec.Resources = kapi.ResourceRequirements{
		Limits:   make(map[kapi.ResourceName]resource.Quantity),
		Requests: make(map[kapi.ResourceName]resource.Quantity),
	}
	for k, v := range obj.Spec.Resources.Limits {
		tgt.Spec.Resources.Limits[kapi.ResourceName(k)] = v
	}
	for k, v := range obj.Spec.Resources.Requests {
		tgt.Spec.Resources.Requests[kapi.ResourceName(k)] = v
	}
	tgt.Spec.PostCommit = buildapi.BuildPostCommitSpec{
		Command: make([]string, 0),
		Args:    make([]string, 0),
		Script:  obj.Spec.PostCommit.Script,
	}
	for _, ele := range obj.Spec.PostCommit.Command {
		tgt.Spec.PostCommit.Command = append(tgt.Spec.PostCommit.Command, ele)
	}
	for _, ele := range obj.Spec.PostCommit.Args {
		tgt.Spec.PostCommit.Args = append(tgt.Spec.PostCommit.Args, ele)
	}

	tgt.Spec.CompletionDeadlineSeconds = obj.Spec.CompletionDeadlineSeconds

	for _, ele := range obj.Spec.TriggeredBy {
		val := buildapi.BuildTriggerCause{
			Message: ele.Message,
		}
		if ele.GenericWebHook != nil {
			val.GenericWebHook = &buildapi.GenericWebHookCause{
				Secret: ele.GenericWebHook.Secret,
			}
			if ele.GenericWebHook.Revision != nil {
				val.GenericWebHook.Revision = new(buildapi.SourceRevision)
				if ele.GenericWebHook.Revision.Git != nil {
					val.GenericWebHook.Revision.Git = &buildapi.GitSourceRevision{
						Commit: ele.GenericWebHook.Revision.Git.Commit,
						Author: buildapi.SourceControlUser{
							Name:  ele.GenericWebHook.Revision.Git.Author.Name,
							Email: ele.GenericWebHook.Revision.Git.Author.Email,
						},
						Committer: buildapi.SourceControlUser{
							Name:  ele.GenericWebHook.Revision.Git.Committer.Name,
							Email: ele.GenericWebHook.Revision.Git.Committer.Email},
						Message: ele.GenericWebHook.Revision.Git.Message,
					}
				}
			}
		}
		if ele.GitHubWebHook != nil {
			val.GitHubWebHook = &buildapi.GitHubWebHookCause{
				Secret: ele.GitHubWebHook.Secret,
			}
			if ele.GitHubWebHook.Revision != nil {
				val.GitHubWebHook.Revision = new(buildapi.SourceRevision)
				if ele.GitHubWebHook.Revision.Git != nil {
					val.GitHubWebHook.Revision.Git = &buildapi.GitSourceRevision{
						Commit: ele.GitHubWebHook.Revision.Git.Commit,
						Author: buildapi.SourceControlUser{
							Name:  ele.GitHubWebHook.Revision.Git.Author.Name,
							Email: ele.GitHubWebHook.Revision.Git.Author.Email,
						},
						Committer: buildapi.SourceControlUser{
							Name:  ele.GitHubWebHook.Revision.Git.Committer.Name,
							Email: ele.GitHubWebHook.Revision.Git.Committer.Email},
						Message: ele.GitHubWebHook.Revision.Git.Message,
					}
				}
			}
		}
		if ele.ImageChangeBuild != nil {
			val.ImageChangeBuild = &buildapi.ImageChangeCause{
				ImageID: ele.ImageChangeBuild.ImageID,
			}
			if ele.ImageChangeBuild.FromRef != nil {
				val.ImageChangeBuild.FromRef = &kapi.ObjectReference{
					Kind:            ele.ImageChangeBuild.FromRef.Kind,
					Namespace:       ele.ImageChangeBuild.FromRef.Namespace,
					Name:            ele.ImageChangeBuild.FromRef.Name,
					UID:             ele.ImageChangeBuild.FromRef.UID,
					APIVersion:      ele.ImageChangeBuild.FromRef.APIVersion,
					ResourceVersion: ele.ImageChangeBuild.FromRef.ResourceVersion,
					FieldPath:       ele.ImageChangeBuild.FromRef.FieldPath,
				}
			}
		}
		tgt.Spec.TriggeredBy = append(tgt.Spec.TriggeredBy, val)
	}

	if obj.Status.Config != nil {
		tgt.Status.Config = &kapi.ObjectReference{
			Kind:            obj.Status.Config.Kind,
			Namespace:       obj.Status.Config.Namespace,
			Name:            obj.Status.Config.Name,
			UID:             obj.Status.Config.UID,
			APIVersion:      obj.Status.Config.APIVersion,
			ResourceVersion: obj.Status.Config.ResourceVersion,
			FieldPath:       obj.Status.Config.FieldPath,
		}
	}

	return tgt
}

func ConvertV1IntoBuild(data []byte, obj *buildapiv1.Build) ([]byte, *buildapi.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, convertV1IntoBuild] ")

	var hco *codec.Object
	var err error
	if len(data) > 0 && obj == nil {
		if hco, err = codec.JSON.Decode(data).One(); err != nil {
			logger.Printf("Could not setup decoder (Build): %+v", err)
			return nil, nil, err
		}
		obj = new(buildapiv1.Build)
		if err := hco.Object(obj); err != nil {
			logger.Printf("Could not decode into build: %+v", err)
			return nil, nil, err
		}
	}
	if obj == nil {
		return nil, nil, errBadRequest
	}
	tgt := new(buildapi.Build)

	b := &bytes.Buffer{}
	if err = codec.JSON.Encode(b).One(&obj.TypeMeta); err != nil {
		logger.Printf("Could not serialize build (TypeMeta): %+v", err)
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
		logger.Printf("Could not serialize build (ObjectMeta): %+v", err)
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
	if tgt.ObjectMeta.Name == "" || tgt.ObjectMeta.Namespace == "" {
		glog.Errorf("Invalid destination object from meta: %s, %s", tgt.ObjectMeta.Name, tgt.ObjectMeta.Namespace)
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Spec); err != nil {
		logger.Printf("Could not serialize build (BuildSpec): %+v", err)
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
		logger.Printf("Could not serialize build (BuildStatus): %+v", err)
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

func ConvertBuildIntoV1(data []byte, obj *buildapi.Build) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, convertBuildIntoV1] ")

	var hco *codec.Object
	var err error
	if len(data) > 0 && obj == nil {
		if hco, err = codec.JSON.Decode(data).One(); err != nil {
			logger.Printf("Could not setup decoder (Build): %+v", err)
			return nil, nil, err
		}
		obj = new(buildapi.Build)
		if err := hco.Object(obj); err != nil {
			logger.Printf("Could not decode into build: %+v", err)
			return nil, nil, err
		}
	}
	if obj == nil {
		return nil, nil, errBadRequest
	}
	tgt := new(buildapiv1.Build)

	b := &bytes.Buffer{}
	if err = codec.JSON.Encode(b).One(&obj.TypeMeta); err != nil {
		logger.Printf("Could not serialize build (TypeMeta): %+v", err)
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
		logger.Printf("Could not serialize build (ObjectMeta): %+v", err)
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
	if tgt.ObjectMeta.Name == "" || tgt.ObjectMeta.Namespace == "" {
		glog.Errorf("Invalid destination object from meta: %s, %s", tgt.ObjectMeta.Name, tgt.ObjectMeta.Namespace)
		return nil, nil, errUnexpected
	}
	b.Reset()
	if err = codec.JSON.Encode(b).One(&obj.Spec); err != nil {
		logger.Printf("Could not serialize build (BuildSpec): %+v", err)
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
		logger.Printf("Could not serialize build (BuildStatus): %+v", err)
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

func CreateIntoBuildWithV1(obj *buildapiv1.Build) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, CreateIntoBuildWithV1] ")

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

	return createIntoBuild(oc, data, obj)
}

func CreateBuild(obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	return createBuild(nil, obj)
}

func createBuild(data []byte, obj *buildapi.Build) ([]byte, *buildapi.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, createIntoBuild] ")

	f := util.NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		logger.Printf("Could not create openshift origin client: %+v", err)
		return nil, nil, err
	}
	glog.V(10).Infof("Create openshift origin client: %+v", oc)

	if len(data) == 0 {
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
		b := new(bytes.Buffer)
		if err := codec.JSON.Encode(b).One(obj); err != nil {
			glog.Errorf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
		data = b.Bytes()
		kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
		if data, err = runtime.Encode(kapi.Codecs.LegacyCodec(buildapi.SchemeGroupVersion),
			obj); err != nil {
			logger.Printf("Could not serialize runtime object: %+v", err)
			return nil, nil, err
		}
	}
	glog.V(10).Infof("Build object: %s\n", string(data))

	if obj == nil {
		hco, err := codec.JSON.Decode(data).One()
		if err != nil {
			glog.Errorf("Could not setup openshift origin codec: %s", err)
			return nil, nil, err
		}
		obj = new(buildapi.Build)
		if err := hco.Object(obj); err != nil {
			logger.Printf("Could not codec with openshift origin: %s", err)
			return nil, nil, err
		}
		/*if obj.Name == "" || obj.Namespace == "" {
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
		}*/
	}
	glog.V(10).Infof("Build: %+v\n", obj)

	d, v1, err := ConvertBuildIntoV1(data, obj)
	if err != nil {
		return nil, nil, err
	}
	raw, result, err := createIntoBuild(oc, d, v1)
	if err != nil {
		return nil, nil, err
	}
	rd, ro, err := ConvertV1IntoBuild(raw, result)
	if err != nil {
		return nil, nil, err
	}
	return rd, ro, nil
}

// gitRef: branch name, tag name, or commit revision
func CreateDockerBuildExample(name, projectName string, gitSecret map[string]string, gitURI, gitRef, contextDir string, sourceImages []map[string]interface{}, dockerfile string, buildSecrets []map[string]interface{}, buildStrategy map[string]interface{}) ([]byte, *buildapi.Build, error) {
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

	return CreateBuild(obj)
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

	return CreateIntoBuildWithV1(obj)
}

func RetrieveBuild(namespace, name string) ([]byte, *buildapiv1.Build, error) {
	logger.SetPrefix("[appliance/openshift/origin, RetrieveBuild] ")

	if len(name) == 0 || len(namespace) == 0 {
		return nil, nil, errUnexpected
	}
	f := util.NewClientCmdFactory()
	oc, _, err := f.Clients()
	if err != nil {
		glog.Errorf("Could not create openshift client: %+v", err)
		return nil, nil, err
	}
	logger.Printf("get builds with namespace: %+v name: %v\n", namespace, name)

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
		glog.Errorf("Could not find any data")
		return nil, nil, errUnexpected
	}
	if bytes.IndexAny(raw, "404:") == 0 {
		logger.Printf("Result:\n%s\n", string(raw))
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
	glog.V(10).Infof("Meta: %+v", meta)
	if ok := strings.EqualFold("Build", meta.Kind); !ok {
		if strings.EqualFold("Status", meta.Kind) {
			status := new(unversioned.Status)
			if err := hco.Object(status); err != nil {
				glog.Errorf("Could not parse status: %+v", err)
				return nil, nil, err
			}
			glog.Warningf("Status resource is received: %+v", status)
			return nil, nil, nil
		}
		glog.Errorf("Could not know metadata: %+v", meta)
		return nil, nil, errUnexpected
	}
	glog.V(10).Infof("Helm Object: %+v", hco)
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
