package origin

import (
	"log"
	"os"

	"testing"
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

func TestBuild_simplecreate(t *testing.T) {
	data, _, err := CreateBuild(fakeBuild, fakeProject, nil, fakeGitURI, fakeGitRef, fakeContextDir, nil, fakeDockerfile, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Build:\n%s\n", string(data))
}

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

func TestProject_create(t *testing.T) {
	data, _, err := CreateProject(fakeProject)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Project:\n%s\n", string(data))
}

func TestProject_retrieve(t *testing.T) {
	data, _, err := RetrieveProject("default")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Project:\n%s\n", string(data))
}

func TestProject_delete(t *testing.T) {
	if err := DeleteProject(fakeProject); err != nil {
		t.Fatal(err)
	}
}

func TestReadProjects(t *testing.T) {
	if err := RetrieveProjects(); err != nil {
		t.Fatal(err)
	}
}

func TestBuildConfig_retrieve(t *testing.T) {
	data, _, err := RetrieveBuildConfig("default", "fake")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Logf("BuildConfig:\n%s\n", string(data))
	} else {
		t.Log("nothing")
	}
}

func TestBuild_retrieve(t *testing.T) {
	data, _, err := RetrieveBuild("default", "fake")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 0 {
		t.Logf("Build:\n%s\n", string(data))
	} else {
		t.Log("nothing")
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
