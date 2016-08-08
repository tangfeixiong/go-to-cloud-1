package service

import (
	"golang.org/x/net/context"
	//"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

func (u *UserResource) EnterWorkspace(ctx context.Context,
	in *osopb3.CreateOriginProjectArbitraryRequest) (*osopb3.CreateOriginProjectArbitraryResponse, error) {
	return nil, errNotImplemented
}
