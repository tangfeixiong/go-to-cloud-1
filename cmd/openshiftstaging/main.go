package main

import (
    
    "os"

	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
    
    "github.com/spf13/cobra"
    "github.com/spf13/pflag"    
)

func main() {
    command := oc()
    if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}

func oc() *cobra.Command {
    return cli.CommandFor("oc")
}

func kubectl() *cobra.Command {
    return cli.CommandFor("kubectl")
}

func login() {
    f := clientcmd.New(pflag.CommandLine)
    //f := clientcmd.New(cmds.PersistentFlags())
    
    command := cmd.NewCmdLogin("login", f, os.Stdin, os.Stdout)
}