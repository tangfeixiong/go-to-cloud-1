package builder

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"testing"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/helm/helm-classic/codec"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
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

	exampleBuild string = "/examples/github101.json"
	buildName    string = "osobuilds"
	buildProject string = "tangfx"
)

func TestOriginDockerBuilder(t *testing.T) {
	_ = flag.Int("loglevel", 5, "loglevel binding with glog v")
	flag.Parse()
	if f := flag.Lookup("v"); f != nil {
		f.Value.Set("5")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	f := wd + "/../../../../.." + exampleBuild

	b, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("read json:\n%s\n", string(b))

	data := &buildapiv1.Build{}
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDecoder(), b, data); err != nil {
		t.Fatal(err)
	}
	//if err := runtime.DecodeInto(kapi.Codecs.LegacyCodec(buildapiv1.SchemeGroupVersion), b, data); err != nil {
	//	t.Fatal(err)
	//}
	//	if hco, err := codec.JSON.Decode(b).One(); err != nil {
	//		t.Fatal(err)
	//	} else {
	//		if err = hco.Object(&data); err != nil {
	//			t.Fatal(err)
	//		}
	//	}

	obj := &buildapi.Build{}
	buf := &bytes.Buffer{}
	if err := codec.JSON.Encode(buf).One(data.TypeMeta); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf.Bytes()))
	if hco, err := codec.JSON.Decode(buf.Bytes()).One(); err != nil {
		t.Fatal(err)
	} else {
		obj.TypeMeta = unversioned.TypeMeta{}
		if err = hco.Object(&obj.TypeMeta); err != nil {
			t.Fatal(err)
		}
	}
	buf.Reset()
	if err := codec.JSON.Encode(buf).One(data.ObjectMeta); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf.Bytes()))
	if hco, err := codec.JSON.Decode(buf.Bytes()).One(); err != nil {
		t.Fatal(err)
	} else {
		obj.ObjectMeta = kapi.ObjectMeta{}
		if err = hco.Object(&obj.ObjectMeta); err != nil {
			t.Fatal(err)
		}
	}
	buf.Reset()
	if err := codec.JSON.Encode(buf).One(data.Spec); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf.Bytes()))
	if hco, err := codec.JSON.Decode(buf.Bytes()).One(); err != nil {
		t.Fatal(err)
	} else {
		obj.Spec = buildapi.BuildSpec{}
		if err = hco.Object(&obj.Spec); err != nil {
			t.Fatal(err)
		}
	}
	//	buf.Reset()
	//	if err := codec.JSON.Encode(buf).One(data.Spec.CommonSpec); err != nil {
	//		t.Fatal(err)
	//	}
	//	if hco, err := codec.JSON.Decode(buf.Bytes()).One(); err != nil {
	//		t.Fatal(err)
	//	} else {
	//		obj.Spec.CommonSpec = buildapi.CommonSpec{}
	//		if err = hco.Object(&obj.Spec.CommonSpec); err != nil {
	//			t.Fatal(err)
	//		}
	//	}
	buf.Reset()
	if err := codec.JSON.Encode(buf).One(data.Status); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(buf.Bytes()))
	if hco, err := codec.JSON.Decode(buf.Bytes()).One(); err != nil {
		t.Fatal(err)
	} else {
		obj.Status = buildapi.BuildStatus{}
		if err = hco.Object(&obj.Status); err != nil {
			t.Fatal(err)
		}
	}

	fmt.Printf("type and object: %+v\n", obj)
	fmt.Printf("build spec >>>\ndockerfile: %+v\ngit: %+v\nstrategy: %+v\noutput: %+v\n",
		obj.Spec.Source.Dockerfile, obj.Spec.Source.Git,
		obj.Spec.Strategy, obj.Spec.Output)
	if obj.Spec.Source.Dockerfile != nil {
		fmt.Printf("Dockerfile: %+v\n", *obj.Spec.Source.Dockerfile)
	}
	if obj.Spec.Source.Git != nil {
		fmt.Printf("Git: %+v\n", *obj.Spec.Source.Git)
	}
	if obj.Spec.Output.To != nil {
		fmt.Printf("Output to: %+v\n", *obj.Spec.Output.To)
	}
	fmt.Printf("build status: %+v\n", obj.Status)

	ccf := util.NewClientCmdFactory()
	obj, err = RunDockerBuild(os.Stdout, obj, ccf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(obj)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	<-c

}

func fakeDockerBuild() {
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
