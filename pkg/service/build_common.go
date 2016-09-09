package service

import (
	"fmt"

	"github.com/docker/engine-api/types"
	buildapi "github.com/openshift/origin/pkg/build/api/v1"

	kapi "k8s.io/kubernetes/pkg/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/kubernetes"
)

var (
	openshift_origin_serviceaccount_builder string = "builder"
	_dockerfile                             string = "FROM busybox\nCMD [\"sh\"]"
	_timeout                                int64  = 900
)

func secretname_for_pull_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-for", buildname)
}

func secretname_for_push_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-to", buildname)
}

func verifyDockerSecret(req *osopb3.DockerBuildRequestData, obj *buildapi.Build) error {
	orchestra := kubernetes.NewOrchestration()
	if obj.Spec.Strategy.SourceStrategy != nil &&
		obj.Spec.Strategy.SourceStrategy.PullSecret == nil &&
		req.Configuration.CommonSpec.Strategy.SourceStrategy != nil &&
		req.Configuration.CommonSpec.Strategy.SourceStrategy.DockerconfigJson != nil {
		secret := secretname_for_pull_with_dockerbuilder(obj.Name)
		for k, v := range req.Configuration.CommonSpec.Strategy.DockerStrategy.DockerconfigJson.AuthConfigs {
			_, _, _, err := orchestra.VerifyDockerConfigJsonSecretAndServiceAccount(
				obj.Namespace, secret, types.AuthConfig{
					Username:      v.Username,
					Password:      v.Password,
					ServerAddress: k,
				}, openshift_origin_serviceaccount_builder)
			if err != nil {
				return err
			}
		}
		obj.Spec.Strategy.DockerStrategy.PullSecret = &kapi.LocalObjectReference{secret}
	}
	if obj.Spec.Output.PushSecret == nil &&
		req.Configuration.CommonSpec.Output.DockerconfigJson != nil {
		secret := secretname_for_push_with_dockerbuilder(obj.Name)
		for k, v := range req.Configuration.CommonSpec.Output.DockerconfigJson.AuthConfigs {
			_, _, _, err := orchestra.VerifyDockerConfigJsonSecretAndServiceAccount(
				obj.Namespace, secret, types.AuthConfig{
					Username:      v.Username,
					Password:      v.Password,
					ServerAddress: k,
				}, openshift_origin_serviceaccount_builder)
			if err != nil {
				return err
			}
		}
		obj.Spec.Output.PushSecret = &kapi.LocalObjectReference{secret}
	}
	return nil
}
