package origin

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/helm/helm-classic/codec"
	"github.com/spf13/cobra"

	"golang.org/x/net/context"

	kapi "k8s.io/kubernetes/pkg/api"
	kerrors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/fields"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/runtime/serializer"

	buildapi "github.com/openshift/origin/pkg/build/api"
	buildapiv1 "github.com/openshift/origin/pkg/build/api/v1"

	osclient "github.com/openshift/origin/pkg/client"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/openshift/origin/pkg/generate/git"
	oerrors "github.com/openshift/origin/pkg/util/errors"
	"github.com/openshift/source-to-image/pkg/tar"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/gnatsd"
)

type StartBuildOptions struct {
	cmd.StartBuildOptions
	Ctx   context.Context
	Req   *osopb3.DockerBuildRequestData
	Resp  *osopb3.DockerBuildResponseData
	OP    *PaaS
	Raw   []byte
	Obj   *buildapiv1.Build
	BC    *buildapiv1.BuildConfig
	buf   *bytes.Buffer
	mutex *sync.Mutex
}

func NewCmdStartBuild(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) (*cobra.Command, *StartBuildOptions) {
	o := &StartBuildOptions{
		StartBuildOptions: cmd.StartBuildOptions{
			LogLevel:        "5",
			Follow:          true,
			WaitForComplete: true,
		},
		buf:   &bytes.Buffer{},
		mutex: &sync.Mutex{},
	}

	cmd := &cobra.Command{
		Use:        "start-build (BUILDCONFIG | --from-build=BUILD)",
		Short:      "Start a new build",
		Long:       startBuildLong,
		Example:    fmt.Sprintf(startBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(o.Complete(f, in, out, cmd, args))
			kcmdutil.CheckErr(o.Run())
		},
	}
	cmd.Flags().StringVar(&o.LogLevel, "build-loglevel", o.LogLevel, "Specify the log level for the build log output")
	cmd.Flags().Lookup("build-loglevel").NoOptDefVal = "5"
	cmd.Flags().StringSliceVarP(&o.Env, "env", "e", o.Env, "Specify key value pairs of environment variables to set for the build container.")

	cmd.Flags().StringVar(&o.FromBuild, "from-build", o.FromBuild, "Specify the name of a build which should be re-run")

	cmd.Flags().BoolVar(&o.Follow, "follow", o.Follow, "Start a build and watch its logs until it completes or fails")
	cmd.Flags().Lookup("follow").NoOptDefVal = "true"
	cmd.Flags().BoolVar(&o.WaitForComplete, "wait", o.WaitForComplete, "Wait for a build to complete and exit with a non-zero return code if the build fails")

	cmd.Flags().StringVar(&o.FromFile, "from-file", o.FromFile, "A file to use as the binary input for the build; example a pom.xml or Dockerfile. Will be the only file in the build source.")
	cmd.Flags().StringVar(&o.FromDir, "from-dir", o.FromDir, "A directory to archive and use as the binary input for a build.")
	cmd.Flags().StringVar(&o.FromRepo, "from-repo", o.FromRepo, "The path to a local source code repository to use as the binary input for a build.")
	cmd.Flags().StringVar(&o.Commit, "commit", o.Commit, "Specify the source code commit identifier the build should use; requires a build based on a Git repository")

	cmd.Flags().StringVar(&o.ListWebhooks, "list-webhooks", o.ListWebhooks, "List the webhooks for the specified build config or build; accepts 'all', 'generic', or 'github'")
	cmd.Flags().StringVar(&o.FromWebhook, "from-webhook", o.FromWebhook, "Specify a webhook URL for an existing build config to trigger")

	cmd.Flags().StringVar(&o.GitPostReceive, "git-post-receive", o.GitPostReceive, "The contents of the post-receive hook to trigger a build")
	cmd.Flags().StringVar(&o.GitRepository, "git-repository", o.GitRepository, "The path to the git repository for post-receive; defaults to the current directory")

	// cmdutil.AddOutputFlagsForMutation(cmd)
	return cmd, o
}

