package builder

import (
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

type CredentialAuth struct {
	Username string
	Password string
	Auth     string
	Token    string
}

type ObjectMetaTemplateOption struct {
	Name        string
	Namespace   string
	Annotations []utility.NameValueStringPair
	Labels      []utility.KeyValueStringPair
}

type BinarySourceTemplateOption struct {
	AsFile     string
	ArchiveRaw []byte
	ContextDir string
}

type GitSourceTemplateOption struct {
	URI  string
	Ref  string
	Path string
}

type ImageSourcePath struct {
	SourcePath     string
	DestinationDir string
}

type AuthTemplateOption struct {
	Name string
	CredentialAuth
}

type SecretBuildSourceTemplateOption struct {
	AuthTemplateOption
	DestinationDir string
}

type ImageSourceTemplateOption struct {
	Kind     string
	Name     string
	Paths    []ImageSourcePath
	PullAuth AuthTemplateOption
}

type SourceRevisionTemplateOption struct {
	Type              string
	GitCommit         string
	GitAuthorName     string
	GitAuthorEmail    string
	GitCommitterName  string
	GitCommitterEmail string
	GitMessage        string
}

type BasicStrategyTemplateOption struct {
	ImageKind string
	ImageName string
	PullAuth  AuthTemplateOption
	Env       []utility.NameValueStringPair
	ForcePull bool
}

type SecretSpec struct {
	AuthTemplateOption
	MountPath string
}

type CustomStrategy struct {
	BasicStrategyTemplateOption
	ExposeDockerSocket bool
	Secrets            []SecretSpec
	BuildAPIVersion    string
}

type DockerStrategy struct {
	BasicStrategyTemplateOption
	NoCache        bool
	DockerfilePath string
}

type SourceStrategy struct {
	BasicStrategyTemplateOption
	Scripts          string
	Incremental      bool
	RuntimeImageKind string
	RuntimeImageName string
	RuntimeArtifacts []ImageSourcePath
}

type JenkinsPipelineStrategy struct {
	JenkinsfilePath string
	Jenkinsfile     string
}

type BuildStrategyTemplateOption struct {
	StrategyType            string
	DockerStrategy          DockerStrategy
	SourceStrategy          SourceStrategy
	CustomStrategy          CustomStrategy
	JenkinsPipelineStrategy JenkinsPipelineStrategy
}

type BuildOutputTemplateOption struct {
	ImageKind string
	ImageName string
	PushAuth  AuthTemplateOption
}

type ResourceTemplateOption struct {
	Name     string
	Quantity string
}

type CommonSpecTemplateOption struct {
	SourceType                string
	BinarySource              BinarySourceTemplateOption
	Dockerfile                string
	GitSource                 GitSourceTemplateOption
	ImageSource               []ImageSourceTemplateOption
	ContextDir                string
	SourceAuth                AuthTemplateOption
	SecretBuildSource         []SecretBuildSourceTemplateOption
	SourceRevision            SourceRevisionTemplateOption
	Strategy                  BuildStrategyTemplateOption
	Output                    BuildOutputTemplateOption
	ResourceLimits            []ResourceTemplateOption
	ResourceRequests          []ResourceTemplateOption
	PostCommitCommand         []string
	PostCommitArgs            []string
	PostCommitScript          string
	CompletionDeadlineSeconds int64
	SimpleGitOption
}

type SimpleGitOption struct {
	GitURI    string
	GitRef    string
	FromKind  string
	FromName  string
	ForcePull bool
	FromAuth  AuthTemplateOption
	ToKind    string
	ToName    string
	ToAuth    AuthTemplateOption
}

type BuildTemplateOption struct {
	ObjectMetaTemplateOption
	CommonSpecTemplateOption
	ManuallyBuildMessage string
}

type GitHubHookTemplateOption struct {
	WebHookTemplateOption
}

type WebHookTemplateOption struct {
	Secret string
	CredentialAuth
	AllowEnv bool
}

type ImageChangeHookTemplateOption struct {
	Kind string
	Name string
}

type ConfigChangeHookTemplateOption struct{}

type TriggerPolicyOption struct {
	GitHubWebHook    []GitHubHookTemplateOption
	GenericWebHook   []WebHookTemplateOption
	ImageChangeHook  []ImageChangeHookTemplateOption
	ConfigChangeHook ConfigChangeHookTemplateOption
}

type BuildConfigTemplateOption struct {
	ObjectMetaTemplateOption
	//TriggerPolicy *TriggerPolicyOption
	TriggerPolicy []interface{}
	RunPolicy     string
	CommonSpecTemplateOption
}

var (
	SourceBuildConfigTemplate map[string]string = map[string]string{
		"BuildConfigForGitWithoutTriggers": `
       {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            {{- if .Annotations }}{{ $length := len .Annotations }}
            "annotations": {
                {{- range $i, $e := $.Annotations }}
                  {{ quote $e.Name }}: {{ quote $e.Value }}{{ if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            {{- if .Labels }}{{ $length := len .Labels }}
            "labels": {
                {{- range $i, $e := $.Labels }}
                {{ quote $e.Key }}: {{ quote $e.Value }}{{- if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            "name": {{ quote .Name }},
            "namespace": {{ default "default" .Namespace | quote }}
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
            {{- if .Annotations }}{{ $length := len .Annotations }}
            "annotations": {
                {{- range $i, $e := $.Annotations }}
                  {{ quote $e.Name }}: {{ quote $e.Value }}{{ if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            {{- if .Labels }}{{ $length := len .Labels }}
            "labels": {
                {{- range $i, $e := $.Labels }}
                {{ quote $e.Key }}: {{ quote $e.Value }}{{- if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            "name": {{ quote .Name }},
            "namespace": {{ default "default" .Namespace | quote }}
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
		"buildconfig.json.tpl": `
       {
         "kind": "BuildConfig",
         "apiVersion": "v1",
         "metadata": {
            {{- if .Annotations }}{{ $length := len .Annotations }}
            "annotations": {
                {{- range $i, $e := $.Annotations }}
                  {{ quote $e.Name }}: {{ quote $e.Value }}{{ if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            {{- if .Labels }}{{ $length := len .Labels }}
            "labels": {
                {{- range $i, $e := $.Labels }}
                {{ quote $e.Key }}: {{ quote $e.Value }}{{- if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            "name": {{ quote .Name }},
            "namespace": {{ default "default" .Namespace | quote }}
         },
         "spec": {
            "triggers": {{ if .TriggerPolicy }}[
               {{- $length := len .TriggerPolicy -}}
               {{- range $idx, $ele := .TriggerPolicy }}
               {
                  {{- if typeIs "builder.GitHubHookTemplateOption" $ele }}
                  "type": "GitHub",
                  "github": {
                     "secret": {{ quote $ele.Secret }}
                  }
                  {{- else if typeIs "builder.WebHookTemplateOption" $ele }}
                  "type": "Generic",
                  "generic": {
                     "secret": {{ quote $ele.Secret }},
                     "allowEnv": true
                  }
                  {{- else if typeIs "builder.ImageChangeHookTemplateOption" $ele }}
                  "type": "ImageChange",
                  "imageChange": {{ if and $ele.Kind $ele.Name }}{
                      "from": {
                          "kind": {{ quote $ele.Kind }},
                          "name": {{ quote $ele.Name }}
                      }
                  }{{ else }}null{{ end }}
                  {{- else if typeIs "builder.ConfigChangeHookTemplateOption" $ele }}
                  "type": "ConfigChange"
                  {{- end }}
               }{{- if lt (plus1 $idx) $length -}},{{- end }}
               {{- end }}
            ]{{ else -}}[]{{- end }},
            "runPolicy": {{default "Serial" .RunPolicy | quote}},
            "serviceAccount": "builder",
            "source": {
               "type": {{default "Dockerfile" .SourceType | quote}},               
               "binary": {{ if eq "Binary" .SourceType -}}{
                   "asFile": {{ default "" .BinarySource.AsFile | quote }} 
               }{{ else -}}null{{- end }},
               "dockerfile": {{if .Dockerfile}}{{quote .Dockerfile}}{{else}}null{{end}},
               "git": {{ if .GitSource.URI -}}{
                  "uri": {{quote .GitSource.URI}},
                  "ref": {{default "" .GitSource.Ref | quote}},
                  "httpProxy": null,
                  "httpsProxy": null
               }{{ else -}}null{{- end }},               
               "images": {{ if .ImageSource -}}[
                  {{- range $idx, $ele := .ImageSource }}
                  {
                     "from": {
                        "kind": {{ quote $ele.Kind }},
                        "name": {{ quote $ele.Name }}
                     },
                     "paths": [
                        {{ range $i, $e := $ele.Paths -}}
                        {
                           "sourcePath": {{ quote $e.SourcePath }},
                           "destinationDir": {{ quote $e.DestinationDir }}
                        }{{- if not (last $i $ele.Paths) -}},{{- end }}
                        {{- end }}
                     ],
                     "pullSecret": {{if $ele.PullAuth.Name}}{
                        "name": {{quote $ele.PullAuth.Name}}
                     }{{else}}null{{end}}
                  }{{- if not (last $idx $.ImageSource) -}},{{- end }}
                  {{- end }}
               ]{{ else -}}[]{{- end -}},
               "contextDir": {{if eq "Git" .SourceType -}}{{- default "" .GitSource.Path | quote -}}
                  {{- else if eq "Binary" .SourceType -}}{{- default "" .BinarySource.ContextDir | quote -}}
                  {{- else -}}{{- default "" .ContextDir | quote }}{{- end }},
               "sourceSecret": {{if .SourceAuth.Name -}}{
                  "name": {{quote .SourceAuth.Name}}
               }{{else}}null{{end}},
               "secrets": {{ if .SecretBuildSource -}}[
                  {{- range $idx, $ele := .SecretBuildSource }}
                  {
                     "secret": {
                        "name": {{quote $ele.Name}}
                     },
                     "destinationDir": {{quote $ele.DestinationDir}}
                  }{{- if not (last $idx $.SecretBuildSource) -}},{{- end }}
                  {{- end }}
               ]{{ else -}}[]{{- end }}
            },
            "revision": {{if .SourceRevision.GitCommit -}}{
               "type": {{quote .SourceRevision.Type}},
               "git": {
                  "commit": {{quote .SourceRevision.GitCommit}},
                  "author": {
                     "name": {{quote .SourceRevision.GitAuthorName}},
                     "email": {{quote .SourceRevision.GitAuthorEmail}}
                  },
                  "committer": {
                     "name": {{quote .SourceRevision.GitCommitterName}},
                     "email": {{quote .SourceRevision.GitCommitterEmail}}
                  },
                  "message": {{quote .SourceRevision.GitMessage}}
               }
            }{{else}}null{{end}},
            "strategy": {
               "type": {{quote .Strategy.StrategyType}},
               "sourceStrategy": {{if and .Strategy.SourceStrategy.ImageKind .Strategy.SourceStrategy.ImageName}}{
                  "from": {
                     "kind": {{quote .Strategy.SourceStrategy.ImageKind}},
                     "name": {{quote .Strategy.SourceStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.SourceStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.SourceStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "env": {{if .Strategy.SourceStrategy.Env}}[
                     {{ range $i, $e := .Strategy.SourceStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.SourceStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "scripts": {{default "" .Strategy.SourceStrategy.Scripts | quote}},
                  "incremental": {{.Strategy.SourceStrategy.Incremental}},
                  "forcePull": {{.Strategy.SourceStrategy.ForcePull}},
                  "runtimeImage": {{if .Strategy.SourceStrategy.RuntimeImageKind}}{
                      "kind": {{quote .Strategy.SourceStrategy.RuntimeImageKind}},
                      "name": {{quote .Strategy.SourceStrategy.RuntimeImageName}}
                  }{{else}}null{{end}},
                  "runtimeArtifacts": {{if .Strategy.SourceStrategy.RuntimeArtifacts}}[
                     {{ range $i, $e := .Strategy.SourceStrategy.RuntimeArtifacts -}}
                     {
                        "sourcePath": {{ quote $e.SourcePath }},
                        "destinationDir": {{ quote $e.DestinationDir }}
                     }{{- if not (last $i $.Strategy.SourceStrategy.RuntimeArtifacts) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}}
               }{{else}}null{{end}},
               "dockerStrategy": {{if and .Strategy.DockerStrategy.ImageKind .Strategy.DockerStrategy.ImageName}}{
                  "from": {
                     "kind": {{quote .Strategy.DockerStrategy.ImageKind}},
                     "name": {{quote .Strategy.DockerStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.DockerStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.DockerStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "noCache": {{.Strategy.DockerStrategy.NoCache}},
                  "env": {{if .Strategy.DockerStrategy.Env}}[
                     {{ range $i, $e := .Strategy.DockerStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.DockerStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "forcePull": {{.Strategy.DockerStrategy.ForcePull}},
                  "dockerfilePath": {{default "" .Strategy.DockerStrategy.DockerfilePath | quote}}
               }{{else}}null{{end}},
               "customStrategy": {{if and  .Strategy.CustomStrategy.ImageKind .Strategy.CustomStrategy.ImageName}}{
                  "from": {
                     "kind": {{quote .Strategy.CustomStrategy.ImageKind}},
                     "name": {{quote .Strategy.CustomStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.CustomStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.CustomStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "env": {{if .Strategy.CustomStrategy.Env}}[
                     {{ range $i, $e := .Strategy.CustomStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.CustomStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "exposeDockerSocket": {{.Strategy.CustomStrategy.ExposeDockerSocket}},
                  "forcePull": {{.Strategy.CustomStrategy.ForcePull}},
                  "secrets": {{ if .Strategy.CustomStrategy.Secrets -}}[
                     {{- range $idx, $ele := .Strategy.CustomStrategy.Secrets }}
                     {
                        "secretSource": {
                           "name": {{quote $ele.Name}}
                        },
                        "mountPath": {{quote $ele.MountPath}}
                     }{{- if not (last $idx $.Strategy.CustomStrategy.Secrets) -}},{{- end }}
                     {{- end }}
                  ]{{ else -}}[]{{- end -}},
                  "buildAPIVersion": {{default "" .Strategy.CustomStrategy.BuildAPIVersion | quote}}
               }{{else}}null{{end}},
               "jenkinsPipelineStrategy": {{if or .Strategy.JenkinsPipelineStrategy.JenkinsfilePath .Strategy.JenkinsPipelineStrategy.Jenkinsfile}}{
                  "jenkinsfilePath": {{default "" .Strategy.JenkinsPipelineStrategy.JenkinsfilePath | quote }},
                  "jenkinsfile": {{default "" .Strategy.JenkinsPipelineStrategy.Jenkinsfile | quote}}
               }{{else}}null{{end}}
            },
            "output": {
               "to": {
                  "kind": {{quote .Output.ImageKind}},
                  "name": {{quote .Output.ImageName}}
               },
               "pushSecret": {{if .Output.PushAuth.Name}}{
                   "name": {{quote .Output.PushAuth.Name}}
               }{{else}}null{{end}}
            },
            "resources": {{if and .ResourceLimits .ResourceRequests}}{
               "limits": {{if .ResourceLimits}}{
                  {{- range $idx, $ele := .ResourceLimits }}
                     {{quote $ele.Name}}: {{quote $ele.Quantity}}{{if not (last $idx $.ResourceLimits)}},{{end}}
                  {{- end}}
               }{{else}}{}{{end}},
               "requests": {{if .ResourceRequests}}{
                  {{- range $idx, $ele := .ResourceRequests }}
                     {{quote $ele.Name}}: {{quote $ele.Quantity}}{{if not (last $idx $.ResourceRequests)}},{{end}}
                  {{- end}}
               }{{else}}{}{{end}}
            }{{else}}{}{{end}},
            "postCommit": {{if and .PostCommitCommand .PostCommitArgs .PostCommitScript}}{
               "command": {{ if .PostCommitCommand -}}[
                   {{- range $idx, $ele := .PostCommitCommand }}
                   {{quote $ele}}{{if not (last $idx $.PostCommitCommand)}},{{end}}
                   {{- end}}            
               ]{{ else -}}[]{{- end -}}{{- if and .PostCommitArgs .PostCommitScript -}},{{- end }}
               "args": {{ if .PostCommitArgs -}}[
                   {{- range $idx, $ele := .PostCommitArgs }}
                   {{quote $ele}}{{if not (last $idx $.PostCommitArgs)}},{{end}}
                   {{- end}}            
               ]{{ else -}}[]{{- end -}}{{- if .PostCommitScript -}},
               "script": {{quote .PostCommitScript}}{{- end }}
            }{{else}}{}{{end}},
            "completionDeadlineSeconds": {{if .CompletionDeadlineSeconds -}}{{.CompletionDeadlineSeconds}}{{else}}null{{end}}
         }
       }`,
		"BuildConfig YAML": `---
        kind: BuildConfig,
        apiVersion: "v1,
          metadata:
            {{- if .Annotations }}
            annotations:
              {{- range $i, $e := $.Annotations }}
              {{ quote $e.Name }}: {{ quote $e.Value }}
              {{- end }}
            {{- end }}
            {{- if .Labels }}
            labels:
              {{- range $i, $e := $.Labels }}
              {{ quote $e.Key }}: {{ quote $e.Value }}
              {{- end }}
            {{- end }}
            name: {{ quote .Name }},
            namespace: {{ default "default" .Namespace | quote }}
          spec:
            {{- if .TriggerPolicy }}{{ $length := len .TriggerPolicy }}
            "triggers": [
               {{- range $idx, $ele := .TriggerPolicy }}
               {
                  {{- if typeIs "builder.GitHubHookTemplateOption" $ele }}
                  "type": "GitHub",
                  "github": {
                     "secret": {{ quote $ele.Secret }}
                  }
                  {{- else if typeIs "builder.WebHookTemplateOption" $ele }}
                  "type": "Generic",
                  "generic": {
                     "secret": {{ quote $ele.Secret }},
                     "allowEnv": true
                  }
                  {{- else if typeIs "builder.ImageChangeHookTemplateOption" $ele }}
                  "type": "ImageChange",
                  "imageChange": {{ if and $ele.Kind $ele.Name }}{
                      "from": {
                          "kind": {{ quote $ele.Kind }},
                          "name": {{ quote $ele.Name }}
                      }
                  }{{ else }}null{{ end }}
                  {{- else if typeIs "builder.ConfigChangeHookTemplateOption" $ele }}
                  "type": "ConfigChange"
                  {{- end }}
               }
               {{- if lt (plus1 $idx) $length -}},{{- end }}
               {{- end }}
            ],
            {{- end }}
            runPolicy: "Serial",
            serviceAccount: "builder",
            source:
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
            strategy:
               type: Source,
               sourceStrategy:
                  "from": {
                     "kind": "{{.FromKind}}",
                     "name": "{{.FromName}}"
                  },
                  "pullSecret": null,
                  "forcePull": {{.ForcePull}},
                  "runtimeImage": null,
                  "runtimeArtifacts": []
            output:
              to:
                kind": "{{.ToKind}}
                name": "{{.ToName}}
              pushSecret: null
            resources: {}
            postCommit: {}
            completionDeadlineSeconds: null
        `,
		"BuildJSON": `
       {
         "kind": "Build",
         "apiVersion": "v1",
         "metadata": {
            {{- if .Annotations }}{{ $length := len .Annotations }}
            "annotations": {
                {{- range $i, $e := $.Annotations }}
                  {{ quote $e.Name }}: {{ quote $e.Value }}{{ if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            {{- if .Labels }}{{ $length := len .Labels }}
            "labels": {
                {{- range $i, $e := $.Labels }}
                {{ quote $e.Key }}: {{ quote $e.Value }}{{- if lt (plus1 $i) $length -}},{{- end }}
                {{- end }}
            },
            {{- end }}
            "name": {{ quote .Name }},
            "namespace": {{ default "default" .Namespace | quote }}
         },
         "spec": {
            "serviceAccount": "builder",
            "source": {
               "type": {{default "Dockerfile" .SourceType | quote}},               
               "binary": {{ if eq "Binary" .SourceType -}}{
                   "asFile": {{ default "" .BinarySource.AsFile | quote }} 
               }{{ else -}}null{{- end }},
               "dockerfile": {{if .Dockerfile}}{{quote .Dockerfile}}{{else}}null{{end}},
               "git": {{ if .GitSource.URI -}}{
                  "uri": {{quote .GitSource.URI}},
                  "ref": {{default "" .GitSource.Ref | quote}},
                  "httpProxy": null,
                  "httpsProxy": null
               }{{ else -}}null{{- end }},               
               "images": {{ if .ImageSource -}}[
                  {{- range $idx, $ele := .ImageSource }}
                  {
                     "from": {
                        "kind": {{ quote $ele.Kind }},
                        "name": {{ quote $ele.Name }}
                     },
                     "paths": [
                        {{ range $i, $e := $ele.Paths -}}
                        {
                           "sourcePath": {{ quote $e.SourcePath }},
                           "destinationDir": {{ quote $e.DestinationDir }}
                        }{{- if not (last $i $ele.Paths) -}},{{- end }}
                        {{- end }}
                     ],
                     "pullSecret": {{if $ele.PullAuth.Name}}{
                        "name": {{quote $ele.PullAuth.Name}}
                     }{{else}}null{{end}}
                  }{{- if not (last $idx $.ImageSource) -}},{{- end }}
                  {{- end }}
               ]{{ else -}}[]{{- end -}},
               "contextDir": {{if eq "Git" .SourceType -}}{{- default "" .GitSource.Path | quote -}}
                  {{- else if eq "Binary" .SourceType -}}{{- default "" .BinarySource.ContextDir | quote -}}
                  {{- else -}}{{- default "" .ContextDir | quote }}{{- end }},
               "sourceSecret": {{if .SourceAuth.Name -}}{
                  "name": {{quote .SourceAuth.Name}}
               }{{else}}null{{end}},
               "secrets": {{ if .SecretBuildSource -}}[
                  {{- range $idx, $ele := .SecretBuildSource }}
                  {
                     "secret": {
                        "name": {{quote $ele.Name}}
                     },
                     "destinationDir": {{quote $ele.DestinationDir}}
                  }{{- if not (last $idx $.SecretBuildSource) -}},{{- end }}
                  {{- end }}
               ]{{ else -}}[]{{- end }}
            },
            "revision": {{if .SourceRevision.GitCommit -}}{
               "type": {{default "" .SourceRevision.Type | quote}},
               "git": {
                  "commit": {{quote .SourceRevision.GitCommit}},
                  "author": {
                     "name": {{default "" .SourceRevision.GitAuthorName | quote }},
                     "email": {{default "" .SourceRevision.GitAuthorEmail | quote }}
                  },
                  "committer": {
                     "name": {{default "" .SourceRevision.GitCommitterName | quote }},
                     "email": {{default "" .SourceRevision.GitCommitterEmail | quote }}
                  },
                  "message": {{default "" .SourceRevision.GitMessage | quote}}
               }
            }{{else}}null{{end}},
            "strategy": {
               "type": {{quote .Strategy.StrategyType}},
               "sourceStrategy": {{if .Strategy.SourceStrategy}}{
                  "from": {
                     "kind": {{quote .Strategy.SourceStrategy.ImageKind}},
                     "name": {{quote .Strategy.SourceStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.SourceStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.SourceStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "env": {{if .Strategy.SourceStrategy.Env}}[
                     {{ range $i, $e := .Strategy.SourceStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.SourceStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "scripts": {{default "" .Strategy.SourceStrategy.Scripts | quote}},
                  "incremental": {{.Strategy.SourceStrategy.Incremental}},
                  "forcePull": {{.Strategy.SourceStrategy.ForcePull}},
                  "runtimeImage": {{if .Strategy.SourceStrategy.RuntimeImageKind}}{
                      "kind": {{quote .Strategy.SourceStrategy.RuntimeImageKind}},
                      "name": {{quote .Strategy.SourceStrategy.RuntimeImageName}}
                  }{{else}}null{{end}},
                  "runtimeArtifacts": {{if .Strategy.SourceStrategy.RuntimeArtifacts}}[
                     {{ range $i, $e := .Strategy.SourceStrategy.RuntimeArtifacts -}}
                     {
                        "sourcePath": {{ quote $e.SourcePath }},
                        "destinationDir": {{ quote $e.DestinationDir }}
                     }{{- if not (last $i $.Strategy.SourceStrategy.RuntimeArtifacts) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}}
               }{{else}}null{{end}},
               "dockerStrategy": {{if .Strategy.DockerStrategy}}{
                  "from": {
                     "kind": {{quote .Strategy.DockerStrategy.ImageKind}},
                     "name": {{quote .Strategy.DockerStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.DockerStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.DockerStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "noCache": {{.Strategy.DockerStrategy.NoCache}},
                  "env": {{if .Strategy.DockerStrategy.Env}}[
                     {{ range $i, $e := .Strategy.DockerStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.DockerStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "forcePull": {{.Strategy.DockerStrategy.ForcePull}},
                  "dockerfilePath": {{default "" .Strategy.DockerStrategy.DockerfilePath | quote}}
               }{{else}}null{{end}},
               "customStrategy": {{if .Strategy.CustomStrategy}}{
                  "from": {
                     "kind": {{quote .Strategy.CustomStrategy.ImageKind}},
                     "name": {{quote .Strategy.CustomStrategy.ImageName}}
                  },
                  "pullSecret": {{if .Strategy.CustomStrategy.PullAuth.Name}}{
                      "name": {{quote .Strategy.CustomStrategy.PullAuth.Name}}
                  }{{else}}null{{end}},
                  "env": {{if .Strategy.CustomStrategy.Env}}[
                     {{ range $i, $e := .Strategy.CustomStrategy.Env -}}
                     {
                        "Name": {{ quote $e.Name }},
                        "Value": {{ quote $e.Value }}
                     }{{- if not (last $i $.Strategy.CustomStrategy.Env) -}},{{- end }}
                     {{- end }}
                  ]{{else}}[]{{end}},
                  "exposeDockerSocket": {{.Strategy.CustomStrategy.ExposeDockerSocket}},
                  "forcePull": {{.Strategy.CustomStrategy.ForcePull}},
                  "secrets": {{ if .Strategy.CustomStrategy.Secrets -}}[
                     {{- range $idx, $ele := .Strategy.CustomStrategy.Secrets }}
                     {
                        "secretSource": {
                           "name": {{quote $ele.Name}}
                        },
                        "mountPath": {{quote $ele.MountPath}}
                     }{{- if not (last $idx $.Strategy.CustomStrategy.Secrets) -}},{{- end }}
                     {{- end }}
                  ]{{ else -}}[]{{- end -}},
                  "buildAPIVersion": {{default "" .Strategy.CustomStrategy.BuildAPIVersion | quote}}
               }{{else}}null{{end}},
               "jenkinsPipelineStrategy": {{if .Strategy.JenkinsPipelineStrategy.Jenkinsfile}}{
                  "jenkinsfilePath": {{default "" .Strategy.JenkinsPipelineStrategy.JenkinsfilePath | quote }},
                  "jenkinsfile": {{quote .Strategy.JenkinsPipelineStrategy.Jenkinsfile}}
               }{{else}}null{{end}}
            },
            "output": {
               "to": {
                  "kind": {{quote .Output.ImageKind}},
                  "name": {{quote .Output.ImageName}}
               },
               "pushSecret": {{if .Output.PushAuth.Name}}{
                   "name": {{quote .Output.PushAuth.Name}}
               }{{else}}null{{end}}
            },
            "resources": {{if and .ResourceLimits .ResourceRequests}}{
               "limits": {{if .ResourceLimits}}{
                  {{- range $idx, $ele := .ResourceLimits }}
                     {{quote $ele.Name}}: {{quote $ele.Quantity}}{{if not (last $idx $.ResourceLimits)}},{{end}}
                  {{- end}}
               }{{else}}{}{{end}},
               "requests": {{if .ResourceRequests}}{
                  {{- range $idx, $ele := .ResourceRequests }}
                     {{quote $ele.Name}}: {{quote $ele.Quantity}}{{if not (last $idx $.ResourceRequests)}},{{end}}
                  {{- end}}
               }{{else}}{}{{end}}
            }{{else}}{}{{end}},
            "postCommit": {{if and .PostCommitCommand .PostCommitArgs .PostCommitScript}}{
               "command": {{ if .PostCommitCommand -}}[
                   {{- range $idx, $ele := .PostCommitCommand }}
                   {{quote $ele}}{{if not (last $idx $.PostCommitCommand)}},{{end}}
                   {{- end}}            
               ]{{ else -}}[]{{- end -}}{{- if and .PostCommitArgs .PostCommitScript -}},{{- end }}
               "args": {{ if .PostCommitArgs -}}[
                   {{- range $idx, $ele := .PostCommitArgs }}
                   {{quote $ele}}{{if not (last $idx $.PostCommitArgs)}},{{end}}
                   {{- end}}            
               ]{{ else -}}[]{{- end -}}{{- if .PostCommitScript -}},
               "script": {{quote .PostCommitScript}}{{- end }}
            }{{else}}{}{{end}},
            "completionDeadlineSeconds": {{if .CompletionDeadlineSeconds -}}{{.CompletionDeadlineSeconds}}{{else}}null{{end}},
            "triggeredBy": {{if .ManuallyBuildMessage }}[
               {
                  "message": {{quote .ManuallyBuildMessage}},
                  "githubWebHook": null,
                  "genericWebHook": null,
                  "imageChangeBuild": null
               }
            ]{{else}}[]{{end}}
         }
       }`,
	}
)
