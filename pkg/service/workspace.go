package service

import (
	"golang.org/x/net/context"

	_ "google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/proto/api/paas/ci/openshift"
)

func (u *UserResource) EnterWorkspace(ctx context.Context, in *openshift.EnterWorkspaceRequest) (*openshift.EnterWorkspaceResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) LeaveWorkspace(ctx context.Context, in *openshift.LeaveWorkspaceRequest) (*openshift.LeaveWorkspaceResponse, error) {
	return nil, errNotImplemented
}