func (o *StartBuildOptions) TrackWith(ctx context.Context,
	req *osopb3.DockerBuildRequestData, resp *osopb3.DockerBuildResponseData,
	op *PaaS, raw []byte, obj *buildapiv1.Build, bc *buildapiv1.BuildConfig) func() {
	o.Ctx = ctx
	o.Req = req
	o.OP = op
	o.Raw = raw
	o.Obj = obj
	o.BC = bc
	o.Resp = resp
	o.ClientConfig = o.OP.Factory().OpenShiftClientConfig
	return o.tracker
}

func GenerateResponseData(raw []byte, obj *buildapiv1.Build) *osopb3.DockerBuildResponseData {
	return &osopb3.DockerBuildResponseData{
		Status: &osopb3.OsoBuildStatus{
			Phase:                      string(obj.Status.Phase),
			Cancelled:                  obj.Status.Cancelled,
			Reason:                     string(obj.Status.Reason),
			StartTimestamp:             obj.Status.StartTimestamp,
			CompletionTimestamp:        obj.Status.CompletionTimestamp,
			Duration:                   int64(obj.Status.Duration),
			OutputDockerImageReference: obj.Status.OutputDockerImageReference,
			Config:        obj.Status.Config,
			OsoBuildPhase: osopb3.OsoBuildStatus_OsoBuildPhase(osopb3.OsoBuildStatus_OsoBuildPhase_value[string(obj.Status.Phase)]),
		},
		Raw: &osopb3.RawJSON{
			ObjectGVK: unversioned.GroupVersionKind{
				Group:   "",
				Version: obj.TypeMeta.APIVersion,
				Kind:    obj.TypeMeta.Kind,
			}.String(),
			ObjectJSON: raw,
		},
	}
}

func Subject(ns, name string) string {
	return fmt.Sprintf("/namespaces/%s/builds/%s", ns, name)
}

func convertIntoV1WithRuntimeObject(obj runtime.Object) (b []byte, v *buildapiv1.Build, ok bool) {
	buf := &bytes.Buffer{}
	if err := codec.JSON.Encode(buf).One(obj); err != nil {
		glog.Errorf("Failed to decode runtime object: %+v", err)
		return
	}
	b = buf.Bytes()
	hco, err := codec.JSON.Decode(b).One()
	if err != nil {
		glog.Errorf("Failed to decode runtime object: %+v", err)
		return
	}
	v = new(buildapiv1.Build)
	if err := hco.Object(obj); err != nil {
		glog.Errorf("Failed to decode runtime object: %+v", err)
		return
	}
	ok = true
	return
}

func (o *StartBuildOptions) cacheBuilds(raw []byte, obj *buildapiv1.Build) {
	/*raw, obj, err := ConvertBuildIntoV1(nil, v)
	if err != nil {
		glog.Warningf("Failed to read runtime object: %+v", err)
		return
	}*/

	resp := GenerateResponseData(raw, obj)
	if bytes.Compare(raw, o.Raw) != 0 {
		o.Raw = raw
		o.Obj = obj
		o.Resp = resp
		if b, err := o.Resp.Marshal(); err != nil {
			glog.Infof("cache: %+v", string(b))
			//gnatsd.Publish([]string{}, nil, nil, Subject(o.Namespace, o.Req.Name), b)
		}
	} else {
		glog.Infof("noting to cache")
	}
}

func (o *StartBuildOptions) cacheLogs() {
	var b []byte
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.Resp != nil {
		if o.Resp.Status != nil {
			o.Resp.Status.Message = o.buf.String()
		}
		var err error
		b, err = o.Resp.Marshal()
		if err != nil || len(b) == 0 {
			glog.Errorf("Failed to operate message: %+v, data: %+v", err, o.Resp)
			return
		}
		gnatsd.Publish([]string{}, nil, nil, Subject(o.Namespace, o.Req.Name), b)
		if glog.V(2) {
			glog.V(2).Infof("Publish message: %+v", o.Resp)
		} else {
			glog.Infof("Publish message: %+v", o.Resp)
		}
	}
}

