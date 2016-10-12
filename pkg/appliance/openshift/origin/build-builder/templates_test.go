package builder

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/template"

	"github.com/helm/helm-classic/codec"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

var (
	objmetaTplOpt = ObjectMetaTemplateOption{
		Name:      "osospringbootapp",
		Namespace: "fake",
		Annotations: []utility.NameValueStringPair{
			{Name: "anno1name", Value: "anno1value"},
		},
		Labels: []utility.KeyValueStringPair{
			{Key: "label1key", Value: "label1value"},
			{Key: "label2key", Value: "label2value"},
		},
	}

	commonspecTplOpt = CommonSpecTemplateOption{
		SourceType:   "Git",
		BinarySource: BinarySourceTemplateOption{ArchiveRaw: []byte("abc"), AsFile: "/web/app.war"},
		Dockerfile:   `FROM busybox\nCMD [\"/bin/sh\"]`,
		GitSource: GitSourceTemplateOption{
			URI:  "https://github.com/tangfeixiong/osev3-examples",
			Ref:  "master",
			Path: "/spring-boot/sample-microservices-springboot/web",
		},
		ImageSource: []ImageSourceTemplateOption{
			{
				Kind: "DockerImage",
				Name: "busybox:latest",
				Paths: []ImageSourcePath{
					{
						SourcePath:     "/go",
						DestinationDir: "/go",
					},
				},
				PullAuth: AuthTemplateOption{
					Name:           "docker auth",
					CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
				},
			},
			{
				Kind: "ImageStreamImage",
				Name: "busybox:latest",
				Paths: []ImageSourcePath{
					{
						SourcePath:     "/go",
						DestinationDir: "/go",
					},
				},
			},
		},
		ContextDir: "/spring-boot/sample-microservices-springboot/web",
		SourceAuth: AuthTemplateOption{
			Name:           "Github Auth",
			CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
		},
		SecretBuildSource: []SecretBuildSourceTemplateOption{
			{
				AuthTemplateOption: AuthTemplateOption{Name: "Referenced secret"},
				DestinationDir:     "/",
			},
		},
		SourceRevision: SourceRevisionTemplateOption{},
		Strategy: BuildStrategyTemplateOption{
			StrategyType: "Docker",
			SourceStrategy: SourceStrategy{
				BasicStrategyTemplateOption: BasicStrategyTemplateOption{
					ImageKind: "Docker Image",
					ImageName: "busybox:latest",
					PullAuth: AuthTemplateOption{
						Name:           "Docker Auth",
						CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
					},
					Env: []utility.NameValueStringPair{
						{Name: "ENV1NAME", Value: "ENV1VALUE"},
						{Name: "ENV2NAME", Value: "ENV2VALUE"},
					},
				},
				Incremental:      true,
				RuntimeImageKind: "DockerImage",
				RuntimeImageName: "busybox:latest",
				RuntimeArtifacts: []ImageSourcePath{
					{SourcePath: "/go", DestinationDir: "/go"},
				},
			},
			DockerStrategy: DockerStrategy{
				BasicStrategyTemplateOption: BasicStrategyTemplateOption{
					ImageKind: "Docker Image",
					ImageName: "busybox:latest",
					PullAuth: AuthTemplateOption{
						Name:           "Docker Auth",
						CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
					},
					Env: []utility.NameValueStringPair{
						{Name: "ENV1NAME", Value: "ENV1VALUE"},
					},
				},
				NoCache:        true,
				DockerfilePath: "/docker",
			},
			CustomStrategy: CustomStrategy{
				BasicStrategyTemplateOption: BasicStrategyTemplateOption{
					ImageKind: "Docker Image",
					ImageName: "busybox:latest",
					PullAuth: AuthTemplateOption{
						Name:           "Docker Auth",
						CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
					},
				},
				ExposeDockerSocket: true,
				Secrets: []SecretSpec{
					{
						AuthTemplateOption: AuthTemplateOption{Name: "fake"},
						MountPath:          "/build",
					},
				},
			},
			JenkinsPipelineStrategy: JenkinsPipelineStrategy{
				JenkinsfilePath: "fake",
			},
		},
		Output: BuildOutputTemplateOption{
			ImageKind: "DockerImage",
			ImageName: "172.17.4.50:30005/tangfeixiong/demo:latest",
			PushAuth: AuthTemplateOption{
				Name:           "local",
				CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
			},
		},
		/*ResourceLimits: []ResourceTemplateOption{
			{Name: "cpu", Quantity: "500m"},
			{Name: "memory", Quantity: "5Gi"},
			{Name: "stroage", Quantity: "500Gi"},
			{Name: "alpha.kubernetes.io/nvidia-gpu", Quantity: "2"},
		},
		ResourceRequests: []ResourceTemplateOption{
			{Name: "cpu", Quantity: "500m"},
			{Name: "memory", Quantity: "5Gi"},
			{Name: "stroage", Quantity: "500Gi"},
			{Name: "alpha.kubernetes.io/nvidia-gpu", Quantity: "2"},
		},*/
		PostCommitCommand:         []string{"cmd"},
		PostCommitArgs:            []string{"arg1", "arg2"},
		CompletionDeadlineSeconds: 30,
		// Deprecated
		SimpleGitOption: SimpleGitOption{
			GitURI:    "https://github.com/tangfeixiong/osev3-examples",
			GitRef:    "master",
			FromKind:  "DockerImage",
			FromName:  "tangfeixiong/springboot-sti:gitcommit-1125149-0901T2236",
			ForcePull: false,
			ToKind:    "DockerImage",
			ToName:    "172.17.4.50:30005/tangfx/osospringbootapp",
		},
	}

	bTmplOpt = BuildTemplateOption{
		ObjectMetaTemplateOption: ObjectMetaTemplateOption{
			Name:      "osospringbootapp",
			Namespace: "fake",
			Annotations: []utility.NameValueStringPair{
				{Name: "anno1name", Value: "anno1value"},
			},
		},
		CommonSpecTemplateOption: commonspecTplOpt,
	}

	bTmpl = SourceBuildConfigTemplate["BuildForGitByManuallyTriggered"]

	bcTmplOpt = BuildConfigTemplateOption{
		ObjectMetaTemplateOption: ObjectMetaTemplateOption{
			Name:      "osospringbootapp",
			Namespace: "fake",
			Labels: []utility.KeyValueStringPair{
				{Key: "label1key", Value: "label1value"},
				{Key: "label2key", Value: "label2value"},
			},
		},
		TriggerPolicy: []interface{}{
			GitHubHookTemplateOption{
				WebHookTemplateOption{Secret: "git secret"},
			},
			WebHookTemplateOption{Secret: "web secret"},
			ImageChangeHookTemplateOption{},
			ImageChangeHookTemplateOption{Kind: "ImageStreamTag", Name: "not from default strategy"},
			ConfigChangeHookTemplateOption{},
		},
		CommonSpecTemplateOption: commonspecTplOpt,
	}

	bcTmpl = SourceBuildConfigTemplate["buildconfig.json.tpl"]
)

