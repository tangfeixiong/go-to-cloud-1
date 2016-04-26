/*
 Copyright 2016, All rights reserved.
 
 Author <tangfx128@gmail.com>
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/tangfeixiong/go-to-cloud-1/cmd/imgb/app/server"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	s := server.NewApiServer()

	if err := s.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
