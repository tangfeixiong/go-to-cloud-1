package builder

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

	fake_sourcebuilds_commontplopt = []FakeBuildTemplateOption{
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

	fake_sourcebuilds_bctplopt []string = []FakeBuildConfigTemplateOption{
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

	fake_sourcebuild_bcTemplate []string = []string{`
      {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}"
         },
         "spec": {
            "triggers": [
               {
                  "type": "GitHub",
                  "github": {
                     "secret": "{{.GithubTriggerSecret}}"
                  }
               },
               {
                  "type": "Generic",
                  "generic": {
                     "secret": "{{.GenericTriggerSecret}}"
                  }
               },
               {
                  "type": "ImageChange",
                  "imageChange": {}
               }
            ],
            "source": {
               "type": "Git",
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}"
               },
               "contextDir": "{{.ContextDir}}"
            },
            "strategy": {
               "type": "Source",
               "sourceStrategy": {
                  "from": {
                     "kind": "{{.SourceStrategyFromKind}}",
                     "name": "{{.SourceStrategyFromName}}"
                  }
               }
            },
            "output": {
               "to": {
                  "kind": "{{.OutputToKind}}",
                  "name": "{{.OutputToRegistry}}/{{.Name}}:{{.OutputToTag}}"
               }
            },
            "resources": {}
         }
      }`, `
      {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}"
         },
         "spec": {
            "triggers": [
               {
                  "type": "GitHub",
                  "github": {
                     "secret": "{{.GithubTriggerSecret}}"
                  }
               },
               {
                  "type": "Generic",
                  "generic": {
                     "secret": "{{.GenericTriggerSecret}}"
                  }
               },
               {
                  "type": "ImageChange",
                  "imageChange": {}
               }
            ],
            "source": {
               "type": "Git",
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}"
               },
               "contextDir": "{{.ContextDir}}"
            },
            "strategy": {
               "type": "Source",
               "sourceStrategy": {
                  "from": {
                     "kind": "{{.SourceStrategyFromKind}}",
                     "name": "{{.SourceStrategyFromName}}"
                  }
               }
            },
            "output": {
               "to": {
                  "kind": "{{.OutputToKind}}",
                  "name": "{{.OutputToRegistry}}/{{.Name}}:{{.OutputToTag}}"
               }
            },
            "resources": {}
         }
      }`,
	}

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
