package openshift

import (
	"testing"

	"github.com/openshift/origin/pkg/version"
)

var (
	u   = "tangfeixiong"
	p   = "tangfeixiong"
	ns  = "tangfeixiong"
	bld = "tangfeixiong"
)

func TestVersion(t *testing.T) {
	thisVersion := version.Get().String()
	t.Log(thisVersion)
}

func TestSignio(t *testing.T) {
	token, err := SignIn(u, p)
	if err != nil {
		t.Fatal(err)
	}
	if err := SignOut(token); err != nil {
		t.Fatal(err)
	}
}

func TestRetrieveProjects(t *testing.T) {
	ws := EnterWorkspace(u, p)
	if ws == nil {
		t.Fatal(errUnexpected)
	}
	app := NewProjectAppliance(ws)
	val, err := app.RetrieveProjects()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(val)
}

var aws Workspace

func TestShowProject(t *testing.T) {
	if aws == nil {
		aws = EnterWorkspace(u, p)
		if aws == nil {
			t.Fatal(errUnexpected)
		}
	}
	app := aws.ProjectAppliance()
	val, err := app.RetrieveProject(ns)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(val)
}

func TestRetrieveDockerImageBuilders(t *testing.T) {
	if aws == nil {
		aws = EnterWorkspace(u, p)
		if aws == nil {
			t.Fatal(errUnexpected)
		}
	}
	app := aws.DockerImageAppliance()
	val, obj, err := app.RetrieveDockerImageBuilders(ns)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rawdata: %s\nbuildlist: %v\n", string(val), obj)
}

func TestRetrieveDockerImageBuilder(t *testing.T) {
	if aws == nil {
		aws = EnterWorkspace(u, p)
		if aws == nil {
			t.Fatal(errUnexpected)
		}
	}
	app := aws.DockerImageAppliance()
	val, obj, err := app.RetrieveDockerImageBuilder(ns, bld)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rawdata: %s\nbuild: %v\n", string(val), obj)
}

var (
	dockerfile      string = `"FROM alpine:edge\nRUN apk add --update netcat-openbsd && rm -rf /var/cache/apk/*\nCOPY entrypoint.sh /\nENTRYPOINT [\"/entrypoint.sh\"]\nCMD [\"nc\"]"`
	gituri          string = "https://github.com/tangfeixiong/docker-nc.git"
	branchTagCommit string = "master"
	contextDir      string = "latest"
	gitsecret       string = "github-qingyuancloud-tangfx"
)

func TestBuildPushDockerImage(t *testing.T) {
	if aws == nil {
		aws = EnterWorkspace(u, p)
		if aws == nil {
			t.Fatal(errUnexpected)
		}
	}
	app := aws.DockerImageAppliance()
	val, obj, err := app.BuildDockerImageIntoRegistryFrom("tangfeixiong", "netcat-http-dev", nil, gituri, branchTagCommit, contextDir, nil, dockerfile, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("rawdata: %s\nbuild: %v\n", string(val), obj)
}
