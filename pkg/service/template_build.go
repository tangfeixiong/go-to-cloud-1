package service

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/docker/engine-api/types"
	"github.com/golang/glog"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	kapi "k8s.io/kubernetes/pkg/api"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/cicd/pb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/build-builder"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/utility"
)

var (
	buildconfig_json_tpl_default = "buildconfig.json.tpl"
	vendor_create_annotation_key = "vendor/created-by"
)

func (u *UserResource) TemplateBuildingOntoStream(stream pb3.ContainerImageBuildService_TemplateBuildingOntoStreamServer) error {
	var mutex = &sync.Mutex{}
	var wg sync.WaitGroup
	var ech chan error = make(chan error)
	var flag byte
	queue := &templatizedBuilderRequestQueue{nodes: make([]*templatizedBuilderRequestQueueItem, 1)}
	var req *pb3.TemplatizedBuilderRequest
	var feed *pb3.Feed
	var ar []string = make([]string, 0) /* store received archive/file stream for later build (binary source) */
	var objbc *buildapiv1.BuildConfig
	//var objbuild *buildapiv1.Build

	wg.Add(1)
	go func() {
		defer wg.Done()
	LOOP:
		for {
			mutex.Lock()
			qitem := queue.Pop()
			mutex.Unlock()
			if qitem == nil {
				if flag > 0 {
					ech <- nil
					return
				}
				time.Sleep(time.Second * 1)
				continue
			}
			if qitem.request.Feed != nil && qitem.feed == nil {
				req = qitem.request
				if err := validateFeedSpec(req.Feed); err != nil {
					ech <- err
					return
				}
				for i := range req.Feed.Auth {
					if err := validateAuth(req.Feed.Auth[i]); err != nil {
						ech <- err
						return
					}
				}
				feed = &pb3.Feed{
					Metadata: &pb3.ObjectMeta{
						Name:        req.Feed.Option.Name,
						Namespace:   req.Feed.Option.Project,
						Annotations: make(map[string]string),
						Labels:      make(map[string]string),
					},
					Spec:   req.Feed,
					Status: &pb3.Feed_FeedStatus{},
				}
				if len(req.Feed.Option.Annotations) != 0 {
					feed.Metadata.Annotations = req.Feed.Option.Annotations
				}
				if len(req.Feed.Option.Labels) != 0 {
					feed.Metadata.Labels = req.Feed.Option.Labels
				}

				bcOpt, err := generateOption(req, feed)
				if err != nil {
					ech <- err
					return
				}
				// template engine and API call
				//
				buf := &bytes.Buffer{}
				te := template.New("conf-tpl-exec").Funcs(utility.TplFns).Funcs(utility.SprigTxtTplFns)
				if req.Feed.Template == nil || req.Feed.Template.Metadata == nil || len(req.Feed.Template.Metadata.Name) == 0 {
					te = template.Must(te.Parse(builder.SourceBuildConfigTemplate["buildconfig.json.tpl"]))
				} else {
					ech <- fmt.Errorf("service not ready")
					return
				}
				if err := te.Execute(buf, bcOpt); err != nil {
					ech <- err
					return
				}
				glog.Infof("Template executed: %+v", buf.String())
				// invoke openshift/kubernetes api
				//
				ci := origin.NewPaaS()
				if err := ci.RequestProjectCreation(req.Feed.Option.Project); err != nil {
					ech <- err
					return
				}
				for i := range req.Feed.Auth {
					switch strings.ToLower(req.Feed.Auth[i].Auth) {
					case "dockerauthconfig":
						if err := ci.RequestBuilderSecretCreationWithDockerRegistry(req.Feed.Option.Project, req.Feed.Auth[i].Id, "builder", types.AuthConfig{
							Username:      req.Feed.Auth[i].Username,
							Password:      req.Feed.Auth[i].Password,
							ServerAddress: req.Feed.Auth[i].Server}); err != nil {
							ech <- err
							return
						}
					default:
						ech <- errNotImplemented
						return
					}
				}
				bcdata, bc, err := ci.RequestBuildConfigCreation(buf.Bytes())
				if err != nil {
					glog.V(9).Infof("Openshift API called: %+v", string(bcdata))
					ech <- err
					return
				}
				// Build at once...
				//
				if req.Feed.DisableBuildAtOnce || len(req.Feed.ArchiveFilePath) != 0 {
					// write into response
					if len(feed.Metadata.Annotations) == 0 {
						feed.Metadata.Annotations = make(map[string]string)
					}
					feed.Metadata.Annotations[vendor_create_annotation_key] = buf.String()
					if len(feed.Metadata.Labels) == 0 {
						feed.Metadata.Labels = make(map[string]string)
					}
					feed.Status.Phase = fmt.Sprintf("LastVersion=%d", bc.Status.LastVersion)
				} else {
					buildname := req.Feed.BuildAtOnceName
					buildmessage := req.Feed.BuildAtOnceMessage
					if len(buildname) == 0 {
						buildname = bc.Name
					}
					if len(buildmessage) == 0 {
						buildmessage = "Manually triggered"
					}
					bdata, obj, _, err := ci.RequestBuildCreation(buildname, buildmessage, bc)
					if err != nil {
						glog.V(9).Infof("Openshift API called: %+v", string(bdata))
						ech <- err
						return
					}
					// write into response
					if len(feed.Metadata.Annotations) == 0 {
						feed.Metadata.Annotations = make(map[string]string)
					}
					feed.Metadata.Annotations[vendor_create_annotation_key] = buf.String()
					if len(feed.Metadata.Labels) == 0 {
						feed.Metadata.Labels = make(map[string]string)
					}
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
					dpch := make(chan error)
					go func(ch chan<- error) {
						//defer writer.Close()
						if err := ci.StreamBuildLog(bdata, nil); err != nil {
							fmt.Fprintln(os.Stderr, "could not read output:", err)
							ch <- err
						} else {
							ch <- nil
						}
						return
					}(dpch)
					if err := <-dpch; err != nil {
						ech <- err
					}
				}

				if err := stream.Send(&pb3.TemplatizedBuilderResponse{feed}); err != nil {
					ech <- err
					return
				}
				// laterly at once binary build maybe available
				objbc = bc
				qitem.feed = feed
			}
			// proxy archive for binary build
			if archive := qitem.request.Archive; archive != nil {
				if err := validateArchive(archive); err != nil {
					ech <- err
					return
				}
				glog.Warningln("Currently not implemented, just mock")

				// store while check archive laterly
				ar = append(ar, archive.FilePath)
			}

			// Build at once...
			//     whether FeedSpec is processed (archive arrived at first)
			//     re-queue if bin not ready
			//
			if req.Feed == nil || req.Feed.DisableBuildAtOnce {
				continue
			}
			for i := range req.Feed.ArchiveFilePath {
				for j := range ar {
					if strings.Compare(ar[j], req.Feed.ArchiveFilePath[i]) != 0 {
						queue.Push(qitem)
						continue LOOP
					}
				}
			}

			var r io.Reader = os.Stdin
			var w io.Writer = os.Stdout
			if len(req.Feed.ArchiveFilePath) > 0 {
				r, w = io.Pipe()
			}
			ci := origin.NewPaaS()
			messageReader, err := ci.OSO_startbuild_NewCmdStartBuild(objbc, nil, r)
			if err != nil {
				ech <- err
				return
			}
			w.Write([]byte("abc"))
			scanner := bufio.NewScanner(messageReader)
			for scanner.Scan() {
				fmt.Println(scanner.Text()) // Println will add back the final '\n'
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				break
			}
		}
		ech <- nil
	}()
	/*
	  https://github.com/grpc/grpc-go/blob/master/examples/route_guide/server/server.go#L133
	*/
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			flag = 1
			wg.Wait()
			return <-ech
		}
		if err != nil {
			flag = 2
			wg.Wait()
			return err
		}
		if in != nil && (in.Feed != nil || in.Archive != nil) {
			mutex.Lock()
			queue.Push(&templatizedBuilderRequestQueueItem{request: in})
			mutex.Unlock()
		}
	}
}

