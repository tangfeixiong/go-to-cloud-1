/*
 Copyright 2016, All rights reserved.

 Author <tangfx128@gmail.com>
*/

package main

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/tangfeixiong/go-to-cloud-1/cmd/cib/app"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())

	basename := filepath.Base(os.Args[0])
	app.Start(basename)
	//if err := app.Start(basename); err != nil {
	//	os.Exit(1)
	//}
}
