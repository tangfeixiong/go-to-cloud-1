package service

import (
	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

func (u *UserResource) EnterWorkspace(ctx context.Context,
	in *osopb3.RawData) (*osopb3.RawData, error) {
	return nil, errNotImplemented
}
