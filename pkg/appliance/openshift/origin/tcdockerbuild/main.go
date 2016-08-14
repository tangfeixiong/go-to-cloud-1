package main

import (
	"flag"
	"fmt"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
)

var (
	fakeUser    string = "system:admin"
	fakeProject string = "tangfx"
	fakeBuild   string = "netcat-http"

	fakeDockerfile string = `"FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"`

	fakeGitSecrets        = map[string]string{"gogs": "tangfx:tangfx"}
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

func main() {
	flag.Parse()
	f := flag.Lookup("v")
	if f != nil {
		f.Value.Set("10")
	}

	if _, _, err := origin.CreateDockerBuildV1Example(fakeBuild, fakeProject,
		nil, fakeGitURI, fakeGitRef, fakeContextDir,
		nil, fakeDockerfile, nil, nil); err != nil {
		fmt.Printf("Failed: %s", err)
	}
}
