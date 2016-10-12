package builder

import (
	"bytes"
	"os"
	"testing"
	"text/template"

	"github.com/helm/helm-classic/codec"
	buildapi "github.com/openshift/origin/pkg/build/api/v1"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/cmd-util"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

type FakeCommonBuildTemplateOption struct {
	Name                   string
	GitURI                 string
	GitRef                 string
	ContextDir             string
	SourceStrategyFromKind string
	SourceStrategyFromName string
	OutputToKind           string
	OutputToRegistry       string
	OutputToRepo           string
	OutputToTag            string
	OutputToName           string
}

type FakeBuildTemplateOption struct {
	FakeCommonBuildTemplateOption
}

type FakeBuildConfigTemplateOption struct {
	GithubTriggerSecret  string
	GenericTriggerSecret string
	FakeCommonBuildTemplateOption
}

var (
	fake_sourceprojectname = "tangfx"

	fake_sourcebuilds_commontplopt = []FakeCommonBuildTemplateOption{
		{
			Name:                   "springbootms-web",
			GitURI:                 "https://github.com/tangfeixiong/osev3-examples",
			GitRef:                 "master",
			ContextDir:             "/spring-boot/sample-microservices-springboot/web",
			SourceStrategyFromKind: "ImageStreamTag",
			SourceStrategyFromName: "tangfeixiong/springboot-sti:gitcommit1125149-0901T2236",
			OutputToKind:           "DockerImage",
			OutputToRegistry:       "172.17.4.50:30005/tangfx",
			OutputToRepo:           "",
			OutputToTag:            "latest",
			OutputToName:           "",
		},
		{
			Name:                   "springbootms-data",
			GitURI:                 "https://github.com/tangfeixiong/osev3-examples",
			GitRef:                 "master",
			ContextDir:             "/spring-boot/sample-microservices-springboot/repositories-mem",
			SourceStrategyFromKind: "ImageStreamTag",
			SourceStrategyFromName: "tangfeixiong/springboot-sti:gitcommit1125149-0901T2236",
			OutputToKind:           "DockerImage",
			OutputToRegistry:       "172.17.4.50:30005/tangfx",
			OutputToRepo:           "",
			OutputToTag:            "latest",
			OutputToName:           "",
		},
	}

	fake_sourcebuilds_bctplopt []FakeBuildConfigTemplateOption = []FakeBuildConfigTemplateOption{
		{
			GithubTriggerSecret:           "",
			GenericTriggerSecret:          "",
			FakeCommonBuildTemplateOption: fake_sourcebuilds_commontplopt[0],
		},
		{
			GithubTriggerSecret:           "",
			GenericTriggerSecret:          "",
			FakeCommonBuildTemplateOption: fake_sourcebuilds_commontplopt[1],
		},
	}

	fakeGitSecrets = map[string]string{"gogs": "tangfx:tangfx"}

	fake_sourcebuild_Dockerfile string = "FROM tangfeixiong/springboot-sti:gitcommit1125149-0901T2236\nRUN mvn install\nCMD [\"java\",\"-Djava.security.egd=file:/dev/./urandom\",\"-jar\",\"/opt/openshift/app.jar\"]"

	fake_sourcebuild_pullauth     string = ""
	fake_sourcebuild_pullusername string = ""
	fake_sourcebuild_pullpassword string = ""

	fake_sourcebuild_pushauth     string = ""
	fake_sourcebuild_pushusername string = "tangfx"
	fake_sourcebuild_pushpassword string = "tangfx"

	fake_sourcebuild_pullsecret string = ""
	fake_sourcebuild_pushsecret string = "dockerconfigjson-springbootms-web-to"

	fake_sourcebuilds_istemplate []string = []string{`
      {
         "kind": "ImageStream",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}"
         },
         "spec": {
            "dockerImageRepository": "",
            "tags": [
               {
                  "name": "latest"
               }
            ]
         }
      }`, `
      {
         "kind": "ImageStream",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}"
         },
         "spec": {
            "dockerImageRepository": "",
            "tags": [
               {
                  "name": "latest"
               }
            ]
         }
      }`,
	}
)

/*
  GOPATH=/work:/go:/data go test -v -run=Source ./pkg/appliance/openshift/origin/build-builder --args --loglevel=2
*/
func TestSource_One(t *testing.T) {
	tmpl := template.New("source build").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)
	tmpl = template.Must(tmpl.Parse(bTmpl))
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, bTmplOpt); err != nil {
		t.Fatal(err)
	}
	t.Log(buf.String())

	hco, err := codec.JSON.Decode(buf.Bytes()).One()
	if err != nil {
		t.Fatal(err)
	}
	bld := new(buildapi.Build)
	if err := hco.Object(bld); err != nil {
		t.Fatal(err)
	}

	ccf := util.NewClientCmdFactory()

	if err := RunS2IBuild(os.Stdout, bld, ccf); err != nil {
		t.Fatal(err)
	}
	t.Log(bld)

}
