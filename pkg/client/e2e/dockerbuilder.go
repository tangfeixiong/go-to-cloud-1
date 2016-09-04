package e2e

import (
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"

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

	_dockerpull_secret = "dockerconfigjson-osobuilds-for"

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
		"Name":              _build_name,
		"Project":           _project,
		"GitURI":            _git_hub,
		"GitRef":            _git_ref,
		"GitPath":           _context_path,
		"DockerfilePath":    "",
		"Dockerfile":        _override_dockerfile,
		"OverrideBaseImage": _override_baseimage,
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

func secretname_for_pull_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-for", buildname)
}

func secretname_for_push_with_dockerbuilder(buildname string) string {
	return fmt.Sprintf("dockerconfigjson-%s-to", buildname)
}

func DockerBuildWithNewConfig() (status int, ok bool) {
	exam := example_request_data()

	util := osoc.NewDockerBuildRequestDataUtility()
	pullAuth := exam["DockerPullAuth"].(map[string]string)
	pushAuth := exam["DockerPushAuth"].(map[string]string)
	data, err := util.Builder(exam["Project"].(string), exam["Name"].(string)).
		Dockerfile(exam["Dockerfile"].(string)).
		Git(exam["GitURI"].(string), exam["GitRef"].(string), exam["GitPath"].(string)).
		DockerBuildStrategy(exam["OverrideBaseImage"].(string), "", ".", true, false).
		DockerBuildOutputOption(exam["DockerPushRepo"].(string), "").
		DockerPullCredential(pullAuth["ServerAddress"], pullAuth["Username"], pullAuth["Password"]).
		DockerPushCredential(pushAuth["ServerAddress"], pushAuth["Username"], pushAuth["Password"]).
		RequestDataForPOST()
	if err != nil {
		return
	}

	factory := osoc.NewIntegrationFactory(_apaas_grpc_addresses[0])
	received, err := factory.CreateDockerBuilderIntoImage(data)
	if err != nil {
		return
	}
	if received == nil {
		return
	}
	if glog.V(2) {
		glog.V(2).Infof("DockerBuildWithNewConfig: %+v", received)
	} else {
		glog.Infof("DockerBuildWithNewConfig: %+v", received)
	}

	switch received.Status.OsoBuildPhase {
	case osopb3.OsoBuildStatus_New, osopb3.OsoBuildStatus_Pending, osopb3.OsoBuildStatus_Running:
		status = 1
	case osopb3.OsoBuildStatus_Complete:
		status = 3
	case osopb3.OsoBuildStatus_Failed, osopb3.OsoBuildStatus_Cancelled, osopb3.OsoBuildStatus_Error:
		status = 2
	//case osopb3.OsoBuildStatus_Cancelled:
	//	status = 4
	default:
		status = 0
	}
	ok = true
	return
}

func TrackingDockerBuild() (status int, ok bool) {
	exam := example_request_data()

	send := &osopb3.DockerBuildRequestData{
		Name:        exam["Name"].(string),
		ProjectName: exam["Project"].(string),
	}

	f := osoc.NewIntegrationFactory(_apaas_grpc_addresses[0])
	received, err := f.TrackDockerBuild(send)
	if glog.V(2) {
		glog.V(2).Infof("Tacking docker build: %+v, %+v", received, err)
	} else {
		glog.Infof("Tacking docker build: %+v, %+v", received, err)
	}
	if err != nil {
		return
	}
	if received == nil || received.Status == nil {
		return 1, true
	}

	switch received.Status.OsoBuildPhase {
	case osopb3.OsoBuildStatus_New, osopb3.OsoBuildStatus_Pending, osopb3.OsoBuildStatus_Running:
		status = 1
	case osopb3.OsoBuildStatus_Complete:
		status = 3
	case osopb3.OsoBuildStatus_Failed, osopb3.OsoBuildStatus_Cancelled, osopb3.OsoBuildStatus_Error:
		status = 2
	//case osopb3.OsoBuildStatus_Cancelled:
	//	status = 4
	default:
		status = 0
	}
	ok = true
	return
}

func FindDockerBuild() (status int, ok bool) {
	exam := example_request_data()

	send := &osopb3.DockerBuildRequestData{
		Name:        exam["Name"].(string),
		ProjectName: exam["Project"].(string),
	}

	f := osoc.NewIntegrationFactory(_apaas_grpc_addresses[0])
	received, err := f.RetrieveDockerBuild(send)
	if err != nil {
		return
	}
	if received == nil {
		return
	}
	if glog.V(2) {
		glog.V(2).Infof("find docker build: %+v", received)
	} else {
		glog.Infof("find docker build: %+v", received)
	}

	switch received.Status.OsoBuildPhase {
	case osopb3.OsoBuildStatus_New, osopb3.OsoBuildStatus_Pending, osopb3.OsoBuildStatus_Running:
		status = 1
	case osopb3.OsoBuildStatus_Complete:
		status = 3
	case osopb3.OsoBuildStatus_Failed, osopb3.OsoBuildStatus_Cancelled, osopb3.OsoBuildStatus_Error:
		status = 2
	//case osopb3.OsoBuildStatus_Cancelled:
	//	status = 4
	default:
		status = 0
	}
	ok = true
	return
}
