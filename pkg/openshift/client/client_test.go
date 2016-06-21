package client

import (
	"log"
	"os"

	"testing"
)

var (
	fakeUser    string = "tangfeixiong"
	fakeProject string = "gogogo"
	fakeBuild   string = "netcat-http"

	fakeDockerfile string = `"FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"`
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

func TestWhoAmI(t *testing.T) {
	if _, err := WhoAmI(); err != nil {
		t.Fatal(err)
	}
}

func TestIsUserExist(t *testing.T) {
	if _, err := RetrieveUser(fakeUser); err != nil {
		t.Fatal(err)
	}
}

func TestCreateProject(t *testing.T) {
	if _, _, err := CreateProject(fakeProject); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteProject(t *testing.T) {
	if err := DeleteProject(fakeProject); err != nil {
		t.Fatal(err)
	}
}

func TestRetrieveProjects(t *testing.T) {
	if err := RetrieveProjects(); err != nil {
		t.Fatal(err)
	}
}

func TestRetrieveProject(t *testing.T) {
	if err := RetrieveProject(fakeProject); err != nil {
		t.Fatal(err)
	}
}

func TestCreateBuild(t *testing.T) {
	if _, _, err := CreateBuild(fakeBuild, fakeProject, nil, fakeGitURI, fakeGitRef, fakeContextDir, nil, fakeDockerfile, nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	//	user, err := WhoAmI()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", user)

	//	users, err := Whole()
	//	if err != nil {
	//		log.Printf("\nCould not know user: %s", err)
	//		os.Exit(1)
	//	}
	//	log.Printf("\nuser: %v", users)

	if err := DoBasicAuth(); err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
