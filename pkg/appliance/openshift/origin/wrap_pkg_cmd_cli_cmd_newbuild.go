package origin

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	kapi "k8s.io/kubernetes/pkg/api"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/runtime"
	"k8s.io/kubernetes/pkg/util/errors"

	buildapi "github.com/openshift/origin/pkg/build/api"
	//buildapiv1 "github.com/openshift/origin/pkg/build/api"
	cmdclicmd "github.com/openshift/origin/pkg/cmd/cli/cmd"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	configcmd "github.com/openshift/origin/pkg/config/cmd"
	newapp "github.com/openshift/origin/pkg/generate/app"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
)

type OsoNewBuildCommand struct {
	NewBuildCmd *cobra.Command
	*cmdclicmd.NewBuildOptions
}

func newOsoNewBuildCommand() *OsoNewBuildCommand {
	fullName := "oc"
	f := NewClientCmdFactory()
	in := os.Stdin
	out := os.Stdout

	c, o := NewCmdNewBuild(fullName, f, in, out)
	cmd := &OsoNewBuildCommand{NewBuildCmd: c, NewBuildOptions: &(o.NewBuildOptions)}

	overrideStringFlag(cmd.NewBuildCmd, "output", _output, "", "output formt", _output)
	overrideStringFlag(cmd.NewBuildCmd, "output-version", _output_version, "", "output version", _output_version)
	/*overrideStringFlag(cmd.NewBuildCmd, "template", _template, "", "output template", _template)
	overrideStringFlag(cmd.NewBuildCmd, "sort-by", _sort_by, "", "sort output", _sort_by)
	overrideStringSliceFlag(cmd.NewBuildCmd, "label-columns", _label_columns, "", "output columns", _label_columns)
	overrideBoolFlag(cmd.NewBuildCmd, "no-headers", _no_headers, "", "output table format", _no_headers)
	overrideBoolFlag(cmd.NewBuildCmd, "shall-all", _show_all, "", "output all", _show_all)
	overrideBoolFlag(cmd.NewBuildCmd, "shall-labels", _show_labels, "", "output more", _show_labels)
	overrideBoolFlag(cmd.NewBuildCmd, "watch", _is_watch, "", "long polling", _is_watch)
	overrideBoolFlag(cmd.NewBuildCmd, "watch-only", _is_watch, "", "long exec", _is_watch)*/

	return cmd
}

func (nb *OsoNewBuildCommand) Execute(args []string, project string, out io.Writer, in io.Reader) error {
	logger.SetPrefix("[openshift/origin, OsoNewBuildCommand.Execute] ")

	options := nb.NewBuildOptions
	fullName := os.Args[0]
	f := NewClientCmdFactory()
	c := nb.NewBuildCmd
	//c.SetArgs(args)

	logger.Println("options.Complete")
	if err := options.Complete(fullName, f, c, args, out, in); err != nil {
		kcmdutil.CheckErr(err)
		return err
	}
	logger.Printf("options: %+v\n", options)
	logger.Printf("options.AppConfig: %+v\n", options.Config)
	logger.Printf("options.Action: %+v\n", options.Action)
	if len(project) > 0 {
		options.Config.OriginNamespace = project
	}
	logger.Println("options.Run")
	err := options.Run()
	/*if err == cmdutil.ErrExit {
		os.Exit(1)
	}*/
	kcmdutil.CheckErr(err)
	return err
}

type NewBuildOptions struct {
	cmdclicmd.NewBuildOptions
	In   io.Reader
	PaaS *PaaS
}

/*type NewBuildOptions struct {
	Action configcmd.BulkAction
	Config *newcmd.AppConfig

	CommandPath string
	CommandName string

	Out, ErrOut   io.Writer
	Output        string
	PrintObject   func(obj runtime.Object) error
	LogsForObject LogsForObjectFunc
}*/

