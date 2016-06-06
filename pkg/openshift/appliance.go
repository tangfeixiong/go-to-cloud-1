package openshift

import (
	"errors"

	"github.com/golang/glog"

	"golang.org/x/net/context"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/api/paas/ci/openshift"
	"github.com/tangfeixiong/go-to-cloud-1/pkg/server"
)

type appliance struct {
}

var (
	errNotFound   error = errors.New("not found")
	errUnexpected error = errors.New("unexpected")

	Appliance appliance
)

func init() {
	openshift.RegisterSimpleServiceServer(server.ApiServer.GrpcRootServer, Appliance)
}

func (app *appliance) EnterWorkspace(context.Context, *openshift.EnterWorkspaceRequest) (*openshift.EnterWorkspaceResponse, error) {

}
