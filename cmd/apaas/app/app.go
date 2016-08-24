package app

import (
	"fmt"
	"net"
	"os"

	"github.com/openshift/origin/pkg/cmd/flagtypes"
	"github.com/spf13/cobra"

	"google.golang.org/grpc"

	// "github.com/tangfeixiong/go-to-cloud-1/cmd/apaas/app/flagtypes"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/server"
)

func init() {
	_ = server.AppServer.GRPCServer(rootServer)
}

func Start(basename string) error {
	pf := rootCommand.PersistentFlags()
	pf.StringVarP(&addr, "listen", "l", ":50051", "The address:port to listen on")
	pf.StringVarP(&namespace, "namespace", "n", "", "The namespace Tiller calls home")

	flagtypes.GLog(pf)
	rootCommand.Use = "apaas"
	rootCommand.Short = "Docker image and ACI image build server"
	rootCommand.Long = `Docker build images via OpenShift and Kubernetes`
	rootCommand.Run = func(c *cobra.Command, args []string) {
		server.Run()
	}

	return rootCommand.Execute()
}

// rootServer is the root gRPC server.
//
// Each gRPC service registers itself to this server during init().
var rootServer = grpc.NewServer()

var (
	addr      = ":44134"
	namespace = ""
)

const globalUsage = `The Kubernetes Helm server.

Tiller is the server for Helm. It provides in-cluster resource management.

By default, Tiller listens for gRPC connections on port 44134.
`

var rootCommand = &cobra.Command{
	Use:   "tiller",
	Short: "The Kubernetes Helm server.",
	Long:  globalUsage,
	Run:   start,
}

func start(c *cobra.Command, args []string) {
	//setNamespace()
	lstn, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Tiller is running on %s\n", addr)

	if err := rootServer.Serve(lstn); err != nil {
		fmt.Fprintf(os.Stderr, "Server died: %s\n", err)
		os.Exit(1)
	}
}