func generateOption(req *pb3.TemplatizedBuilderRequest, feed *pb3.Feed) (*builder.BuildConfigTemplateOption, error) {
	bcOpt := &builder.BuildConfigTemplateOption{
		ObjectMetaTemplateOption: builder.ObjectMetaTemplateOption{
			Name: req.Feed.Option.Name,
			/*Namespace: req.Feed.Option.Project,
			Annotations: []utility.NameValueStringPair{
			    {Name: "name1", Value: "value1"},
			},
			Labels: []utility.KeyValueStringPair{
			    {Key: "key1", Value: "value1"},
			},*/
		},
		/*TriggerPolicy: []interface{}{
			builder.GitHubHookTemplateOption{
				builder.WebHookTemplateOption{Secret: "git secret"},
			},
			builder.WebHookTemplateOption{Secret: "web secret"},
			builder.ImageChangeHookTemplateOption{},
			builder.ImageChangeHookTemplateOption{Kind: "ImageStreamTag", Name: "not from default strategy"},
			builder.ConfigChangeHookTemplateOption{},
		},*/
		RunPolicy: "Serial",
		CommonSpecTemplateOption: builder.CommonSpecTemplateOption{
			SourceType: "Dockerfile",
			//BinarySource: builder.BinarySourceTemplateOption{ArchiveRaw: []byte("abc"), AsFile: "/web/app.war"},
			//Dockerfile: `FROM busybox\nCMD [\"/bin/sh\"]`,
			/*GitSource: builder.GitSourceTemplateOption{
				URI:  "https://github.com/tangfeixiong/osev3-examples",
				Ref:  "master",
				Path: "/spring-boot/sample-microservices-springboot/web",
			},*/
			/*ImageSource: []ImageSourceTemplateOption{
				{
					Kind: "DockerImage",
					Name: "busybox:latest",
					Paths: []ImageSourcePath{
						{ SourcePath: "/go/bin", DestinationDir: "/opt/app" },
					},
					PullAuth: AuthTemplateOption{
						Name:           "registry-auth",
						CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
					},
				},
				{
					Kind: "ImageStreamTag",
					Name: "busybox:latest",
					Paths: []ImageSourcePath{
						{ SourcePath: "/go/bin", DestinationDir: "/bin" },
					},
				},
			},*/
			//ContextDir: "/spring-boot/sample-microservices-springboot/web",
			/*SourceAuth: builder.AuthTemplateOption{
				Name:           "source-auth",
				CredentialAuth: builder.CredentialAuth{Username: "fake", Password: "fake"},
			},*/
			/*AuthAsBuildSource: []Builder.SecretBuildSourceTemplateOption{
				{
					AuthTemplateOption: builder.AuthTemplateOption{
						Name:           "referenced-auth",
						CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
					},
					DestinationDir: "/go/src",
				},
			},*/
			SourceRevision: builder.SourceRevisionTemplateOption{},
			/*Strategy: BuildStrategyTemplateOption{
				StrategyType: "Docker",
				SourceStrategy: SourceStrategy{
					BasicStrategyTemplateOption: BasicStrategyTemplateOption{
						ImageKind: "DockerImage",
						ImageName: "busybox:latest",
						PullAuth: AuthTemplateOption{
							Name:           "registry-auth",
							CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
						},
						Env: []utility.NameValueStringPair{
							{Name: "NAME1", Value: "VALUE1"},
							{Name: "NAME2", Value: "VALUE2"},
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
						ImageKind: "DockerImage",
						ImageName: "busybox:latest",
						PullAuth: AuthTemplateOption{
							Name:           "registry-auth",
							CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
						},
						Env: []utility.NameValueStringPair{
							{Name: "NAME1", Value: "VALUE1"},
						},
					},
					NoCache:        true,
					DockerfilePath: "/docker/dockerfile",
				},
				CustomStrategy: CustomStrategy{
					BasicStrategyTemplateOption: BasicStrategyTemplateOption{
						ImageKind: "Docker Image",
						ImageName: "busybox:latest",
						PullAuth: AuthTemplateOption{
							Name:           "registry-auth",
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
			},*/
			/*Output: BuildOutputTemplateOption{
				ImageKind: "DockerImage",
				ImageName: "172.17.4.50:30005/tangfeixiong/demo:latest",
				PushAuth: AuthTemplateOption{
					Name:           "local",
					CredentialAuth: CredentialAuth{Username: "fake", Password: "fake"},
				},
			},*/
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
			PostCommitCommand:         req.Feed.Option.PostCommitCommand,
			PostCommitArgs:            req.Feed.Option.PostCommitArgs,
			PostCommitScript:          req.Feed.Option.PostCommitScript,
			CompletionDeadlineSeconds: req.Feed.Option.CompletionDeadlineSeconds,
		},
	}
	// metadata
	bcOpt.ObjectMetaTemplateOption.Namespace = req.Feed.Option.Project
	for k, v := range req.Feed.Option.Annotations {
		bcOpt.ObjectMetaTemplateOption.Annotations = append(bcOpt.ObjectMetaTemplateOption.Annotations,
			utility.NameValueStringPair{Name: k, Value: v})
	}
	for k, v := range req.Feed.Option.Labels {
		bcOpt.ObjectMetaTemplateOption.Labels = append(bcOpt.ObjectMetaTemplateOption.Labels,
			utility.KeyValueStringPair{Key: k, Value: v})
	}
	// Triggers and run policy
	for i := range req.Feed.Option.GithubWebHook {
		if auth := req.Feed.Option.GithubWebHook[i].Auth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid GithubWebHook trigger policy: %v", err)
			}
			bcOpt.TriggerPolicy = append(bcOpt.TriggerPolicy, builder.GitHubHookTemplateOption{
				builder.WebHookTemplateOption{Secret: auth.Id}})
		}
	}
	for i := range req.Feed.Option.GenericWebHook {
		if auth := req.Feed.Option.GenericWebHook[i].Auth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid VCS WebHook trigger policy: %v", err)
			}
			bcOpt.TriggerPolicy = append(bcOpt.TriggerPolicy, builder.WebHookTemplateOption{
				Secret: auth.Id, AllowEnv: req.Feed.Option.GenericWebHook[i].AllowEnv})
		}
	}
	for i := range req.Feed.Option.ImageChangeHook {
		bcOpt.TriggerPolicy = append(bcOpt.TriggerPolicy, builder.ImageChangeHookTemplateOption{
			req.Feed.Option.ImageChangeHook[i].Kind, req.Feed.Option.ImageChangeHook[i].Name})
	}
	bcOpt.RunPolicy = req.Feed.Option.RunPolicy
	// Build Source
	//
	bcOpt.SourceType = req.Feed.Option.SourceType
	//Binary and Git
	switch req.Feed.Option.SourceType {
	case "Binary":
		if len(req.Feed.Option.BinaryAsFile) > 0 {
			bcOpt.BinarySource.AsFile = req.Feed.Option.BinaryAsFile
			bcOpt.BinarySource.ContextDir = req.Feed.Option.ContextDir
		}
	case "Git":
		bcOpt.GitSource.URI = req.Feed.Option.GitSource.Uri
		bcOpt.GitSource.Ref = req.Feed.Option.GitSource.Ref
		if len(req.Feed.Option.GitSource.Path) == 0 {
			req.Feed.Option.GitSource.Path = req.Feed.Option.ContextDir
		} else {
			req.Feed.Option.ContextDir = req.Feed.Option.GitSource.Path
		}
		bcOpt.GitSource.Path = req.Feed.Option.GitSource.Path
	}
	//Dockerfile
	bcOpt.Dockerfile = req.Feed.Option.Dockerfile
	//ImageSource
	for i := range req.Feed.Option.SidecarImageSource {
		is := builder.ImageSourceTemplateOption{
			Kind: req.Feed.Option.SidecarImageSource[i].Kind,
			Name: req.Feed.Option.SidecarImageSource[i].Name,
		}
		for j := range req.Feed.Option.SidecarImageSource[i].Paths {
			if req.Feed.Option.SidecarImageSource[i].Paths[j] != nil {
				is.Paths = append(is.Paths, builder.ImageSourcePath{
					req.Feed.Option.SidecarImageSource[i].Paths[j].SourcePath,
					req.Feed.Option.SidecarImageSource[i].Paths[j].DestinationDir})
			}
		}
		if len(is.Paths) == 0 {
			return nil, fmt.Errorf("%v: invalid image source", errUnexpected)
		}
		if auth := req.Feed.Option.SidecarImageSource[i].RegistryAuth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid image as source: %v", err)
			}
			is.PullAuth = builder.AuthTemplateOption{Name: auth.Id}
		}
		bcOpt.ImageSource = append(bcOpt.ImageSource, is)
	}
	//ContextDir
	bcOpt.ContextDir = req.Feed.Option.ContextDir
	//SourceSecret
	if auth := req.Feed.Option.RepositoryAuth; auth != nil {
		if err := mergeAuth(feed.Spec, auth); err != nil {
			return nil, fmt.Errorf("invalid VCS auth: %v", err)
		}
		bcOpt.SourceAuth = builder.AuthTemplateOption{Name: auth.Id}
	}
	//Secrets
	for i := range req.Feed.Option.AuthAsBuildSource {
		if auth := req.Feed.Option.AuthAsBuildSource[i].Auth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid auth as source: %v", err)
			}
			bcOpt.SecretBuildSource = append(bcOpt.SecretBuildSource, builder.SecretBuildSourceTemplateOption{
				builder.AuthTemplateOption{Name: auth.Id}, req.Feed.Option.AuthAsBuildSource[i].DestinationDir})
			continue
		}
		return nil, fmt.Errorf("%v: auth data not ready", errUnexpected)
	}
	// Source revision
	//
	bcOpt.SourceRevision.Type = req.Feed.Option.SourceRevisionType
	if req.Feed.Option.GitSourceRevision != nil {
		bcOpt.SourceRevision.GitCommit = req.Feed.Option.GitSourceRevision.Commit
		if req.Feed.Option.GitSourceRevision.Author != nil {
			bcOpt.SourceRevision.GitAuthorName = req.Feed.Option.GitSourceRevision.Author.Name
			bcOpt.SourceRevision.GitAuthorEmail = req.Feed.Option.GitSourceRevision.Author.Email
		}
		if req.Feed.Option.GitSourceRevision.Committer != nil {
			bcOpt.SourceRevision.GitCommitterName = req.Feed.Option.GitSourceRevision.Committer.Name
			bcOpt.SourceRevision.GitCommitterEmail = req.Feed.Option.GitSourceRevision.Committer.Email
		}
		bcOpt.SourceRevision.GitMessage = req.Feed.Option.GitSourceRevision.Message
	}
	// Build Strategy
	//
	bcOpt.Strategy.StrategyType = req.Feed.Option.BuildStrategyType
	if strategy := req.Feed.Option.CustomBuildStrategy; strategy != nil {
		bcOpt.Strategy.CustomStrategy.ImageKind = strategy.ImageKind
		bcOpt.Strategy.CustomStrategy.ImageName = strategy.ImageName
		if auth := strategy.RegistryAuth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid custom build strategy: %v", err)
			}
			bcOpt.Strategy.CustomStrategy.PullAuth = builder.AuthTemplateOption{Name: auth.Id}
		}
		bcOpt.Strategy.CustomStrategy.ExposeDockerSocket = strategy.ExposeDockerSocket
		for k, v := range strategy.Env {
			bcOpt.Strategy.CustomStrategy.Env = append(bcOpt.Strategy.CustomStrategy.Env,
				utility.NameValueStringPair{Name: k, Value: v})
		}
		bcOpt.Strategy.CustomStrategy.ForcePull = strategy.ForcePull
		for i := range strategy.AuthVol {
			if auth := strategy.AuthVol[i].Auth; auth != nil {
				if err := mergeAuth(feed.Spec, auth); err != nil {
					return nil, fmt.Errorf("invalid custom build strategy: %v", err)
				}
				bcOpt.Strategy.CustomStrategy.Secrets = append(bcOpt.Strategy.CustomStrategy.Secrets,
					builder.SecretSpec{AuthTemplateOption: builder.AuthTemplateOption{Name: auth.Id},
						MountPath: req.Feed.Option.CustomBuildStrategy.AuthVol[i].MountPath})
				continue
			}
			return nil, fmt.Errorf("%v: auth data not ready", errUnexpected)
		}
		bcOpt.Strategy.CustomStrategy.BuildAPIVersion = strategy.BuildAPIVersion
	}
	if strategy := req.Feed.Option.DockerBuildStrategy; strategy != nil {
		bcOpt.Strategy.DockerStrategy.ImageKind = strategy.ImageKind
		bcOpt.Strategy.DockerStrategy.ImageName = strategy.ImageName
		if auth := strategy.RegistryAuth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid docker build strategy: %v", err)
			}
			bcOpt.Strategy.DockerStrategy.PullAuth = builder.AuthTemplateOption{Name: auth.Id}
		}
		bcOpt.Strategy.DockerStrategy.NoCache = strategy.NoCache
		for k, v := range strategy.Env {
			bcOpt.Strategy.DockerStrategy.Env = append(bcOpt.Strategy.DockerStrategy.Env,
				utility.NameValueStringPair{Name: k, Value: v})
		}
		bcOpt.Strategy.DockerStrategy.ForcePull = strategy.ForcePull
		bcOpt.Strategy.DockerStrategy.DockerfilePath = strategy.DockerfilePath
	}
	if strategy := req.Feed.Option.SourceBuildStrategy; strategy != nil {
		bcOpt.Strategy.SourceStrategy.ImageKind = strategy.ImageKind
		bcOpt.Strategy.SourceStrategy.ImageName = strategy.ImageName
		if auth := strategy.RegistryAuth; auth != nil {
			if err := mergeAuth(feed.Spec, auth); err != nil {
				return nil, fmt.Errorf("invalid source build strategy: %v", err)
			}
			bcOpt.Strategy.SourceStrategy.PullAuth = builder.AuthTemplateOption{Name: auth.Id}
		}
		for k, v := range strategy.Env {
			bcOpt.Strategy.SourceStrategy.Env = append(bcOpt.Strategy.SourceStrategy.Env,
				utility.NameValueStringPair{Name: k, Value: v})
		}
		bcOpt.Strategy.SourceStrategy.Scripts = strategy.Scripts
		bcOpt.Strategy.SourceStrategy.Incremental = strategy.Incremental
		bcOpt.Strategy.SourceStrategy.ForcePull = strategy.ForcePull
		bcOpt.Strategy.SourceStrategy.RuntimeImageKind = strategy.RuntimeImageKind
		bcOpt.Strategy.SourceStrategy.RuntimeImageName = strategy.RuntimeImageName
		for i := range strategy.RuntimeArtifacts {
			bcOpt.Strategy.SourceStrategy.RuntimeArtifacts = append(bcOpt.Strategy.SourceStrategy.RuntimeArtifacts,
				builder.ImageSourcePath{strategy.RuntimeArtifacts[i].SourcePath, strategy.RuntimeArtifacts[i].DestinationDir})
		}
	}
	if strategy := req.Feed.Option.JenkinsPipelineStrategy; strategy != nil {
		bcOpt.Strategy.JenkinsPipelineStrategy.JenkinsfilePath = strategy.JenkinsfilePath
		bcOpt.Strategy.JenkinsPipelineStrategy.Jenkinsfile = strategy.Jenkinsfile
	}
	// Build Output
	//
	bcOpt.Output.ImageKind = req.Feed.Option.ImageKind
	bcOpt.Output.ImageName = req.Feed.Option.ImageName
	if auth := req.Feed.Option.RegistryAuth; auth != nil {
		if err := mergeAuth(feed.Spec, auth); err != nil {
			return nil, fmt.Errorf("invalid build output: %v", err)
		}
		bcOpt.Output.PushAuth = builder.AuthTemplateOption{Name: auth.Id}
	}
	// Resource
	//
	for k, v := range req.Feed.Option.ResourceLimits {
		bcOpt.ResourceLimits = append(bcOpt.ResourceLimits,
			builder.ResourceTemplateOption{Name: k, Quantity: v})
	}
	for k, v := range req.Feed.Option.ResourceRequests {
		bcOpt.ResourceRequests = append(bcOpt.ResourceRequests,
			builder.ResourceTemplateOption{Name: k, Quantity: v})
	}
	// Post commit and deadline
	//
	return bcOpt, nil
}