func (o *StartBuildOptions) restClient() (*restclient.RESTClient, error) {
	config, err := o.ClientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	kapi.Scheme.AddKnownTypes(buildapiv1.SchemeGroupVersion, &buildapiv1.BuildConfig{})
	kapi.Scheme.AddKnownTypes(buildapi.SchemeGroupVersion, &buildapi.BuildConfig{})
	cf := serializer.NewCodecFactory(kapi.Scheme)
	config.GroupVersion = &buildapi.SchemeGroupVersion
	config.NegotiatedSerializer = &cf
	return restclient.RESTClientFor(config)
}

func (o *StartBuildOptions) tracker() {
	var (
		wg       sync.WaitGroup
		exitErr  error
		c        osclient.BuildInterface = o.Client.Builds(o.Namespace)
		name     string                  = o.Obj.Name
		newBuild *buildapi.Build         = V1ToBuild(o.Obj)
	)
	o.Out = o.buf
	o.ErrOut = o.buf

	// Wait for the build to complete
	wg.Add(1)
	go func() {
		defer wg.Done()
		//exitErr = WaitForBuildComplete(o.Client.Builds(o.Namespace), newBuild.Name)
		for {
			//raw, err := o.Client.RESTClient.Verb("GET").Namespace(project).Resource("Builds").
			//	VersionedParams(kapiv1.ListOptions{FieldSelector: fields.Set{"name": name}.AsSelector()}).
			//	DoRaw()
			list, err := c.List(kapi.ListOptions{FieldSelector: fields.Set{"name": name}.AsSelector()})
			if err != nil {
				glog.Errorf("Failed to list builds (%s): %+v", name, err)
				exitErr = err
				return
			}
			if len(list.Items) == 0 {
				glog.Warningln("Unexpected")
			}
			for i := range list.Items {
				if name == list.Items[i].Name &&
					list.Items[i].Status.Phase == buildapi.BuildPhaseComplete {
					glog.Infof("Build %+v is completed", name)
					exitErr = nil
					return
				}
				if name != list.Items[i].Name ||
					list.Items[i].Status.Phase == buildapi.BuildPhaseFailed ||
					list.Items[i].Status.Phase == buildapi.BuildPhaseCancelled ||
					list.Items[i].Status.Phase == buildapi.BuildPhaseError {
					glog.Errorf("Unexpected %s/%s status: %+v", list.Items[i].Namespace, list.Items[i].Name, list.Items[i].Status.Phase)
					exitErr = fmt.Errorf("the build %s/%s status is %q", list.Items[i].Namespace, list.Items[i].Name, list.Items[i].Status.Phase)
					return
				}
			}

			rv := list.ResourceVersion
			c, err := o.restClient()
			if err != nil {
				glog.Errorf("Failed to setup restclient: %+v", err)
				exitErr = err
				return
			}
			w, err := c.Verb("GET").Prefix("watch").Namespace(o.Namespace).Resource("builds").
				VersionedParams(&kapi.ListOptions{
					FieldSelector:   fields.Set{"name": name}.AsSelector(),
					ResourceVersion: rv,
				}, kapi.ParameterCodec).Watch()
			//w, err := c.Watch(kapi.ListOptions{FieldSelector: fields.Set{"name": name}.AsSelector(), ResourceVersion: rv})
			if err != nil {
				glog.Errorf("Failed to setup watcher (%s): %+v", name, err)
				exitErr = err
				return
			}
			defer w.Stop()
			if w == nil {
				glog.Warningln("Failed to setup watcher: nil")
				exitErr = fmt.Errorf("Failed to setup watcher: nil")
				return
			}

			for {
				val, ok := <-w.ResultChan()
				if !ok {
					// reget and re-watch
					glog.Infoln("reget and re-watch")
					break
				}
				//if e, ok := val.Object.(*buildapi.Build); ok {
				if b, e, ok := convertIntoV1WithRuntimeObject(val.Object); ok {
					o.cacheBuilds(b, e)
					if name == e.Name && e.Status.Phase == buildapiv1.BuildPhaseComplete {
						glog.Infoln("completed")
						exitErr = nil
						return
					}
					if name != e.Name || e.Status.Phase == buildapiv1.BuildPhaseFailed ||
						e.Status.Phase == buildapiv1.BuildPhaseCancelled ||
						e.Status.Phase == buildapiv1.BuildPhaseError {
						exitErr = fmt.Errorf("The build %s/%s status is %q", e.Namespace, name, e.Status.Phase)
						glog.Warningf("failed: %+v", exitErr)
						return
					}
				}
			}
		}
	}()

	// Stream the logs from the build
	wg.Add(1)
	go func() {
		// if --wait option is set, then don't wait for logs to finish streaming
		// but wait for the build to reach its final state
		if o.WaitForComplete {
			wg.Done()
		} else {
			defer wg.Done()
		}
		opts := buildapi.BuildLogOptions{
			Follow: true,
			NoWait: false,
		}
		for {
			rd, err := o.Client.BuildLogs(o.Namespace).Get(newBuild.Name, opts).Stream()
			if err != nil {
				// if --wait options is set, then retry the connection to build logs
				// when we hit the timeout.
				if o.WaitForComplete && oerrors.IsTimeoutErr(err) {
					continue
				}
				fmt.Fprintf(o.ErrOut, "error getting logs: %v\n", err)
				o.cacheLogs()
				return
			}
			defer rd.Close()
			if _, err = io.Copy(o.Out, rd); err != nil {
				fmt.Fprintf(o.ErrOut, "error streaming logs: %v\n", err)
			}
			o.cacheLogs()
			break
		}
	}()

	wg.Wait()
	if exitErr != nil {
		if glog.V(2) {
			glog.V(2).Infof("Failed to watch building: %+v", exitErr)
		} else {
			glog.Warningf("Failed to watch building: %+v", exitErr)
		}
	}
}

