package origin

import (
	"flag"
	"os"
	"testing"

	"github.com/golang/glog"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/logger"
)

var (
	fakeKubeconfigPath string = "/data/src/github.com/openshift/origin/etc/kubeconfig"
	fakeKubectlContext string = "openshift-origin-single"

	fakeOconfigPath string = "/data/src/github.com/openshift/origin/openshift.local.config/master/admin.kubeconfig"
	fakeOcContext   string = "default/172-17-4-50:30443/system:admin"

	fakeAdmin   string = "system:admin"
	fakeDefault string = "default"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if f := flag.Lookup("v"); f != nil {
		f.Value.Set("10")
		glog.V(10).Infoln("Set glog level with 10")
	} else {
		logger.SetPrefix("[openshift/origin, TestMain] ")
		logger.SetOutput(os.Stderr)
		logger.Println("glog not loaded")
	}

	ret := m.Run()

	os.Exit(ret)
}
