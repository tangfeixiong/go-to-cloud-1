package origin

import (
	"flag"
	"os"
	"testing"

	"github.com/golang/glog"

	//"github.com/tangfeixiong/go-to-cloud-1/pkg/logger"
)

func TestMain(m *testing.M) {
	flag.Parse()
	f := flag.Lookup("v")
	if f != nil {
		f.Value.Set("10")
	}
	glog.V(10).Infoln("Set glog level with 10")

	ret := m.Run()

	os.Exit(ret)
}
