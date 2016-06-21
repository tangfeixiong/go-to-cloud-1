package main

import (
	"flag"

	"github.com/golang/glog"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/logger"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/openshift/client"
)

var (
	fakeUser    string = "tangfeixiong"
	fakeProject string = "tangfeixiong"
	fakeBuild   string = "netcat-http"

	fakeDockerfile string = "FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"
	fakeGitSecrets        = map[string]string{"github-qingyuancloud-tangfx": "user-account:keep-secret"}
	fakeGitURI     string = "https://github.com/tangfeixiong/docker-nc.git"
	fakeGitRef     string = "master"
	fakeContextDir string = "latest"

	fakeImagePath    = map[string]string{"sourcePath": "/go", "destinationDir": "/workspace"}
	fakeSourceImages = []map[string]interface{}{{
		"DockerImage": map[string]interface{}{
			"from":       "openshift/hello-openshift",
			"paths":      [...]map[string]string{fakeImagePath},
			"pullSecret": "base64:encoding"}}}
)

func init() {
	flag.Parse()
	flag.Lookup("v").Value.Set("10")
	glog.V(10).Infoln("Set glog level with 10")
}

func main() {
	if _, _, err := client.CreateBuild(fakeBuild, fakeProject, nil, fakeGitURI, fakeGitRef, fakeContextDir, nil, fakeDockerfile, nil, nil); err != nil {
		logger.Logger.Printf("Failed: %s", err)
	}
}