// NewCmdNewBuild implements the OpenShift cli new-build command
func NewCmdNewBuild(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) (*cobra.Command, *NewBuildOptions) {
	config := newcmd.NewAppConfig()
	config.ExpectToBuild = true
	config.AddEnvironmentToBuild = true
	//options := &cmdclicmd.NewBuildOptions{Config: config}
	options := &NewBuildOptions{
		NewBuildOptions: cmdclicmd.NewBuildOptions{Config: config},
	}

	cmd := &cobra.Command{
		Use:        "new-build (IMAGE | IMAGESTREAM | PATH | URL ...)",
		Short:      "Create a new build configuration",
		Long:       fmt.Sprintf(newBuildLong, fullName),
		Example:    fmt.Sprintf(newBuildExample, fullName),
		SuggestFor: []string{"build", "builds"},
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete(fullName, f, c, args, out, in))
			err := options.Run()
			if err == cmdutil.ErrExit {
				os.Exit(1)
			}
			kcmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringSliceVar(&config.SourceRepositories, "code", config.SourceRepositories, "Source code in the build configuration.")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image", "", config.ImageStreams, "Name of an image stream to to use as a builder. (deprecated)")
	cmd.Flags().MarkDeprecated("image", "use --image-stream instead")
	cmd.Flags().StringSliceVarP(&config.ImageStreams, "image-stream", "i", config.ImageStreams, "Name of an image stream to to use as a builder.")
	cmd.Flags().StringSliceVar(&config.DockerImages, "docker-image", config.DockerImages, "Name of a Docker image to use as a builder.")
	cmd.Flags().StringSliceVar(&config.Secrets, "build-secret", config.Secrets, "Secret and destination to use as an input for the build.")
	cmd.Flags().StringVar(&config.Name, "name", "", "Set name to use for generated build artifacts.")
	cmd.Flags().StringVar(&config.To, "to", "", "Push built images to this image stream tag (or Docker image repository if --to-docker is set).")
	cmd.Flags().BoolVar(&config.OutputDocker, "to-docker", false, "Have the build output push to a Docker repository.")
	cmd.Flags().StringSliceVarP(&config.Environment, "env", "e", config.Environment, "Specify key value pairs of environment variables to set into resulting image.")
	cmd.Flags().StringVar(&config.Strategy, "strategy", "", "Specify the build strategy to use if you don't want to detect (docker|source).")
	cmd.Flags().StringVarP(&config.Dockerfile, "dockerfile", "D", "", "Specify the contents of a Dockerfile to build directly, implies --strategy=docker. Pass '-' to read from STDIN.")
	cmd.Flags().BoolVar(&config.BinaryBuild, "binary", false, "Instead of expecting a source URL, set the build to expect binary contents. Will disable triggers.")
	cmd.Flags().StringP("labels", "l", "", "Label to set in all generated resources.")
	cmd.Flags().BoolVar(&config.AllowMissingImages, "allow-missing-images", false, "If true, indicates that referenced Docker images that cannot be found locally or in a registry should still be used.")
	cmd.Flags().BoolVar(&config.AllowMissingImageStreamTags, "allow-missing-imagestream-tags", false, "If true, indicates that image stream tags that don't exist should still be used.")
	cmd.Flags().StringVar(&config.ContextDir, "context-dir", "", "Context directory to be used for the build.")
	cmd.Flags().BoolVar(&config.NoOutput, "no-output", false, "If true, the build output will not be pushed anywhere.")
	cmd.Flags().StringVar(&config.SourceImage, "source-image", "", "Specify an image to use as source for the build.  You must also specify --source-image-path.")
	cmd.Flags().StringVar(&config.SourceImagePath, "source-image-path", "", "Specify the file or directory to copy from the source image and its destination in the build directory. Format: [source]:[destination-dir].")

	options.Action.BindForOutput(cmd.Flags())
	cmd.Flags().String("output-version", "", "The preferred API versions of the output objects")

	return cmd, options
}

