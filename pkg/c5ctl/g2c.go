package c5ctl

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/helm/helm-classic/codec"
	"github.com/spf13/cobra"

	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"

	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"

	c5cmd "github.com/tangfeixiong/go-to-cloud-1/pkg/c5ctl/cmd"
)

func AdditionalCLI() *cobra.Command {
	var g2c = &cobra.Command{
		Use:   "g2c (run-build | docker-build | source-build | list ...)",
		Short: "go-to-cloud-1 commands",
		Long: `g2c is a Kubernetes/Openshift/Helm/... production manager.

The 'docker-build' command will let you to build docker image at PaaS:

	$ c5 g2c docker-build

Cheers.

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,

		//Run: func(cmd *cobra.Command, args []string) { },
	}

	g2c.AddCommand(newRunBuildCmd(os.Stdout))
	g2c.AddCommand(newDockerBuildCmd(os.Stdout, os.Stdin))
	return g2c
}

func newDockerBuildCmd(out io.Writer, in io.Reader) *cobra.Command {
	type Options struct {
		filePath string
		values   map[string]interface{}
		Complete func(fullName string, f *clientcmd.Factory, c *cobra.Command, args []string, out io.Writer, in io.Reader) error
		Run      func() error
	}
	options := &Options{values: make(map[string]interface{})}
	options.Complete = func(fullName string, f *clientcmd.Factory, c *cobra.Command, args []string, out io.Writer, in io.Reader) error {
		var data []byte
		var err error
		var hco *codec.Object
		if options.filePath == "-" {
			data, err = ioutil.ReadAll(in)
			if err != nil {
				return err
			}
		} else {
			data, err = ioutil.ReadFile(options.filePath)
			if err != nil {
				return fmt.Errorf("Error reading file: %+v", err)
			}
		}
		hco, err = codec.JSON.Decode(data).One()
		if err != nil {
			return err
		}
		if err = hco.Object(&options.values); err != nil {
			return err
		}
		if len(options.values) == 0 {
			return fmt.Errorf("content required!")
		}
		return nil
	}
	options.Run = func() error {
		var status int
		var ok bool

		status, ok = c5cmd.DockerBuildWithValues(options.values)
		if !ok || status == 0 {
			fmt.Fprintln(out, "create failed!")
			return cmdutil.ErrExit
		}
		fmt.Fprintf(out, "status: %+v\n", status)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				status, ok = c5cmd.TrackingDockerBuild(options.values)
				if !ok || status == 0 {
					fmt.Fprintln(out, "tracking failed!")
					return
				}
				switch status {
				case 1:
					fmt.Fprintln(out, "Build is continuing")
				case 2:
					fmt.Fprintln(out, "Build is failure")
					return
				case 3:
					fmt.Fprintln(out, "Build is succeeded")
					return
				case 4:
					fmt.Fprintln(out, "Warning")
					return
				default:
					fmt.Fprintln(out, "Unexpected")
					return
				}
			}
		}()
		wg.Wait()
		return nil
	}

	cmd := &cobra.Command{
		Use:   "docker-build (PATH | URL...)",
		Short: "docker build into Openshift",
		Long:  `Tell gRPC service to run build according arguments`,
		Run: func(c *cobra.Command, args []string) {
			kcmdutil.CheckErr(options.Complete("c5", nil, c, args, out, in))
			err := options.Run()
			if err == cmdutil.ErrExit {
				os.Exit(1)
			}
			kcmdutil.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&options.filePath, "contextfile", "f", "", "Specify the contents of a file to build directly. Pass '-' to read from STDIN.")

	return cmd
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
