package service

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/docker/engine-api/types"
	"github.com/golang/glog"
	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/runtime"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/cicd/pb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/build-builder"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

var (
	fake_req *pb3.TemplatizedBuilderRequest = &pb3.TemplatizedBuilderRequest{
		Feed: &pb3.Feed_FeedSpec{
			Builder:  pb3.Feed_FeedSpec_OPENSHIFT_ORIGIN_V3,
			Template: (*pb3.Template)(nil),
			Option: &pb3.BuildConfigOption{
				Name:    "osobuild",
				Project: "default",
				Annotations: map[string]string{
					"foo": "bar",
				},
				Labels: map[string]string{
					"foo": "bar",
				},
				//triggers
				GithubWebHook: []*pb3.GitHubWebHookTrigger{
					{Auth: &pb3.IdentifiedAuth{Id: "foo"}},
				},
				GenericWebHook: []*pb3.GenericWebHookTrigger{
					{Auth: &pb3.IdentifiedAuth{Id: "bar"},
						AllowEnv: true},
				},
				ImageChangeHook: []*pb3.ImageChangeHookTrigger{
					{Kind: pb3.ImageKindType_ImageStreamTag.String(), Name: "is1"},
				},
				ConfigChangeHook: false,
				RunPolicy:        pb3.BuildPolicyType_Serial.String(),
				//common spec - source
				SourceType:   pb3.BuildSourceType_Git.String(),
				BinaryAsFile: "",
				Dockerfile:   `FROM busybox\nCMD [\"/bin/sh\"]`,
				GitSource: &pb3.GitSource{
					Uri:  "https://github.com/tangfeixiong/osev3-examples",
					Ref:  "master",
					Path: "/spring-boot/sample-microservices-springboot/web",
				},
				SidecarImageSource: []*pb3.ImageSource{
					{
						Kind: pb3.ImageKindType_DockerImage.String(),
						Name: "busybox:latest",
						Paths: []*pb3.ImagePathMappingDir{
							{SourcePath: "/go", DestinationDir: "/go"},
						},
						RegistryAuth: &pb3.IdentifiedAuth{Id: "docker-auth"},
					},
					{
						Kind: pb3.ImageKindType_ImageStreamTag.String(),
						Name: "busybox:latest",
						Paths: []*pb3.ImagePathMappingDir{
							{SourcePath: "/go", DestinationDir: "/go"},
						},
					},
				},
				ContextDir:     "",
				RepositoryAuth: &pb3.IdentifiedAuth{Id: "git-auth"},
				AuthAsBuildSource: []*pb3.BuildConfigOption_AuthAsBuildSource{
					{Auth: &pb3.IdentifiedAuth{Id: "docker-auth"},
						DestinationDir: "/mnt"},
				},
				SourceRevisionType:  pb3.SourceRevisionType_Source_rev.String(),
				GitSourceRevision:   (*pb3.GitSourceRevision)(nil),
				BuildStrategyType:   pb3.BuildStrategyType_Source.String(),
				CustomBuildStrategy: (*pb3.CustomBuildStrategy)(nil),
				DockerBuildStrategy: &pb3.DockerBuildStrategy{
					ImageKind:    pb3.ImageKindType_DockerImage.String(),
					ImageName:    "busybox:latest",
					RegistryAuth: &pb3.IdentifiedAuth{Id: "docker-auth"},
					NoCache:      true,
					Env: map[string]string{
						"foo": "bar",
					},
					ForcePull:      false,
					DockerfilePath: "",
				},
				SourceBuildStrategy: &pb3.SourceBuildStrategy{
					ImageKind:    pb3.ImageKindType_ImageStreamTag.String(),
					ImageName:    "busybox:latest",
					RegistryAuth: &pb3.IdentifiedAuth{Id: "docker-auth"},
					Env: map[string]string{
						"foo": "bar",
					},
					Scripts:          "",
					Incremental:      false,
					ForcePull:        true,
					RuntimeImageKind: "",
					RuntimeImageName: "",
					RuntimeArtifacts: []*pb3.ImagePathMappingDir{},
				},
				JenkinsPipelineStrategy: &pb3.JenkinsPipelineStrategy{
					JenkinsfilePath: "",
					Jenkinsfile:     `node ('linux'){\n  stage 'Build and Test'\n  env.PATH = \"${tool 'Maven 3'}/bin:${env.PATH}\"\n  checkout scm\n  sh 'mvn clean package'\n}`,
				},
				ImageKind:                 pb3.ImageKindType_DockerImage.String(),
				ImageName:                 "foo:latest",
				RegistryAuth:              &pb3.IdentifiedAuth{Id: "docker-auth"},
				ResourceLimits:            map[string]string{},
				ResourceRequests:          map[string]string{},
				PostCommitCommand:         []string{},
				PostCommitArgs:            []string{},
				PostCommitScript:          "",
				CompletionDeadlineSeconds: 0,
			},
			Auth: []*pb3.IdentifiedAuth{
				{Id: "docker-auth",
					Auth:     "DockerAuthConfig",
					Server:   "127.0.0.1:5000",
					Username: "foo",
					Password: "bar"},
				{Id: "ssh-auth",
					Auth:           "ssh",
					SshAuthPrivate: "-----RSA-----"},
			},
			BuildAtOnceName:    "",
			ArchiveFilePath:    []string{},
			BuildAtOnceMessage: "Manually build",
			DisableBuildAtOnce: true,
		},

		Archive: (*pb3.Stream_StreamSpec)(nil),
	}

	fake_req_s2i []*pb3.TemplatizedBuilderRequest = []*pb3.TemplatizedBuilderRequest{
		{
			Feed: &pb3.Feed_FeedSpec{
				Builder:  pb3.Feed_FeedSpec_OPENSHIFT_ORIGIN_V3,
				Template: (*pb3.Template)(nil),
				Option: &pb3.BuildConfigOption{
					Name:    "springbootms-web",
					Project: "default",
					//triggers
					GithubWebHook:    []*pb3.GitHubWebHookTrigger{},
					GenericWebHook:   []*pb3.GenericWebHookTrigger{},
					ImageChangeHook:  []*pb3.ImageChangeHookTrigger{},
					ConfigChangeHook: false,
					RunPolicy:        pb3.BuildPolicyType_Serial.String(),
					//common spec - source
					SourceType:   pb3.BuildSourceType_Git.String(),
					BinaryAsFile: "",
					Dockerfile:   "",
					GitSource: &pb3.GitSource{
						Uri:  "https://github.com/tangfeixiong/osev3-examples",
						Ref:  "master",
						Path: "/spring-boot/sample-microservices-springboot/web",
					},
					SidecarImageSource:  []*pb3.ImageSource{},
					ContextDir:          "/spring-boot/sample-microservices-springboot/web",
					RepositoryAuth:      (*pb3.IdentifiedAuth)(nil),
					AuthAsBuildSource:   []*pb3.BuildConfigOption_AuthAsBuildSource{},
					SourceRevisionType:  pb3.SourceRevisionType_Source_rev.String(),
					GitSourceRevision:   (*pb3.GitSourceRevision)(nil),
					BuildStrategyType:   pb3.BuildStrategyType_Source.String(),
					CustomBuildStrategy: (*pb3.CustomBuildStrategy)(nil),
					DockerBuildStrategy: (*pb3.DockerBuildStrategy)(nil),
					SourceBuildStrategy: &pb3.SourceBuildStrategy{
						ImageKind:    pb3.ImageKindType_DockerImage.String(),
						ImageName:    "tangfeixiong/springboot-sti:gitcommit-1125149-0901T2236",
						RegistryAuth: &pb3.IdentifiedAuth{Id: "tangfx-dockerconfigjson"},
						ForcePull:    false,
					},
					ImageKind:                 pb3.ImageKindType_DockerImage.String(),
					ImageName:                 "172.17.4.50:30005/tangfx/osospringbootapp:latest",
					RegistryAuth:              &pb3.IdentifiedAuth{Id: "172-17-4-50-colon-30005-slash-tangfx"},
					ResourceLimits:            map[string]string{},
					ResourceRequests:          map[string]string{},
					PostCommitCommand:         []string{},
					PostCommitArgs:            []string{},
					PostCommitScript:          "",
					CompletionDeadlineSeconds: 0,
				},
				Auth: []*pb3.IdentifiedAuth{
					{Id: "172-17-4-50-colon-30005-slash-tangfx",
						Auth:     "DockerAuthConfig",
						Server:   "172.17.4.50:30005",
						Username: "tangfx",
						Password: "tangfx"},
				},
				BuildAtOnceName:    "springbootms-web",
				ArchiveFilePath:    []string{},
				BuildAtOnceMessage: "Manually build",
				DisableBuildAtOnce: false,
			},

			Archive: (*pb3.Stream_StreamSpec)(nil),
		},
		{
			Feed: &pb3.Feed_FeedSpec{
				Builder:  pb3.Feed_FeedSpec_OPENSHIFT_ORIGIN_V3,
				Template: (*pb3.Template)(nil),
				Option: &pb3.BuildConfigOption{
					Name:    "springbootms-data",
					Project: "default",
					//triggers
					GithubWebHook:    []*pb3.GitHubWebHookTrigger{},
					GenericWebHook:   []*pb3.GenericWebHookTrigger{},
					ImageChangeHook:  []*pb3.ImageChangeHookTrigger{},
					ConfigChangeHook: false,
					RunPolicy:        pb3.BuildPolicyType_Serial.String(),
					//common spec - source
					SourceType:   pb3.BuildSourceType_Git.String(),
					BinaryAsFile: "",
					Dockerfile:   "",
					GitSource: &pb3.GitSource{
						Uri:  "https://github.com/tangfeixiong/osev3-examples",
						Ref:  "master",
						Path: "/spring-boot/sample-microservices-springboot/web",
					},
					SidecarImageSource:  []*pb3.ImageSource{},
					ContextDir:          "/spring-boot/sample-microservices-springboot/web",
					RepositoryAuth:      (*pb3.IdentifiedAuth)(nil),
					AuthAsBuildSource:   []*pb3.BuildConfigOption_AuthAsBuildSource{},
					SourceRevisionType:  pb3.SourceRevisionType_Source_rev.String(),
					GitSourceRevision:   (*pb3.GitSourceRevision)(nil),
					BuildStrategyType:   pb3.BuildStrategyType_Source.String(),
					CustomBuildStrategy: (*pb3.CustomBuildStrategy)(nil),
					DockerBuildStrategy: (*pb3.DockerBuildStrategy)(nil),
					SourceBuildStrategy: &pb3.SourceBuildStrategy{
						ImageKind:    pb3.ImageKindType_DockerImage.String(),
						ImageName:    "tangfeixiong/springboot-sti:gitcommit-1125149-0901T2236",
						RegistryAuth: &pb3.IdentifiedAuth{Id: "tangfx-dockerconfigjson"},
						ForcePull:    false,
					},
					ImageKind:                 pb3.ImageKindType_DockerImage.String(),
					ImageName:                 "172.17.4.50:30005/tangfx/springbootms-data:latest",
					RegistryAuth:              &pb3.IdentifiedAuth{Id: "172-17-4-50-colon-30005-slash-tangfx"},
					ResourceLimits:            map[string]string{},
					ResourceRequests:          map[string]string{},
					PostCommitCommand:         []string{},
					PostCommitArgs:            []string{},
					PostCommitScript:          "",
					CompletionDeadlineSeconds: 0,
				},
				Auth: []*pb3.IdentifiedAuth{
					{Id: "172-17-4-50-colon-30005-slash-tangfx",
						Auth:     "DockerAuthConfig",
						Server:   "172.17.4.50:30005",
						Username: "tangfx",
						Password: "tangfx"},
				},
				BuildAtOnceName:    "springbootms-data",
				ArchiveFilePath:    []string{},
				BuildAtOnceMessage: "Manually build",
				DisableBuildAtOnce: false,
			},

			Archive: (*pb3.Stream_StreamSpec)(nil),
		},
	}
)