// Complete sets any default behavior for the command
func (o *NewBuildOptions) Complete(fullName string, f *clientcmd.Factory, c *cobra.Command, args []string, out io.Writer, in io.Reader) error {
	o.Out = out
	o.ErrOut = c.Out()
	o.Output = kcmdutil.GetFlagString(c, "output")
	// Only output="" should print descriptions of intermediate steps. Everything
	// else should print only some specific output (json, yaml, go-template, ...)
	if len(o.Output) == 0 {
		o.Config.Out = o.Out
	} else {
		o.Config.Out = ioutil.Discard
	}
	o.Config.ErrOut = o.ErrOut

	o.Action.Out, o.Action.ErrOut = o.Out, o.ErrOut
	o.Action.Bulk.Mapper = clientcmd.ResourceMapper(f)
	o.Action.Bulk.Op = configcmd.Create //// tangfeixiong: .........................actual API method................
	// Retry is used to support previous versions of the API server that will
	// consider the presence of an unknown trigger type to be an error.
	o.Action.Bulk.Retry = retryBuildConfig

	o.Config.DryRun = o.Action.DryRun

	o.CommandPath = c.CommandPath()
	o.CommandName = fullName
	mapper, _ := f.Object(false)
	o.PrintObject = cmdutil.VersionedPrintObject(f.PrintObject, c, mapper, out)
	o.LogsForObject = f.LogsForObject
	if err := CompleteAppConfig(o.Config, f, c, args); err != nil {
		return err
	}
	if o.Config.Dockerfile == "-" {
		data, err := ioutil.ReadAll(in)
		if err != nil {
			return err
		}
		o.Config.Dockerfile = string(data)
	}
	if err := setAppConfigLabels(c, o.Config); err != nil {
		return err
	}
	return nil
}

func (o *NewBuildOptions) complete(fullName string, f *clientcmd.Factory, c *cobra.Command, args []string, out io.Writer, in io.Reader) error {
	o.Config.Out = o.Out
	o.Config.ErrOut = o.ErrOut

	o.Action.Out, o.Action.ErrOut = o.Out, o.ErrOut
	o.Action.Bulk.Mapper = clientcmd.ResourceMapper(f)
	o.Action.Bulk.Op = configcmd.Create //// tangfeixiong: .........................actual API method................
	// Retry is used to support previous versions of the API server that will
	// consider the presence of an unknown trigger type to be an error.
	o.Action.Bulk.Retry = retryBuildConfig

	o.Config.DryRun = o.Action.DryRun

	o.CommandPath = "osoc new-build"
	mapper, _ := f.Object(false)
	o.PrintObject = cmdutil.VersionedPrintObject(f.PrintObject, c, mapper, out)
	o.CommandName = fullName

	o.LogsForObject = f.LogsForObject

	o.Config.AllowMissingImages = false
	o.Config.AsSearch = false
	o.Config.SourceImage = ""
	o.Config.SourceImagePath = ""
	if err := CompleteAppConfig(o.Config, f, c, args); err != nil {
		return err
	}
	return nil
}

func (o *NewBuildOptions) runBuildConfig(bc *buildapi.BuildConfig, buildName string) error {
	config := o.Config
	out := o.Out

	checkGitInstalled(out)

	c := o.Config
	installables := []runtime.Object{bc}
	name := buildName
	result := &newcmd.AppResult{
		List:      &kapi.List{Items: installables},
		Name:      name,
		Namespace: c.OriginNamespace,

		GeneratedJobs: true,
	}

	if len(config.Labels) == 0 && len(result.Name) > 0 {
		config.Labels = map[string]string{"build": result.Name}
	}

	if err := setLabels(config.Labels, result, false); err != nil {
		return err
	}
	if err := setAnnotations(map[string]string{newcmd.GeneratedByNamespace: newcmd.GeneratedByNewBuild}, result); err != nil {
		return err
	}

	if o.Action.ShouldPrint() {
		return o.PrintObject(result.List)
	}

	//// tangfeixiong ......actual.run......func (b *BulkAction) Run(list *kapi.List, namespace string) []error
	if errs := o.Action.WithMessage(configcmd.CreateMessage(config.Labels), "created").Run(result.List, result.Namespace); len(errs) > 0 {
		return cmdutil.ErrExit
	}

	if !o.Action.Verbose() || o.Action.DryRun {
		return nil
	}

	indent := o.Action.DefaultIndent()
	for _, item := range result.List.Items {
		switch t := item.(type) {
		case *buildapi.BuildConfig: //// tangfeixiong: because ConfigChange Trigger
			if len(t.Spec.Triggers) > 0 && t.Spec.Source.Binary == nil {
				fmt.Fprintf(out, "%sBuild configuration %q created and build triggered.\n", indent, t.Name)
				fmt.Fprintf(out, "%sRun '%s logs -f bc/%s' to stream the build progress.\n", indent, o.CommandName, t.Name)
			}
		}
	}

	return nil
}