func (o *StartBuildOptions) Complete(f *clientcmd.Factory, in io.Reader, out io.Writer, cmd *cobra.Command, args []string) error {
	o.In = in
	o.Out = out
	o.ErrOut = cmd.Out()
	o.Git = git.NewRepository()
	o.ClientConfig = f.OpenShiftClientConfig

	webhook := o.FromWebhook
	buildName := o.FromBuild
	fromFile := o.FromFile
	fromDir := o.FromDir
	fromRepo := o.FromRepo
	buildLogLevel := o.LogLevel

	switch {
	case len(webhook) > 0:
		if len(args) > 0 || len(buildName) > 0 || len(fromFile) > 0 || len(fromDir) > 0 || len(fromRepo) > 0 {
			return kcmdutil.UsageError(cmd, "The '--from-webhook' flag is incompatible with arguments and all '--from-*' flags")
		}
		return nil

	case len(args) != 1 && len(buildName) == 0:
		return kcmdutil.UsageError(cmd, "Must pass a name of a build config or specify build name with '--from-build' flag")
	}

	if len(buildName) != 0 && (len(fromFile) != 0 || len(fromDir) != 0 || len(fromRepo) != 0) {
		// TODO: we should support this, it should be possible to clone a build to run again with new uploaded artifacts.
		// Doing so requires introducing a new clonebinary endpoint.
		return kcmdutil.UsageError(cmd, "Cannot use '--from-build' flag with binary builds")
	}
	o.AsBinary = len(fromFile) > 0 || len(fromDir) > 0 || len(fromRepo) > 0

	namespace, _, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	client, _, err := f.Clients()
	if err != nil {
		return err
	}
	o.Client = client

	var (
		name     = buildName
		resource = buildapi.Resource("builds")
	)

	if len(name) == 0 && len(args) > 0 && len(args[0]) > 0 {
		mapper, _ := f.Object(false)
		resource, name, err = cmdutil.ResolveResource(buildapi.Resource("buildconfigs"), args[0], mapper)
		if err != nil {
			return err
		}
		switch resource {
		case buildapi.Resource("buildconfigs"):
			// no special handling required
		case buildapi.Resource("builds"):
			if len(o.ListWebhooks) == 0 {
				return fmt.Errorf("use --from-build to rerun your builds")
			}
		default:
			return fmt.Errorf("invalid resource provided: %v", resource)
		}
	}
	// when listing webhooks, allow --from-build to lookup a build config
	if resource == buildapi.Resource("builds") && len(o.ListWebhooks) > 0 {
		build, err := client.Builds(namespace).Get(name)
		if err != nil {
			return err
		}
		ref := build.Status.Config
		if ref == nil {
			return fmt.Errorf("the provided Build %q was not created from a BuildConfig and cannot have webhooks", name)
		}
		if len(ref.Namespace) > 0 {
			namespace = ref.Namespace
		}
		name = ref.Name
	}

	if len(name) == 0 {
		return fmt.Errorf("a resource name is required either as an argument or by using --from-build")
	}

	o.Namespace = namespace
	o.Name = name

	env, _, err := cmdutil.ParseEnv(o.Env, in)
	if err != nil {
		return err
	}
	if len(buildLogLevel) > 0 {
		env = append(env, kapi.EnvVar{Name: "BUILD_LOGLEVEL", Value: buildLogLevel})
	}
	o.EnvVar = env

	return nil
}

