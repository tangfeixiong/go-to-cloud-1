package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang/glog"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/appliance/openshift/origin/docker-builder"
)

func main() {
	from := flag.CommandLine
	lf := from.Lookup("v")
	if lf != nil {
		level := lf.Value.(*glog.Level)
		levelPtr := (*int32)(level)
		//flag.Var(levelPtr, "loglevel", "Set the level of log output (0-9)")
		log.New(os.Stdout, "[appliance/openshift/origin/tc-origin-docker-builder] ",
			log.LstdFlags|log.Lshortfile).Printf("glog level: %+v\n", *levelPtr)
	}
	flag.Parse()
	if lf != nil {
		glog.V(5).Infof("glog level: %#v", lf.Value.String())
	}

	builder.RunDockerBuild()
}
