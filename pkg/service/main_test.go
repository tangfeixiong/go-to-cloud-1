package service

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	var loglevel = flag.Int("loglevel", 5, "loglevel binding with glog v")
	flag.Parse()
	f := flag.Lookup("v")
	if f != nil {
		f.Value.Set(strconv.Itoa(*loglevel))
	}

	if len(os.Args) > 0 {
		fmt.Printf("Reserved for running test by args: %+v", os.Args)
	}

	ret := m.Run()

	os.Exit(ret)
}