// Run contains all the necessary functionality for the OpenShift cli start-build command
func (o *StartBuildOptions) Run() error {
	if len(o.FromWebhook) > 0 {
		return o.RunStartBuildWebHook()
	}
	if len(o.ListWebhooks) > 0 {
		return o.RunListBuildWebHooks()
	}
	buildRequestCauses := []buildapi.BuildTriggerCause{}
	request := &buildapi.BuildRequest{
		TriggeredBy: append(buildRequestCauses,
			buildapi.BuildTriggerCause{
				Message: "Manually triggered",
			},
		),
		ObjectMeta: kapi.ObjectMeta{Name: o.Name},
	}
	if len(o.EnvVar) > 0 {
		request.Env = o.EnvVar
	}
	if len(o.Commit) > 0 {
		request.Revision = &buildapi.SourceRevision{
			Git: &buildapi.GitSourceRevision{
				Commit: o.Commit,
			},
		}
	}

	var err error
	var newBuild *buildapi.Build
	switch {
	case o.AsBinary:
		request := &buildapi.BinaryBuildRequestOptions{
			ObjectMeta: kapi.ObjectMeta{
				Name:      o.Name,
				Namespace: o.Namespace,
			},
			Commit: o.Commit,
		}
		if len(o.EnvVar) > 0 {
			fmt.Fprintf(o.ErrOut, "WARNING: Specifying environment variables with binary builds is not supported.\n")
		}
		if newBuild, err = streamPathToBuild(o.Git, o.In, o.ErrOut, o.Client.BuildConfigs(o.Namespace), o.FromDir, o.FromFile, o.FromRepo, request); err != nil {
			return err
		}
	case len(o.FromBuild) > 0:
		if newBuild, err = o.Client.Builds(o.Namespace).Clone(request); err != nil {
			if isInvalidSourceInputsError(err) {
				return fmt.Errorf("Build %s/%s has no valid source inputs and '--from-build' cannot be used for binary builds", o.Namespace, o.Name)
			}
			return err
		}
	default:
		if newBuild, err = o.Client.BuildConfigs(o.Namespace).Instantiate(request); err != nil {
			if isInvalidSourceInputsError(err) {
				return fmt.Errorf("Build configuration %s/%s has no valid source inputs, if this is a binary build you must specify one of '--from-dir', '--from-repo', or '--from-file'", o.Namespace, o.Name)
			}
			return err
		}
	}

	// TODO: support -o on this command
	fmt.Fprintln(o.Out, newBuild.Name)

	var (
		wg      sync.WaitGroup
		exitErr error
	)

	// Wait for the build to complete
	if o.WaitForComplete {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exitErr = WaitForBuildComplete(o.Client.Builds(o.Namespace), newBuild.Name)
		}()
	}

	// Stream the logs from the build
	if o.Follow {
		wg.Add(1)
		go func() {
			// if --wait option is set, then don't wait for logs to finish streaming
			// but wait for the build to reach its final state
			if o.WaitForComplete {
				wg.Done()
			} else {
				defer wg.Done()
			}
			opts := buildapi.BuildLogOptions{
				Follow: true,
				NoWait: false,
			}
			for {
				rd, err := o.Client.BuildLogs(o.Namespace).Get(newBuild.Name, opts).Stream()
				if err != nil {
					// if --wait options is set, then retry the connection to build logs
					// when we hit the timeout.
					if o.WaitForComplete && oerrors.IsTimeoutErr(err) {
						continue
					}
					fmt.Fprintf(o.ErrOut, "error getting logs: %v\n", err)
					return
				}
				defer rd.Close()
				if _, err = io.Copy(o.Out, rd); err != nil {
					fmt.Fprintf(o.ErrOut, "error streaming logs: %v\n", err)
				}
				break
			}
		}()
	}

	wg.Wait()

	return exitErr
}

