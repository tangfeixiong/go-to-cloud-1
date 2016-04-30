package main

import (
	"flag"
	"log"
	"os"

	"github.com/golang/glog"
	"github.com/tangfeixiong/go-to-cloud-1/cmd/dockerbuilder/builder"
)

func init() {
	from := flag.CommandLine
	if lf := from.Lookup("v"); lf != nil {
		level := lf.Value.(*glog.Level)
		levelPtr := (*int32)(level)
		log.New(os.Stdout, "[main] ", log.LstdFlags|log.Lshortfile).Printf("init level: %+v\n", *levelPtr)
		//flag.Var(levelPtr, "loglevel", "Set the level of log output (0-9)")
	}
	flag.Parse()
	glog.V(5).Infof("glog level: %#v", flag.Lookup("v").Value.String())
}

func main() {
	builder.RunDockerBuild()
}
