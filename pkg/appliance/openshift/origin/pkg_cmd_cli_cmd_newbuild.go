package origin

import (
	//"flag"
	"fmt"
	"io"
	"os"

	command "github.com/openshift/origin/pkg/cmd/cli/cmd"
	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	//configcmd "github.com/openshift/origin/pkg/config/cmd"
	//newapp "github.com/openshift/origin/pkg/generate/app"
	newcmd "github.com/openshift/origin/pkg/generate/app/cmd"
	"github.com/spf13/cobra"

	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

type OsoNewBuildCommand struct {
	NewBuildCmd *cobra.Command
	*command.NewBuildOptions
}

func newOsoNewBuildCommand() *OsoNewBuildCommand {
	fullName := "oc"
	f := NewClientCmdFactory()
	in := os.Stdin
	out := os.Stdout

	c, o := NewCmdNewBuild(fullName, f, in, out)
	cmd := &OsoNewBuildCommand{NewBuildCmd: c, NewBuildOptions: o}

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

// NewCmdNewBuild implements the OpenShift cli new-build command
func NewCmdNewBuild(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) (*cobra.Command, *command.NewBuildOptions) {
	config := newcmd.NewAppConfig()
	config.ExpectToBuild = true
	config.AddEnvironmentToBuild = true
	options := &command.NewBuildOptions{Config: config}

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
