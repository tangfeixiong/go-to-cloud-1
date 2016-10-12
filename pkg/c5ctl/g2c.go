package c5ctl

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

func AdditionalCLI() *cobra.Command {
	var g2c = &cobra.Command{
		Use:   "g2c (run-build | install | get | list ...)",
		Short: "go-to-cloud-1 commands",
		Long: `g2c is a Kubernetes/Openshift/Helm/... production manager.

The 'install' command will let you install G2C Charts into PaaS:

	$ c5 g2c install

Cheers.

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

		//Run: func(cmd *cobra.Command, args []string) { },
	}

	g2c.AddCommand(newRunBuildCmd(g2c.Out()))
	return g2c
}

func newRunBuildCmd(out io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:   "run-build (PATH | URL...)",
		Short: "run build onto Openshift",
		Long:  `Tell gRPC service to run build according arguments`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.Out(), "You can play it soon...")
		},
	}
}
