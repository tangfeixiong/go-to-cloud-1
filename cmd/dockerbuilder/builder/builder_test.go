
package builder

import (
    "bytes"
    _ "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "testing"

	"github.com/ghodss/yaml"
    "github.com/helm/helm/codec"    
    oapi "github.com/openshift/origin/pkg/build/api"
    
)

func TestExportBuild(t *testing.T) {
    
    o := oapi.Build { }
    o.TypeMeta.Kind = "Build"
    o.TypeMeta.APIVersion = "v1"
    
    o.ObjectMeta.Name = "build101"
    o.ObjectMeta.Labels = make(map[string]string)
    o.ObjectMeta.Labels["ci"] = "build101"
    o.ObjectMeta.Labels["name"] = "build101"
    
    fillBuildSpec(&o.Spec)
    
    b := new(bytes.Buffer)
    err := codec.JSON.Encode(b).One(o)
    if err != nil {
        t.Errorf("Could not encode JSON object: %s", err)
    }
    fmt.Println(b.String())
    
    y, err := yaml.Marshal(o)
    if err != nil {
        t.Errorf("Could not encode YAML object: %s", err)
    }
    fmt.Println(string(y))
    
    /*
    o, err := codec.JSON.Decode(b.Bytes()).One()
    if err != nil {
        t.Errorf("Could not build YAML decoder: %s", err)
    }
    yaml, err := o.YAML()
    if err != nil {
        t.Errorf("Could not decode into YAML: %s", err)
    }
    fmt.Println(string(yaml))
    */
}

func TestImportBuild(t *testing.T) {
    wd, err := os.Getwd()
    if err != nil {
        t.Errorf("Could not get PWD: %s", err)
    }
    fmt.Println(wd)
    b, err := ioutil.ReadFile(wd + "/../../../examples/build101.yaml")
    if err != nil {
        t.Errorf("Could not get YAML content: %s", err)
    }
    fmt.Println(string(b))
    o := new(oapi.Build)
    err = yaml.Unmarshal(b, o)
    if err != nil {
        t.Errorf("Could not decode into Config Object: %s", err)
    }
    fmt.Printf("%+v", o)
}

func fillBuildSpec(spec *oapi.BuildSpec) {
  if spec == nil { panic("unexpected") }
  fillBuildSource(&spec.Source)
  fillBuildOutput(&spec.Output)
  fillBuildPostCommitSpec(&spec.PostCommit)
}

func fillBuildSource(src *oapi.BuildSource) {
  s := `From alpine`
  src.Dockerfile = &s
}

func createGitBuildSource() *oapi.GitBuildSource {
    s := &oapi.GitBuildSource {
        URI: "http://github.com/stackdocker/container-ops.git",
        Ref: "gogs",
    }
    return s
}

func fillBuildOutput(out *oapi.BuildOutput) {
  
}

func fillBuildPostCommitSpec(commit *oapi.BuildPostCommitSpec) {
  
}

func TestExportBuildConfig(t *testing.T) {
    
    o := oapi.BuildConfig { }
    o.TypeMeta.APIVersion = "v1"
    o.TypeMeta.Kind = "BuildConfig"
    
    o.ObjectMeta.Name = "buildconfig101"
    o.ObjectMeta.Labels = make(map[string]string)
    o.ObjectMeta.Labels["ci"] = "buildconfig101"
    o.ObjectMeta.Labels["name"] = "buildconfig101"
    
    fillBuildSpec(&o.Spec.BuildSpec)
    
    b := new(bytes.Buffer)
    err := codec.JSON.Encode(b).One(o)
    if err != nil {
        t.Errorf("Could not encode JSON object: %s", err)
    }
    fmt.Println(b.String())
    
    /*
    o, err := codec.JSON.Decode(b.Bytes()).One()
    if err != nil {
        t.Errorf("Could not build YAML decoder: %s", err)
    }
    yaml, err := o.YAML()
    if err != nil {
        t.Errorf("Could not decode into YAML: %s", err)
    }
    fmt.Println(string(yaml))
    */
    y, err := yaml.Marshal(o)
    if err != nil {
        t.Errorf("Could not encode YAML object: %s", err)
    }
    fmt.Println(string(y))
}



var buildStr = `apiVersion: v1
kind: Build
metadata:
  creationTimestamp: null
  name: build101
  labels:
    app: build101
    ci: build101
    name: build101
Spec:
  CompletionDeadlineSeconds: null
  Output:
    PushSecret: null
    To: null
  PostCommit:
    Args: null
    Command: null
    Script: ""
  Resources: {}
  Revision: null
  ServiceAccount: ""
  Source:
    Binary: null
    ContextDir: ""
    Dockerfile: null
    Git: null
    Images: null
    Secrets: null
    SourceSecret: null
  Strategy:
    CustomStrategy: null
    DockerStrategy: null
    SourceStrategy: null
Status:
  Cancelled: false
  CompletionTimestamp: null
  Config: null
  Duration: 0
  Message: ""
  OutputDockerImageReference: ""
  Phase: ""
  Reason: ""
  StartTimestamp: null
`

var jsonBuildConfig = `
{
  "kind": "BuildConfig",
  "apiVersion": "v1",
  "metadata": {
    "name": "ruby-sample-build",
    "creationTimestamp": null,
    "labels": {
      "name": "ruby-sample-build"
    }
  },
  "spec": {
    "triggers": [
      {
        "type": "github",
        "github": {
          "secret": "secret101"
        }
      },
      {
        "type": "generic",
        "generic": {
          "secret": "secret101"
        }
      },
      {
        "type": "imageChange",
        "imageChange": {}
      },
      {
        "type": "ConfigChange"
      }
    ],
    "source": {
      "type": "Git",
      "git": {
        "uri": "git://github.com/openshift/origin.git",
        "ref": "/tree/master/examples/hello-openshift"
      }
    },
    "strategy": {
      "type": "Docker",
      "dockerStrategy": {
        "from": {
          "kind": "ImageStreamTag",
          "name": "ruby-22-centos7:latest"
        },
        "env": [
          {
            "name": "EXAMPLE",
            "value": "sample-app"
          }
        ]
      }
    },
    "output": {
      "to": {
        "kind": "ImageStreamTag",
        "name": "origin-ruby-sample:latest"
      }
    },
    "postCommit": {
      "args": ["bundle", "exec", "rake", "test"]
    },
    "resources": {}
  },
  "status": {
    "lastVersion": 0
  }
}
`