/*
  GOPATH=/work:/go:/data go test -v -run=Template_all ./pkg/service --args --loglevel=5
*/
func TestTemplate_action(t *testing.T) {
	if err := validateFeedSpec(fake_req_s2i[0].Feed); err != nil {
		t.Fatal(err)
	}
	for i := range fake_req_s2i[0].Feed.Auth {
		if err := validateAuth(fake_req_s2i[0].Feed.Auth[i]); err != nil {
			t.Fatal(err)
		}
	}
	t.Log(fake_req_s2i[0].Feed)
	var feed *pb3.Feed = &pb3.Feed{
		Metadata: &pb3.ObjectMeta{
			Name:        fake_req_s2i[0].Feed.Option.Name,
			Namespace:   fake_req_s2i[0].Feed.Option.Project,
			Annotations: make(map[string]string),
			Labels:      make(map[string]string),
		},
		Spec:   fake_req_s2i[0].Feed,
		Status: &pb3.Feed_FeedStatus{},
	}
	if len(fake_req_s2i[0].Feed.Option.Annotations) != 0 {
		feed.Metadata.Annotations = fake_req_s2i[0].Feed.Option.Annotations
	}
	if len(fake_req_s2i[0].Feed.Option.Labels) != 0 {
		feed.Metadata.Labels = fake_req_s2i[0].Feed.Option.Labels
	}

	opt, err := generateOption(fake_req_s2i[0], feed)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(opt)

	buf := &bytes.Buffer{}
	te := template.New("conf-tpl-exec").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)
	if fake_req_s2i[0].Feed.Template == nil || fake_req_s2i[0].Feed.Template.Metadata == nil || len(fake_req_s2i[0].Feed.Template.Metadata.Name) == 0 {
		te = template.Must(te.Parse(builder.SourceBuildConfigTemplate["buildconfig.json.tpl"]))
	} else {
		t.Fatal("service not ready")
	}
	if err := te.Execute(buf, opt); err != nil {
		t.Fatal(err)
	}
	t.Logf("Template executed: %+v", buf.String())

	obj := new(buildapiv1.BuildConfig)

	ccf := util.NewClientCmdFactory()
	mapper, _ := ccf.Object(false)
	kapi.RegisterRESTMapper(mapper)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, obj)
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), buf.Bytes(), obj); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", obj)

	ci := origin.NewPaaS()
	if err := ci.RequestProjectCreation(fake_req_s2i[0].Feed.Option.Project); err != nil {
		t.Fatal(err)
	}
	for i := range fake_req_s2i[0].Feed.Auth {
		switch strings.ToLower(fake_req_s2i[0].Feed.Auth[i].Auth) {
		case "dockerauthconfig":
			if err := ci.RequestBuilderSecretCreationWithDockerRegistry(fake_req_s2i[0].Feed.Option.Project, fake_req_s2i[0].Feed.Auth[i].Id, "builder", types.AuthConfig{
				Username:      fake_req_s2i[0].Feed.Auth[i].Username,
				Password:      fake_req_s2i[0].Feed.Auth[i].Password,
				ServerAddress: fake_req_s2i[0].Feed.Auth[i].Server}); err != nil {
				t.Fatal(err)
			}
		default:
			t.Fatal("%+v: %+v", errNotImplemented, fake_req_s2i[0].Feed.Auth[i])
		}
	}
	bcdata, bc, err := ci.RequestBuildConfigCreation(buf.Bytes())
	if err != nil {
		glog.V(9).Infof("Openshift API called: %+v", string(bcdata))
		t.Fatal(err)
	}
	t.Log(bc)
	// Build at once...
	//
	if fake_req_s2i[0].Feed.DisableBuildAtOnce || len(fake_req_s2i[0].Feed.ArchiveFilePath) != 0 {
		// write into response
		feed.Metadata.Annotations[vendor_create_annotation_key] = buf.String()
		feed.Status.Phase = fmt.Sprintf("LastVersion=%d", bc.Status.LastVersion)
		t.Log(feed.Status.Phase)
	} else {
		buildname := fake_req_s2i[0].Feed.BuildAtOnceName
		buildmessage := fake_req_s2i[0].Feed.BuildAtOnceMessage
		if len(buildname) == 0 {
			buildname = bc.Name
		}
		if len(buildmessage) == 0 {
			buildmessage = "Manually triggered"
		}
		bdata, obj, _, err := ci.RequestBuildCreation(buildname, buildmessage, bc)
		if err != nil {
			t.Fatal(err)
		}
		glog.Infof("Openshift API called: %+v", string(bdata))
		// write into response
		feed.Metadata.Annotations[vendor_create_annotation_key] = buf.String()
		feed.Status.Phase = string(obj.Status.Phase)

		ci.WaitForComplete = true
		ci.Follow = true
		//reader, writer := io.Pipe()
		//buf.Reset()
		//reader := buf
		//writer := buf
		writer := os.Stdout
		ci.Out = writer
		ci.ErrOut = writer
		ech := make(chan error)
		go func(ch chan<- error) {
			//defer writer.Close()
			if err := ci.StreamBuildLog(bdata, nil); err != nil {
				fmt.Fprintln(os.Stderr, "reading output:", err)
				ch <- err
			} else {
				ch <- nil
			}
			return
		}(ech)

		/*defer reader.Close()
		scanner := bufio.NewReader(reader)
		i := 0
		buffer := &bytes.Buffer{}
		for {
			line, isprefix, err := scanner.ReadLine()
			if err != nil && err != io.EOF {
				<-ech
				t.Fatal(err)
			}

			buffer.Write(line)
			if !isprefix {
				buffer.WriteByte('\n')
			}
			//fmt.Println(scanner.Text()) // Println will add back the final '\n'
			if err == io.EOF {
				feed.Status.Message = buffer.String()
				fmt.Print(os.Stdout, buffer.String())
				break
			}
			i += 1
			if i == 50 {
				i = 0
				feed.Status.Message = buffer.String()
				fmt.Print(os.Stdout, buffer.String())
				buffer.Reset()
			}
		}*/
		/*if err := scanner.Err(); err != nil {
			feed.Status.Message = buffer.String()
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			<-ech
			t.Fatal(err)
		}*/

		if err := <-ech; err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second * 1)
		temp := bufio.NewWriter(os.Stdout)
		temp.WriteString(feed.Status.Message)
		temp.Flush()
		io.Copy(os.Stdout, buf)
		t.Log(feed.Status.Phase)
	}
}

