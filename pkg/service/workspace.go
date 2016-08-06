package service

import (
	"golang.org/x/net/context"

	_ "google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

func (u *UserResource) EnterWorkspace(ctx context.Context,
	in *osobp3.EnterWorkspaceRequest) (*osopb3.EnterWorkspaceResponse, error) {
	return nil, errNotImplemented
}

func (u *UserResource) LeaveWorkspace(ctx context.Context,
	in *osopb3.LeaveWorkspaceRequest) (*osopb3.LeaveWorkspaceResponse, error) {
	return nil, errNotImplemented
}
