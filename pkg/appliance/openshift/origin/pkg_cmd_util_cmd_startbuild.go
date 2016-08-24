package origin

import (
	"fmt"
	"io"

	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	"github.com/spf13/cobra"

	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func NewCmdStartBuild(fullName string, f *clientcmd.Factory, in io.Reader, out io.Writer) *cobra.Command {
	o := &cmd.StartBuildOptions{
		LogLevel:        "5",
		Follow:          true,
		WaitForComplete: false,
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
	return cmd

}