func TestTemplate_opt(t *testing.T) {
	var feed *pb3.Feed = &pb3.Feed{
		Metadata: &pb3.ObjectMeta{
			Name:        fake_req.Feed.Option.Name,
			Namespace:   fake_req.Feed.Option.Project,
			Annotations: fake_req.Feed.Option.Annotations,
			Labels:      fake_req.Feed.Option.Labels,
		},
		Spec: fake_req.Feed,
	}

	opt, err := generateOption(fake_req, feed)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(opt)
}

func TestTemplate_tpl(t *testing.T) {
	var feed *pb3.Feed = &pb3.Feed{
		Metadata: &pb3.ObjectMeta{
			Name:        fake_req.Feed.Option.Name,
			Namespace:   fake_req.Feed.Option.Project,
			Annotations: fake_req.Feed.Option.Annotations,
			Labels:      fake_req.Feed.Option.Labels,
		},
		Spec: fake_req.Feed,
	}

	opt, err := generateOption(fake_req, feed)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(opt)

	buf := &bytes.Buffer{}
	te := template.New("conf-tpl-exec").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)
	if fake_req.Feed.Template == nil || fake_req.Feed.Template.Metadata == nil || len(fake_req.Feed.Template.Metadata.Name) == 0 {
		te = template.Must(te.Parse(builder.SourceBuildConfigTemplate["buildconfig.json.tpl"]))
	} else {
		t.Fatal("service not ready")
	}
	if err := te.Execute(buf, opt); err != nil {
		t.Fatal(err)
	}
	t.Logf("Template executed: %+v", buf.String())
}

