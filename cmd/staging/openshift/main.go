package main

import (
	"fmt"
	"io"
	"os"

	"github.com/golang/glog"

	"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/cli/cmd"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kapierrors "k8s.io/kubernetes/pkg/api/errors"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

func main() {
	var command *cobra.Command
	/*
		command = oc()
		if err := command.Execute(); err != nil {
			fmt.Fprintf(os.Stdout, "error: %+v", err)
			os.Exit(1)
		}
		f := clientcmd.New(command.PersistentFlags())
	*/
	f := clientcmd.New(pflag.CommandLine)

	command = NewCmdLogin("staging", f, os.Stdin, os.Stdout)
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stdout, "error: %+v", err)
		os.Exit(1)
	}

	/*
		command = newProject(f)
		if err := command.Execute(); err != nil {
			fmt.Fprintf(os.Stdout, "error: %+v", err)
			os.Exit(1)
		}
	*/
	fmt.Fprintf(os.Stdout, "%s", "done")
}

type stringValue struct {
	value string
}

func (val stringValue) String() string {
	return val.value
}
func (val stringValue) Set(v string) error {
	val.value = v
	return nil
}
func (val stringValue) Type() string {
	return "string"
}

func oc() *cobra.Command {
	return cli.CommandFor("oc")
}

func kubectl() *cobra.Command {
	return cli.CommandFor("kubectl")
}

func NewCmdLogin(fullName string, f *clientcmd.Factory, reader io.Reader, out io.Writer) *cobra.Command {
	//f := clientcmd.New(cmds.PersistentFlags())

	cmds := cmd.NewCmdLogin(fullName, f, reader, out)
	cmds.Run = func(cmd1 *cobra.Command, args []string) {
		options := &cmd.LoginOptions{
			Reader:   reader,
			Out:      out,
			Username: "tangfeixiong",
			Password: "tangfeixiong",
		}
		if f := cmd1.Flags().Lookup("username"); f != nil {
			if err := cmd1.Flags().Set("username", options.Username); err != nil {
				fmt.Fprintf(out, "[tangfx] flag username err: %+v\n", err)
				os.Exit(1)
			}
		} else {
			f = cmd1.Flags().VarPF(stringValue{options.Username}, "username", "", "")
			fmt.Fprintf(out, "[tangfx] flag username: %+v\n", f)
		}
		if f := cmd1.Flags().Lookup("password"); f != nil {
			if err := cmd1.Flags().Set("password", options.Password); err != nil {
				fmt.Fprintf(out, "[tangfx] flag password err: %+v\n", err)
				os.Exit(1)
			}
		} else {
			f = cmd1.Flags().VarPF(stringValue{options.Password}, "password", "", "")
			fmt.Fprintf(out, "[tangfx] flag password: %+v\n", f)
		}
		configPath := "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
		if f := cmd1.Flags().Lookup("config"); f != nil {
			if err := cmd1.Flags().Set("config", configPath); err != nil {
				fmt.Fprintf(out, "[tangfx] flag config err: %+v\n", err)
				os.Exit(1)
			}
		} else {
			f = cmd1.Flags().VarPF(stringValue{configPath}, "config", "", "")
			fmt.Fprintf(out, "[tangfx] flag config: %+v\n", f)
		}
		glog.Infof("[tangfx] login option: %+v\n", options)

		if err := options.Complete(f, cmd1, args); err != nil {
			kcmdutil.CheckErr(err)
		}

		if err := options.Validate(args, kcmdutil.GetFlagString(cmd1, "server")); err != nil {
			kcmdutil.CheckErr(err)
		}

		err := cmd.RunLogin(cmd1, options)

		if kapierrors.IsUnauthorized(err) {
			fmt.Fprintln(out, "Login failed (401 Unauthorized)")

			if err, isStatusErr := err.(*kapierrors.StatusError); isStatusErr {
				if details := err.Status().Details; details != nil {
					for _, cause := range details.Causes {
						fmt.Fprintln(out, cause.Message)
					}
				}
			}

			os.Exit(1)

		} else {
			kcmdutil.CheckErr(err)
		}
	}

	return cmds
}

func newProject(f *clientcmd.Factory) *cobra.Command {

	return cmd.NewCmdRequestProject("staging", "new-project", "staging login", "staging project", f, os.Stdout)
}