func TestTemplate_One(t *testing.T) {
	te := template.New("build and config template").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)

	te = template.Must(te.Parse(bTmpl))
	if err := te.Execute(os.Stdout, bTmplOpt); err != nil {
		t.Fatal(err)
	}

	te = template.Must(te.Parse(bcTmpl))
	if err := te.Execute(os.Stdout, bcTmplOpt); err != nil {
		t.Fatal(err)
	}
}

func TestTemplate_b(t *testing.T) {
	buf := &bytes.Buffer{}
	te := template.New("build template").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)

	te = template.Must(te.Parse(bTmpl))
	if err := te.Execute(buf, bTmplOpt); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	obj := new(buildapiv1.Build)

	hco, err := codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		t.Fatal(err)
	}
	if err := hco.Object(obj); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", obj)
}

func TestTemplate_b1(t *testing.T) {
	buf := &bytes.Buffer{}
	te := template.New("build template").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)

	te = template.Must(te.Parse(SourceBuildConfigTemplate["BuildJSON"]))
	if err := te.Execute(buf, BuildTemplateOption{
		ObjectMetaTemplateOption: ObjectMetaTemplateOption{
			Name:      "osospringbootapp",
			Namespace: "fake",
			Annotations: []utility.NameValueStringPair{
				{Name: "anno1name", Value: "anno1value"},
			},
			Labels: []utility.KeyValueStringPair{
				{Key: "label1key", Value: "label1value"},
			},
		},
		CommonSpecTemplateOption: commonspecTplOpt,
		ManuallyBuildMessage:     "Manually build image",
	}); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	obj := new(buildapiv1.Build)

	ccf := util.NewClientCmdFactory()
	mapper, _ := ccf.Object(false)
	kapi.RegisterRESTMapper(mapper)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.Build{})
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, &buildapiv1.Build{})
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), buf.Bytes(), obj); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", obj)
}

func TestTemplate_bc(t *testing.T) {
	buf := &bytes.Buffer{}
	te := template.New("buildconfig template").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)

	te = template.Must(te.Parse(bcTmpl))
	if err := te.Execute(buf, bcTmplOpt); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	obj := new(buildapiv1.BuildConfig)

	/*hco, err := codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		t.Fatal(err)
	}
	if err := hco.Object(obj); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", obj)

	obj = new(buildapiv1.BuildConfig)*/

	ccf := util.NewClientCmdFactory()
	mapper, _ := ccf.Object(false)
	kapi.RegisterRESTMapper(mapper)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, &buildapiv1.BuildConfig{})
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), buf.Bytes(), obj); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", obj)
}
