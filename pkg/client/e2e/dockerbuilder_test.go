package e2e

import (
	"os"
	"strings"
	"testing"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/client/osoc"
)

var (
	_apaas_grpc_addresses = []string{"172.17.4.50:50051", "10.3.0.22:50051"}

	_kubeconfig     string = "/data/src/github.com/openshift/origin/etc/kubeconfig"
	_kube_context   string = "openshift-origin-single"
	_kube_apiserver string = "https://172.17.4.50"

	_verbose_level = 5
	_cluster       = "notused"

	_project = "tangfx"

	_build_name = "osobuilds"

	_dockerfile = `#netcat hello world http server
FROM alpine/edge
MAINTAINER tangfeixiong <tangfx128@gmail.com>
RUN apk add --update bash ca-certificates libc6-compat netcat-openbsd && rm -rf /var/cache/apk/*
RUN echo "<html><head><title>welcome</title></head><body><h1>hello world</h1></body></html>" >> /tmp/index.html
EXPOSE 80
CMD while true; do nc -l 80 < /tmp/index.html; done`

	_override_baseimage = "gliderlabs/alpine"

	_dockpull_secret = "dockerconfigjson-osobuilds-for"

	_git_hub      = "https://github.com/tangfeixiong/docker-nc.git"
	_git_ref      = "master"
	_context_path = "/latest"

	_override_dockerfile string = "From busybox\nCMD [\"sh\"]"

	_docker_hub string = "172.17.4.50:30005/tangfx/osobuilds:latest"

	_dockerpush_secret string = "dockerconfigjson-osobuilds-to"

	_bc map[string]interface{}
)

func init() {
	if v, ok := os.LookupEnv("APAAS_GRPC_ADDRESSES"); ok && len(v) > 0 {
		_apaas_grpc_addresses = strings.Split(v, ",")
	}
}

func example_request_data() map[string]interface{} {
	_bc = map[string]interface{}{
		"Name":       _build_name,
		"Project":    _project,
		"GitURI":     _git_hub,
		"GitRef":     _git_ref,
		"GitPath":    _context_path,
		"Dockerfile": _override_dockerfile,
		"DockerPullAuth": map[string]string{
			"Username":      "tangfx",
			"Password":      "tangfx",
			"ServerAddress": "172.17.4.50:30005",
		},
		"DockerPushRepo": _docker_hub,
		"DockerPushAuth": map[string]string{
			"Username":      "tangfx",
			"Password":      "tangfx",
			"ServerAddress": "172.17.4.50:30005",
		},
	}
	return _bc
}

func TestDockerBuilder(t *testing.T) {
	exam := example_request_data()

	util := osoc.NewDockerBuildRequestDataUtility()
	data, err := util.Builder(exam["Project"].(string), exam["Name"].(string)).
		Dockerfile(exam["Dockerfile"].(string)).
		Git(exam["GitURI"].(string), exam["GitRef"].(string), exam["GitPath"].(string)).
		DockerBuildStrategy(_override_baseimage, "", ".", true, false).
		DockerBuildOutputOption(exam["DockerPushRepo"].(string), _dockerpush_secret).RequestDataForPOST()
	if err != nil {
		t.Fatal(err)
	}

	c, _, err := clientWithKubeconfig(_kubeconfig, _kube_context, _kube_apiserver)
	if err != nil {
		t.Fatal(err)
	}

	dockerPullSecret := ""
	pullAuth := exam["DockerPullAuth"].(map[string]string)
	if pullAuth["Username"] != "" && pullAuth["Password"] != "" &&
		pullAuth["ServerAddress"] != "" {
		dockerPullSecret = secretname_for_pull_with_dockerbuilder(data.Name)
		dcf, secret, sa, err := verifyDockerConfigJsonSecretAndServiceAccount(c,
			data.ProjectName, dockerPullSecret, DockerAuthConfig{
				Username:      pullAuth["Username"],
				Password:      pullAuth["Password"],
				ServerAddress: pullAuth["ServerAddress"],
			}, "builder")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dcf, secret, sa)
	}

	dockerPushSecret := ""
	pushAuth := exam["DockerPushAuth"].(map[string]string)
	if pushAuth["Username"] != "" && pushAuth["Password"] != "" &&
		pushAuth["ServerAddress"] != "" {
		dockerPushSecret = secretname_for_push_with_dockerbuilder(data.Name)
		dcf, secret, sa, err := verifyDockerConfigJsonSecretAndServiceAccount(c,
			data.ProjectName, dockerPushSecret, DockerAuthConfig{
				Username:      pushAuth["Username"],
				Password:      pushAuth["Password"],
				ServerAddress: pushAuth["ServerAddress"],
			}, "builder")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dcf, secret, sa)
	}

	factory := osoc.NewIntegrationFactory(_apaas_grpc_addresses[0])
	result, err := factory.CreateDockerBuilderIntoImage(data)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Log("Received nothing")
	} else if result.Raw != nil && len(result.Raw.ObjectJSON) > 0 {
		t.Logf("Result: %s", string(result.Raw.ObjectJSON))
	} else {
		t.Logf("Received: %+v", result)
	}

	state := make(map[string]interface{}, 0)
	state["Name"] = data.Name
	state["Project"] = data.ProjectName
	//		result.Labels = r.Metadata.Labels
	//		result.Annotations = r.Metadata.Annotations
	state["Reason"] = result.Status.Reason
	state["Message"] = result.Status.Message
	switch result.Status.OsoBuildPhase {
	case osopb3.OsoBuildStatus_New, osopb3.OsoBuildStatus_Running:
		state["Status"] = 1
	case osopb3.OsoBuildStatus_Complete:
		state["Status"] = 3
	case osopb3.OsoBuildStatus_Failed, osopb3.OsoBuildStatus_Error:
		state["Status"] = 2
	case osopb3.OsoBuildStatus_Cancelled:
		state["Status"] = 4
	default:
		state["Status"] = 0
	}

	t.Log(state)
}
