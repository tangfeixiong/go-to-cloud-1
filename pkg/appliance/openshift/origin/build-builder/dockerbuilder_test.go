package builder

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/cloudfoundry/yagnats"
	"github.com/helm/helm-classic/codec"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
)

var (
	fake_dockerbuildproject string = "tangfx"
	fake_dockerbuildname    string = "osobuilds"

	fakeImagePath    = map[string]string{"sourcePath": "/go", "destinationDir": "/workspace"}
	fakeSourceImages = []map[string]interface{}{{
		"DockerImage": map[string]interface{}{
			"from":       "openshift/hello-openshift",
			"paths":      [...]map[string]string{fakeImagePath},
			"pullSecret": "base64:encoding"}}}

	exampleDockerBuild string = "/examples/github101.json"
)

func TestOriginDockerBuilder(t *testing.T) {
	flag.Parse()
	if f := flag.Lookup("v"); f != nil {
		f.Value.Set("5")
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	f := wd + "/../../../../.." + exampleDockerBuild

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

	clientnats := yagnats.NewClient()
	if err := clientnats.Connect(&yagnats.ConnectionInfo{
		Addr:     _nats_addrs[0],
		Username: _nats_user,
		Password: _nats_password,
	}); err != nil {
		t.Fatal(err)
	}

	mb := new(bytes.Buffer)

	ccf := util.NewClientCmdFactory()
	obj, err = RunDockerBuild(mb, obj, ccf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(obj)

	chjob := make(chan bool, 1)
	subject := obj.Namespace + "/" + obj.Name
	go func() {
		var timeout bool
		var offset int = 0
		var m *sync.Mutex = &sync.Mutex{}
		go func() {
			select {
			case <-time.After(time.Duration(1800) * time.Second):
				m.Lock()
				defer m.Unlock()
				chjob <- false
				timeout = true
			}
		}()
		var md yagnats.Message
		var id int64
		var err error
		id, err = clientnats.Subscribe(subject, func(msg *yagnats.Message) {
			md = *msg
			fmt.Printf("Got message: %d, %s\n", id, msg.Payload)
		})
		if err != nil {
			chjob <- false
			return
		}
		for false == timeout {
			time.Sleep(1000 * time.Millisecond)
			l := mb.Len()
			if l > offset {
				//fmt.Print(string(mb.Next(l - offset)))
				clientnats.Publish(subject, mb.Next(l-offset))
				offset = l
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Block until a signal is received.
	select {
	case <-c:
		fmt.Println("terminated")
	case result := <-chjob:
		fmt.Println("job end in %+v", result)
	}
}