// Run contains all the necessary functionality for the OpenShift cli new-build command
func (o *NewBuildOptions) Run() error {
	config := o.Config
	out := o.Out

	checkGitInstalled(out)

	result, err := config.Run()
	if err != nil {
		return handleBuildError(err, o.CommandName, o.CommandPath)
	}

	if len(config.Labels) == 0 && len(result.Name) > 0 {
		config.Labels = map[string]string{"build": result.Name}
	}

	if err := setLabels(config.Labels, result, false); err != nil {
		return err
	}
	if err := setAnnotations(map[string]string{newcmd.GeneratedByNamespace: newcmd.GeneratedByNewBuild}, result); err != nil {
		return err
	}

	if o.Action.ShouldPrint() {
		fmt.Printf("[cmd/cli/cmd/newbuild.go, NewBuildOptions.Run] result: %+v\n", result)
		return o.PrintObject(result.List)
	}

	//// tangfeixiong ......actual.run......func (b *BulkAction) Run(list *kapi.List, namespace string) []error
	if errs := o.Action.WithMessage(configcmd.CreateMessage(config.Labels), "created").Run(result.List, result.Namespace); len(errs) > 0 {
		return cmdutil.ErrExit
	}

	if !o.Action.Verbose() || o.Action.DryRun {
		return nil
	}

	indent := o.Action.DefaultIndent()
	for _, item := range result.List.Items {
		switch t := item.(type) {
		case *buildapi.BuildConfig:
			if len(t.Spec.Triggers) > 0 && t.Spec.Source.Binary == nil {
				fmt.Fprintf(out, "%sBuild configuration %q created and build triggered.\n", indent, t.Name)
				fmt.Fprintf(out, "%sRun '%s logs -f bc/%s' to stream the build progress.\n", indent, o.CommandName, t.Name)
			}
		}
	}

	return nil
}

func handleBuildError(err error, fullName, commandPath string) error {
	if err == nil {
		return nil
	}
	errs := []error{err}
	if agg, ok := err.(errors.Aggregate); ok {
		errs = agg.Errors()
	}
	groups := errorGroups{}
	for _, err := range errs {
		transformBuildError(err, fullName, commandPath, groups)
	}
	buf := &bytes.Buffer{}
	for _, group := range groups {
		fmt.Fprint(buf, kcmdutil.MultipleErrors("error: ", group.errs))
		if len(group.suggestion) > 0 {
			fmt.Fprintln(buf)
		}
		fmt.Fprint(buf, group.suggestion)
	}
	return fmt.Errorf(buf.String())
}

func transformBuildError(err error, fullName, commandPath string, groups errorGroups) {
	switch t := err.(type) {
	case newapp.ErrNoMatch:
		groups.Add(
			"no-matches",
			heredoc.Docf(`
				The '%[1]s' command will match arguments to the following types:

				  1. Images tagged into image streams in the current project or the 'openshift' project
				     - if you don't specify a tag, we'll add ':latest'
				  2. Images in the Docker Hub, on remote registries, or on the local Docker engine
				  3. Git repository URLs or local paths that point to Git repositories

				--allow-missing-images can be used to force the use of an image that was not matched

				See '%[1]s -h' for examples.`, commandPath,
			),
			t,
			t.Errs...,
		)
		return
	}
	transformError(err, fullName, commandPath, groups)
}
