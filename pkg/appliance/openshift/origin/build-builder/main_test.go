package builder

import (
	"fmt"
	"os"
	"testing"
)

var (
	_nats_addrs    []string = []string{"10.3.0.39:4222"}
	_nats_user     string   = "derek"
	_nats_password string   = "T0pS3cr3t"

	fake_username    string = "system:admin"
	fake_projectname string = "tangfx"
)

func TestMain(m *testing.M) {
	_ = flag.Int("loglevel", 5, "loglevel binding with glog v")
	flag.Parse()
	f := flag.Lookup("v")
	if f != nil {
		f.Value.Set("2")
	}

	if len(os.Args) > 0 {
		fmt.Printf("Reserved for running test by args: %+v", os.Args)
	}

	ret := m.Run()

	os.Exit(ret)
}