func validateFeedSpec(feed *pb3.Feed_FeedSpec) error {
	if feed.Builder != pb3.Feed_FeedSpec_OPENSHIFT_ORIGIN_V3 {
		return fmt.Errorf("%v: currently only implement openshift origion build", errNotImplemented)
	}

	if feed.Auth == nil {
		feed.Auth = make([]*pb3.IdentifiedAuth, 0)
	}

	/*if feed.Template == nil {
		feed.Template = &pb3.Template{
			Metadata: &pb3.Template_ObjectMeta{
				Name: buildconfig_json_tpl_default,
			},
		}
	}
	if feed.Template.Metadata == nil {
		feed.Template.Metadata = &pb3.Template_ObjectMeta{
			Name: buildconfig_json_tpl_default,
		}
	}*/

	if feed.Option == nil || len(feed.Option.Name) == 0 {
		return fmt.Errorf("%v: Build config name required", errBadRequest)
	}

	feed.Option.RunPolicy = strings.Title(feed.Option.RunPolicy)
	switch feed.Option.RunPolicy {
	case "Serial":
	case "Parallel":
	case "Seriallatestonly":
		feed.Option.RunPolicy = "SerialLatestOnly"
	default:
		return fmt.Errorf("%v: unknown build policy", errBadRequest)
	}

	gitNone := feed.Option.GitSource == nil ||
		len(feed.Option.GitSource.Uri) == 0
	dockerfileNone := len(feed.Option.Dockerfile) == 0
	imageNone := len(feed.Option.SidecarImageSource) == 0
	feed.Option.SourceType = strings.Title(feed.Option.SourceType)
	switch feed.Option.SourceType {
	case "Git":
		if gitNone && dockerfileNone {
			return fmt.Errorf("%v: Git URI required", errBadRequest)
		}
	case "Dockerfile":
		if dockerfileNone {
			return fmt.Errorf("%v: Image source required", errBadRequest)
		}
	case "Binary":
		break
	case "Image":
		if imageNone && dockerfileNone {
			return fmt.Errorf("%v: Image source required", errBadRequest)
		}
	case "None":
		break
	default:
		return fmt.Errorf("%v: openshift origin only support Git, Dockerfile, Binary, Image, None", errNotSupported)
	}

	feed.Option.SourceRevisionType = strings.Title(feed.Option.SourceRevisionType)
	switch feed.Option.SourceRevisionType {
	case "Source":
	case "Dockerfile":
	case "Binary":
	case "Images":
	default:
		//return fmt.Errorf("%v: bad source revision type", errBadRequest)
	}

	feed.Option.BuildStrategyType = strings.Title(feed.Option.BuildStrategyType)
	switch feed.Option.BuildStrategyType {
	case "Docker":
	case "Source":
	case "Custome":
	case "Jenkinspipeline":
		feed.Option.BuildStrategyType = "JenkinsPipeline"
	default:
		return fmt.Errorf("%v: bad build strategy type", errBadRequest)
	}

	if feed.Option.CompletionDeadlineSeconds < 0 {
		return fmt.Errorf("%v: must positive integerial number", errBadRequest)
	}
	return nil
}
func validateArchive(ar *pb3.Stream_StreamSpec) error {
	if ar.StreamType == pb3.Stream_StreamSpec_FILE && len(ar.FileContent) == 0 {
		return fmt.Errorf("%v: file content required", errBadRequest)
	}
	if ar.StreamType == pb3.Stream_StreamSpec_URL && len(ar.Url) == 0 {
		return fmt.Errorf("%v: URL required", errBadRequest)
	}
	if len(ar.FilePath) == 0 {
		return fmt.Errorf("%v: destination path required", errBadRequest)
	}
	return nil
}

