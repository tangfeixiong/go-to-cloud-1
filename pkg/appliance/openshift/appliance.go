package openshift

import (
	"errors"

	"github.com/golang/glog"

	"golang.org/x/net/context"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
	// "github.com/tangfeixiong/go-to-cloud-1/pkg/server"
)

var (
	errNotFound       error = errors.New("not found")
	errNotImplemented error = errors.New("not implemented")
	errUnexpected     error = errors.New("unexpected")
)

type Appliance struct {
}

func NewAppliance() *Appliance {
	return &Appliance{}
}

func (a *Appliance) EnterWorkspace(ctx context.Context,
	in *osopb3.EnterWorkspaceRequest) (*osopb3.EnterWorkspaceResponse, error) {
	glog.Info("grpc request")
	return nil, errNotImplemented
}
