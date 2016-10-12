// Maintainer tangfeixiong <tangfx128@gmail.com>
//
// c5 means:
//   Cloud
//   Container
//   Controller
//   Command
//   Client

package main

import (
	"os"
	"path/filepath"
	"runtime"

	//"github.com/openshift/origin/pkg/cmd/cli"
	"github.com/openshift/origin/pkg/cmd/util/serviceability"

	// install all APIs
	_ "github.com/openshift/origin/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/api/install"
	_ "k8s.io/kubernetes/pkg/apis/autoscaling/install"
	_ "k8s.io/kubernetes/pkg/apis/batch/install"
	_ "k8s.io/kubernetes/pkg/apis/extensions/install"

	cli "github.com/tangfeixiong/go-to-cloud-1/pkg/c5ctl"
)

func main() {
	defer serviceability.BehaviorOnPanic(os.Getenv("OPENSHIFT_ON_PANIC"))()
	defer serviceability.Profile(os.Getenv("OPENSHIFT_PROFILE")).Stop()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	basename := filepath.Base(os.Args[0])
	command := cli.CommandFor(basename)
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