func validateAuth(auth *pb3.IdentifiedAuth) error {
	if len(auth.Id) == 0 {
		if len(auth.Server) == 0 || len(auth.Username) == 0 {
			return fmt.Errorf("%v: ID required", errBadRequest)
		}
		sEnc := strings.Replace(strings.Replace(auth.Server, ".", "-dot-", -1), ":", "-colon-", -1)
		auth.Id = fmt.Sprintf("%s-slash-%s", sEnc, auth.Username)
	}
	switch strings.ToLower(auth.Auth) {
	case string(kapi.SecretTypeBasicAuth), "dockerauthconfig", "basic", "basicauth", "basic-auth":
		if len(auth.Username) != 0 && len(auth.Password) != 0 {
			data := fmt.Sprintf("%s:%s", auth.Username, auth.Password)
			sEnc := base64.StdEncoding.EncodeToString([]byte(data))
			auth.Token = sEnc
			break
		}
		if sEnc := auth.Token; len(auth.Token) != 0 {
			sDec, err := base64.StdEncoding.DecodeString(sEnc)
			if err != nil {
				return fmt.Errorf("%v: token invalid", errBadRequest)
			}
			data := strings.Split(string(sDec), ":")
			if len(data) == 2 {
				auth.Username, auth.Password = data[0], data[1]
			}
			break
		}
		return fmt.Errorf("%v: credential/token required", errBadRequest)
	case string(kapi.SecretTypeSSHAuth), "ssh", "ssh-auth", "ssh-rsa":
		if len(auth.SshAuthPrivate) == 0 {
			return fmt.Errorf("%v: SSH private DSA/RSA required", errBadRequest)
		}
	case string(kapi.SecretTypeTLS), "tls", "ssl", "x509":
		if len(auth.TlsCert) == 0 || len(auth.TlsPrivateKey) == 0 {
			return fmt.Errorf("%v: TLS/SSL cert and key required", errBadRequest)
		}
	case string(kapi.SecretTypeOpaque):
		if len(auth.Token) == 0 {
			return fmt.Errorf("%v: opaque token required", errBadRequest)
		}
	case "":
		if len(auth.Auth) != 0 || len(auth.Server) != 0 ||
			len(auth.Username) != 0 || len(auth.Password) != 0 || len(auth.Token) != 0 ||
			len(auth.SshAuthPrivate) != 0 ||
			len(auth.TlsCert) != 0 || len(auth.TlsPrivateKey) != 0 {
			return fmt.Errorf("%v: credential or private data not required", errBadRequest)
		}
	default:
		return fmt.Errorf("%v: unkown auth", errBadRequest)
	}
	return nil
}

