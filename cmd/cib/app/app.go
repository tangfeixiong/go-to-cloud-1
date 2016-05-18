
package app

import (
	"fmt"
	"os"

    "github.com/spf13/cobra"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/servers"
    
)


const globalUsage = `The Docker image and ACI building server.

Cib is the server for OCI and ACI. It provides in-Kubernetes budiling.

By default, it listens for REST-JSON connections on port 8080.
`

var rootCommand = &cobra.Command{
	Use:   "apiserver",
	Short: "The Docker and ACI image building server.",
	Long:  globalUsage,
	Run:   start,
}


func Start(basename string) error {
    return rootCommand.Execute()
}

func start(c *cobra.Command, args []string) {
	s := servers.NewApiServer()
	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
}