// RunListBuildWebHooks prints the webhooks for the provided build config.
func (o *StartBuildOptions) RunListBuildWebHooks() error {
	generic, github := false, false
	prefix := false
	switch o.ListWebhooks {
	case "all":
		generic, github = true, true
		prefix = true
	case "generic":
		generic = true
	case "github":
		github = true
	default:
		return fmt.Errorf("--list-webhooks must be 'all', 'generic', or 'github'")
	}
	client := o.Client

	config, err := client.BuildConfigs(o.Namespace).Get(o.Name)
	if err != nil {
		return err
	}

	for _, t := range config.Spec.Triggers {
		hookType := ""
		switch {
		case t.GenericWebHook != nil && generic:
			if prefix {
				hookType = "generic "
			}
		case t.GitHubWebHook != nil && github:
			if prefix {
				hookType = "github "
			}
		default:
			continue
		}
		url, err := client.BuildConfigs(o.Namespace).WebHookURL(o.Name, &t)
		if err != nil {
			if err != osclient.ErrTriggerIsNotAWebHook {
				fmt.Fprintf(o.ErrOut, "error: unable to get webhook for %s: %v", o.Name, err)
			}
			continue
		}
		fmt.Fprintf(o.Out, "%s%s\n", hookType, url.String())
	}
	return nil
}

func streamPathToBuild(git git.Repository, in io.Reader, out io.Writer, client osclient.BuildConfigInterface, fromDir, fromFile, fromRepo string, options *buildapi.BinaryBuildRequestOptions) (*buildapi.Build, error) {
	count := 0
	asDir, asFile, asRepo := len(fromDir) > 0, len(fromFile) > 0, len(fromRepo) > 0
	if asDir {
		count++
	}
	if asFile {
		count++
	}
	if asRepo {
		count++
	}
	if count > 1 {
		return nil, fmt.Errorf("only one of --from-file, --from-repo, or --from-dir may be specified")
	}

	var r io.Reader
	switch {
	case fromFile == "-":
		return nil, fmt.Errorf("--from-file=- is not supported")

	case fromDir == "-":
		br := bufio.NewReaderSize(in, 4096)
		r = br
		if !isArchive(br) {
			fmt.Fprintf(out, "WARNING: the provided file may not be an archive (tar, tar.gz, or zip), use --from-file=- instead\n")
		}
		fmt.Fprintf(out, "Uploading archive file from STDIN as binary input for the build ...\n")

	default:
		var fromPath string
		switch {
		case asDir:
			fromPath = fromDir
		case asFile:
			fromPath = fromFile
		case asRepo:
			fromPath = fromRepo
		}

		clean := filepath.Clean(fromPath)
		path, err := filepath.Abs(fromPath)
		if err != nil {
			return nil, err
		}

		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if stat.IsDir() {
			commit := "HEAD"
			if len(options.Commit) > 0 {
				commit = options.Commit
			}
			info, gitErr := gitRefInfo(git, clean, commit)
			if gitErr == nil {
				options.Commit = info.GitSourceRevision.Commit
				options.Message = info.GitSourceRevision.Message
				options.AuthorName = info.GitSourceRevision.Author.Name
				options.AuthorEmail = info.GitSourceRevision.Author.Email
				options.CommitterName = info.GitSourceRevision.Committer.Name
				options.CommitterEmail = info.GitSourceRevision.Committer.Email
			} else {
				glog.V(6).Infof("Unable to read Git info from %q: %v", clean, gitErr)
			}

			if asRepo {
				fmt.Fprintf(out, "Uploading %q at commit %q as binary input for the build ...\n", clean, commit)
				if gitErr != nil {
					return nil, fmt.Errorf("the directory %q is not a valid Git repository: %v", clean, gitErr)
				}
				pr, pw := io.Pipe()
				go func() {
					if err := git.Archive(clean, options.Commit, "tar.gz", pw); err != nil {
						pw.CloseWithError(fmt.Errorf("unable to create Git archive of %q for build: %v", clean, err))
					} else {
						pw.CloseWithError(io.EOF)
					}
				}()
				r = pr

			} else {
				fmt.Fprintf(out, "Uploading directory %q as binary input for the build ...\n", clean)

				pr, pw := io.Pipe()
				go func() {
					w := gzip.NewWriter(pw)
					if err := tar.New().CreateTarStream(path, false, w); err != nil {
						pw.CloseWithError(err)
					} else {
						w.Close()
						pw.CloseWithError(io.EOF)
					}
				}()
				r = pr
			}
		} else {
			f, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer f.Close()

			r = f

			if asFile {
				options.AsFile = filepath.Base(path)
				fmt.Fprintf(out, "Uploading file %q as binary input for the build ...\n", clean)
			} else {
				br := bufio.NewReaderSize(f, 4096)
				r = br
				if !isArchive(br) {
					fmt.Fprintf(out, "WARNING: the provided file may not be an archive (tar, tar.gz, or zip), use --as-file\n")
				}
				fmt.Fprintf(out, "Uploading archive file %q as binary input for the build ...\n", clean)
			}
		}
	}
	return client.InstantiateBinary(options, r)
}