func mergeAuth(feed *pb3.Feed_FeedSpec, auth *pb3.IdentifiedAuth) error {
	if len(auth.Id) != 0 && len(auth.Auth) == 0 && len(auth.Server) == 0 &&
		len(auth.Username) == 0 && len(auth.Password) == 0 && len(auth.Token) == 0 &&
		len(auth.SshAuthPrivate) == 0 &&
		len(auth.TlsCert) == 0 && len(auth.TlsPrivateKey) == 0 {
		return nil
	}

	if err := validateAuth(auth); err != nil {
		return err
	}

	for i := range feed.Auth {
		if strings.Compare(auth.Id, feed.Auth[i].Id) == 0 {
			if len(auth.Auth) != 0 && strings.Compare(auth.Auth, feed.Auth[i].Auth) != 0 {
				return fmt.Errorf("%v: AUTH not identical", errBadRequest)
			}
			if len(auth.Server) != 0 && strings.Compare(auth.Server, feed.Auth[i].Server) != 0 {
				return fmt.Errorf("%v: Server not identical", errBadRequest)
			}
			if len(auth.Username) != 0 && len(auth.Password) != 0 &&
				strings.Compare(auth.Username, feed.Auth[i].Username) != 0 &&
				strings.Compare(auth.Password, feed.Auth[i].Password) != 0 {
				return fmt.Errorf("%v: credential not identical", errBadRequest)
			}
			if len(auth.Token) != 0 && strings.Compare(auth.Token, feed.Auth[i].Token) != 0 {
				return fmt.Errorf("%v: Token not identical", errBadRequest)
			}
			if len(auth.SshAuthPrivate) != 0 && strings.Compare(auth.SshAuthPrivate, feed.Auth[i].SshAuthPrivate) != 0 {
				return fmt.Errorf("%v: SSH auth not identical", errBadRequest)
			}
			if len(auth.TlsCert) != 0 && len(auth.TlsPrivateKey) != 0 &&
				strings.Compare(auth.TlsCert, feed.Auth[i].TlsCert) != 0 &&
				strings.Compare(auth.TlsPrivateKey, feed.Auth[i].TlsPrivateKey) != 0 {
				return fmt.Errorf("%v: TLS not identical", errBadRequest)
			}
			auth.Auth = ""
			auth.Server = ""
			auth.Username = ""
			auth.Password = ""
			auth.Token = ""
			auth.SshAuthPrivate = ""
			auth.TlsCert = ""
			auth.TlsPrivateKey = ""
			return nil
		}
	}

	feed.Auth = append(feed.Auth, auth)
	return nil
}
