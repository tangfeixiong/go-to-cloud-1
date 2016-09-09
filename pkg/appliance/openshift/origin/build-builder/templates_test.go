package builder

import (
	"os"
	"testing"
	"text/template"
)

var (
	commonspecTmplOpt = BuildCommonSpecTemplateOption{
		GitURI:     "https://github.com/tangfeixiong/osev3-examples",
		GitRef:     "master",
		ContextDir: "/spring-boot/sample-microservices-springboot/web",
		FromKind:   "DockerImage",
		FromName:   "tangfeixiong/springboot-sti:gitcommit-1125149-0901T2236",
		ForcePull:  false,
		ToKind:     "DockerImage",
		ToName:     "172.17.4.50:30005/tangfx/osospringbootapp",
	}

	bTmplOpt = BuildTemplateOption{
		Name:                          "osospringbootapp",
		Namespace:                     "tangfx",
		BuildCommonSpecTemplateOption: commonspecTmplOpt,
	}

	bTmpl = SourceBuildConfigTemplate["BuildForGitByManuallyTriggered"]

	bcTmplOpt = BuildConfigTemplateOption{
		Name:                          "osospringbootapp",
		Namespace:                     "tangfx",
		BuildCommonSpecTemplateOption: commonspecTmplOpt,
	}

	bcTmpl = SourceBuildConfigTemplate["BuildConfigForGitWithoutTriggers"]
)

func TestTemplate_One(t *testing.T) {
	te := template.New("build template")

	te = template.Must(te.Parse(bTmpl))
	if err := te.Execute(os.Stdout, bTmplOpt); err != nil {
		t.Fatal(err)
	}

	te = template.Must(te.Parse(bcTmpl))
	if err := te.Execute(os.Stdout, bcTmplOpt); err != nil {
		t.Fatal(err)
	}
}