func isArchive(r *bufio.Reader) bool {
	data, err := r.Peek(280)
	if err != nil {
		return false
	}
	for _, b := range [][]byte{
		{0x50, 0x4B, 0x03, 0x04}, // zip
		{0x1F, 0x9D},             // tar.z
		{0x1F, 0xA0},             // tar.z
		{0x42, 0x5A, 0x68},       // bz2
		{0x1F, 0x8B, 0x08},       // gzip
	} {
		if bytes.HasPrefix(data, b) {
			return true
		}
	}
	switch {
	// Unified TAR files have this magic number
	case len(data) > 257+5 && bytes.Equal(data[257:257+5], []byte{0x75, 0x73, 0x74, 0x61, 0x72}):
		return true
	default:
		return false
	}
}

// RunStartBuildWebHook tries to trigger the provided webhook. It will attempt to utilize the current client
// configuration if the webhook has the same URL.
func (o *StartBuildOptions) RunStartBuildWebHook() error {
	repo := o.Git
	hook, err := url.Parse(o.FromWebhook)
	if err != nil {
		return err
	}

	event, err := hookEventFromPostReceive(repo, o.GitRepository, o.GitPostReceive)
	if err != nil {
		return err
	}

	// TODO: should be a versioned struct
	var data []byte
	if event != nil {
		data, err = json.Marshal(event)
		if err != nil {
			return err
		}
	}

	httpClient := http.DefaultClient
	// when using HTTPS, try to reuse the local config transport if possible to get a client cert
	// TODO: search all configs
	if hook.Scheme == "https" {
		config, err := o.ClientConfig.ClientConfig()
		if err == nil {
			if url, _, err := restclient.DefaultServerURL(config.Host, "", unversioned.GroupVersion{}, true); err == nil {
				if url.Host == hook.Host && url.Scheme == hook.Scheme {
					if rt, err := restclient.TransportFor(config); err == nil {
						httpClient = &http.Client{Transport: rt}
					}
				}
			}
		}
	}
	glog.V(4).Infof("Triggering hook %s\n%s", hook, string(data))
	resp, err := httpClient.Post(hook.String(), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == 301 || resp.StatusCode == 302:
		// TODO: follow redirect and display output
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server rejected our request %d\nremote: %s", resp.StatusCode, string(body))
	}
	return nil
}