func TestTemplate_bc(t *testing.T) {
	if err := validateFeedSpec(fake_req.Feed); err != nil {
		t.Fatal(err)
	}
	for i := range fake_req.Feed.Auth {
		if err := validateAuth(fake_req.Feed.Auth[i]); err != nil {
			t.Fatal(err)
		}
	}
	var feed *pb3.Feed = &pb3.Feed{
		Metadata: &pb3.ObjectMeta{
			Name:        fake_req.Feed.Option.Name,
			Namespace:   fake_req.Feed.Option.Project,
			Annotations: fake_req.Feed.Option.Annotations,
			Labels:      fake_req.Feed.Option.Labels,
		},
		Spec: fake_req.Feed,
	}

	opt, err := generateOption(fake_req, feed)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", opt)

	buf := &bytes.Buffer{}
	te := template.New("conf-tpl-exec").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)
	if fake_req.Feed.Template == nil || fake_req.Feed.Template.Metadata == nil || len(fake_req.Feed.Template.Metadata.Name) == 0 {
		te = template.Must(te.Parse(builder.SourceBuildConfigTemplate["buildconfig.json.tpl"]))
	} else {
		t.Fatal("service not ready")
	}
	if err := te.Execute(buf, opt); err != nil {
		t.Fatal(err)
	}
	t.Logf("Template executed: %+v", buf.String())

	obj := new(buildapiv1.BuildConfig)

	ccf := util.NewClientCmdFactory()
	mapper, _ := ccf.Object(false)
	kapi.RegisterRESTMapper(mapper)
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, obj)
	if err := runtime.DecodeInto(kapi.Codecs.UniversalDeserializer(), buf.Bytes(), obj); err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", obj)
}
