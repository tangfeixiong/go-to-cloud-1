package utility

import (
	"github.com/spf13/cobra"
)

const globalUsage = `The Openshift Origin proxy server.

osoc is the server for Openshift Origin. It can provides in-cluster resource management.

By default, osoc listens for gRPC connections on port 50051.
`

var (
	RootCommand *cobra.Command = &cobra.Command{
		Use:   "osoc",
		Short: "openshift origin client.",
		Long:  globalUsage,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)