// hookEventFromPostReceive creates a GenericWebHookEvent from the provided git repository and
// post receive input. If no inputs are available, it will return nil.
func hookEventFromPostReceive(repo git.Repository, path, postReceivePath string) (*buildapi.GenericWebHookEvent, error) {
	// TODO: support other types of refs
	event := &buildapi.GenericWebHookEvent{
		Git: &buildapi.GitInfo{},
	}

	// attempt to extract a post receive body
	refs := []git.ChangedRef{}
	switch receive := postReceivePath; {
	case receive == "-":
		r, err := git.ParsePostReceive(os.Stdin)
		if err != nil {
			return nil, err
		}
		refs = r
	case len(receive) > 0:
		file, err := os.Open(receive)
		if err != nil {
			return nil, fmt.Errorf("unable to open --git-post-receive argument as a file: %v", err)
		}
		defer file.Close()
		r, err := git.ParsePostReceive(file)
		if err != nil {
			return nil, err
		}
		refs = r
	}
	if len(refs) == 0 {
		return nil, nil
	}
	for _, ref := range refs {
		if len(ref.New) == 0 || ref.New == ref.Old {
			continue
		}
		info, err := gitRefInfo(repo, path, ref.New)
		if err != nil {
			glog.V(4).Infof("Could not retrieve info for %s:%s: %v", ref.Ref, ref.New, err)
		}
		info.Ref = ref.Ref
		info.Commit = ref.New
		event.Git.Refs = append(event.Git.Refs, info)
	}
	return event, nil
}

// gitRefInfo extracts a buildapi.GitRefInfo from the specified repository or returns
// an error.
func gitRefInfo(repo git.Repository, dir, ref string) (buildapi.GitRefInfo, error) {
	info := buildapi.GitRefInfo{}
	if repo == nil {
		return info, nil
	}
	out, err := repo.ShowFormat(dir, ref, "%H%n%an%n%ae%n%cn%n%ce%n%B")
	if err != nil {
		return info, err
	}
	lines := strings.SplitN(out, "\n", 6)
	if len(lines) != 6 {
		full := make([]string, 6)
		copy(full, lines)
		lines = full
	}
	info.Commit = lines[0]
	info.Author.Name = lines[1]
	info.Author.Email = lines[2]
	info.Committer.Name = lines[3]
	info.Committer.Email = lines[4]
	info.Message = lines[5]
	return info, nil
}

// WaitForBuildComplete waits for a build identified by the name to complete
func WaitForBuildComplete(c osclient.BuildInterface, name string) error {
	isOK := func(b *buildapi.Build) bool {
		return b.Status.Phase == buildapi.BuildPhaseComplete
	}
	isFailed := func(b *buildapi.Build) bool {
		return b.Status.Phase == buildapi.BuildPhaseFailed ||
			b.Status.Phase == buildapi.BuildPhaseCancelled ||
			b.Status.Phase == buildapi.BuildPhaseError
	}
	for {
		list, err := c.List(kapi.ListOptions{FieldSelector: fields.Set{"name": name}.AsSelector()})
		if err != nil {
			return err
		}
		for i := range list.Items {
			if name == list.Items[i].Name && isOK(&list.Items[i]) {
				return nil
			}
			if name != list.Items[i].Name || isFailed(&list.Items[i]) {
				return fmt.Errorf("the build %s/%s status is %q", list.Items[i].Namespace, list.Items[i].Name, list.Items[i].Status.Phase)
			}
		}

		rv := list.ResourceVersion
		w, err := c.Watch(kapi.ListOptions{FieldSelector: fields.Set{"name": name}.AsSelector(), ResourceVersion: rv})
		if err != nil {
			return err
		}
		defer w.Stop()

		for {
			val, ok := <-w.ResultChan()
			if !ok {
				// reget and re-watch
				break
			}
			if e, ok := val.Object.(*buildapi.Build); ok {
				if name == e.Name && isOK(e) {
					return nil
				}
				if name != e.Name || isFailed(e) {
					return fmt.Errorf("The build %s/%s status is %q", e.Namespace, name, e.Status.Phase)
				}
			}
		}
	}
}

func isInvalidSourceInputsError(err error) bool {
	if err != nil {
		if statusErr, ok := err.(*kerrors.StatusError); ok {
			if kerrors.IsInvalid(statusErr) {
				for _, cause := range statusErr.ErrStatus.Details.Causes {
					if cause.Field == "spec.source" {
						return true
					}
				}
			}
		}
	}
	return false
}
