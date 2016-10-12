package c5ctl

import (
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/openshift/origin/pkg/cmd/templates"
)

// CommandFor returns the appropriate command for this base name,
// or the OpenShift CLI command.
func CommandFor(basename string) *cobra.Command {
	var cmd *cobra.Command

	in, out, errout := os.Stdin, os.Stdout, os.Stderr

	// Make case-insensitive and strip executable suffix if present
	if runtime.GOOS == "windows" {
		basename = strings.ToLower(basename)
		basename = strings.TrimSuffix(basename, ".exe")
	}

	switch basename {
	case "kubectl":
		cmd = cli. /*tangfeixiong*/ NewCmdKubectl(basename, out)
	default:
		cmd = cli. /*tangfeixiong*/ NewCommandCLI(basename, basename, in, out, errout)

		g2c := AdditionalCLI()
		cmd.AddCommand(g2c)
		templates.ActsAsRootCommand(cmd, []string{"g2c"}).
			ExposeFlags(g2c, "grpc-server", "grpc-gateway", "http-server")
	}

	if cmd.UsageFunc() == nil {
		templates.ActsAsRootCommand(cmd, []string{"options"})
	}
	flagtypes.GLog(cmd.PersistentFlags())

	return cmd
}
