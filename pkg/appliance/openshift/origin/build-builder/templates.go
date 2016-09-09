package builder

type BuildCommonSpecTemplateOption struct {
	GitURI     string
	GitRef     string
	ContextDir string
	FromKind   string
	FromName   string
	ForcePull  bool
	ToKind     string
	ToName     string
}

type BuildTemplateOption struct {
	Name      string
	Namespace string
	BuildCommonSpecTemplateOption
}

type BuildConfigTemplateOption struct {
	Name      string
	Namespace string
	BuildCommonSpecTemplateOption
}

var (
	SourceBuildConfigTemplate map[string]string = map[string]string{
		"BuildConfigForGitWithoutTriggers": `
      {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}",
            "namespace": "{{.Namespace}}"
         },
         "spec": {
            "triggers": [],
            "runPolicy": "Serial",
            "serviceAccount": "builder",
            "source": {
               "type": "Git",
               "binary": null,
               "dockerfile": null,
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}",
                  "httpProxy": null,
                  "httpsProxy": null
               },
               "images": [],
               "contextDir": "{{.ContextDir}}",
               "sourceSecret": null,
               "secrets": []
            },
            "strategy": {
               "type": "Source",
               "sourceStrategy": {
                  "from": {
                     "kind": "{{.FromKind}}",
                     "name": "{{.FromName}}"
                  },
                  "pullSecret": null,
                  "forcePull": {{.ForcePull}},
                  "runtimeImage": null,
                  "runtimeArtifacts": []
               }
            },
            "output": {
               "to": {
                  "kind": "{{.ToKind}}",
                  "name": "{{.ToName}}"
               },
               "pushSecret": null
            },
            "resources": {},
            "postCommit": {},
            "CompletionDeadlineSeconds": null
         }
      }`,
		"BuildForGitByManuallyTriggered": `
      {
         "kind": "Build",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}",
            "namespace": "{{.Namespace}}"
         },
         "spec": {
            "serviceAccount": "builder",
            "source": {
               "type": "Git",
               "binary": null,
               "dockerfile": null,
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}",
                  "httpProxy": null,
                  "httpsProxy": null
               },
               "images": [],
               "contextDir": "{{.ContextDir}}",
               "sourceSecret": null,
               "secrets": []
            },
            "revision": null,
            "strategy": {
               "type": "Source",
               "sourceStrategy": {
                  "from": {
                     "kind": "{{.FromKind}}",
                     "name": "{{.FromName}}"
                  },
                  "pullSecret": null,
                  "forcePull": {{.ForcePull}},
                  "runtimeImage": null,
                  "runtimeArtifacts": []
               }
            },
            "output": {
               "to": {
                  "kind": "{{.ToKind}}",
                  "name": "{{.ToName}}"
               },
               "pushSecret": null
            },
            "resources": {},
            "postCommit": {},
            "CompletionDeadlineSeconds": null,
            "triggeredBy": [
               {
                  "message": "Manally triggered"
               }
            ]
         }
      }`,
		"BuildConfig JSON model": `
      {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}",
            "namespace": "{{.Namespace}}"
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
            "runPolicy": "Serial",
            "source": {
               "type": "Git",
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}"
               },
               "contextDir": "{{.ContextDir}}"
            },
            "revision": null,
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
		"build JSON model": `
      {
         "kind": "Build",
         "apiVersion": "v1",
         "metadata": {
            "name": "{{.Name}}",
            "namespace": "{{.Namespace}}"
         },
         "spec": {
            "serviceAccount": "{{.BuilderServiceAccount}}",
            "source": {
               "type": "Git",
               "binary": null,
               "dockerfile": null,
               "git": {
                  "uri": "{{.GitURI}}",
                  "ref": "{{.GitRef}}"
               },
               "images": [],
               "contextDir": "{{.ContextDir}}",
               "sourceSecret": null,
               "secrets": []
            },
            "revision": null,
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
            "resources": {},
            "postCommit": {},
            "CompletionDeadlineSeconds": null,
            "triggeredBy": [
               {
                  "message": "{{.TriggerCauseMessage}}",
                  "gitHubWebHook" : {
                    "type": "GitHub",
                    "github": {
                     "secret": "{{.GithubTriggerSecret}}"
                    }
                  }
               },
               {
                  "message": "{{.TriggerCauseMessage}}",
                  "genericWebHook" : {
                    "type": "Generic",
                    "generic": {
                     "secret": "{{.GenericTriggerSecret}}"
                    }
                  }
               },
               {
                  "message": "{{.TriggerCauseMessage}}",
                  "imageChangeBuild": {
                    "type": "ImageChange",
                    "imageChange": {}
                  }
               }, 
               {
                  "message": "{{.TrggerCauseMessage}}"
               }
            ]
         }
      }`,
	}
)
