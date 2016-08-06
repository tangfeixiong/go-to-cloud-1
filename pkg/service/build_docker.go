package service

import (
	github_com_openshift_origin_pkg_build_api_v1 "github.com/openshift/origin/pkg/build/api/v1"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/tangfeixiong/go-to-cloud-1/pkg/api/proto/paas/ci/osopb3"
)

func (u *UserResource) NewOsoBuildConfig(ctx context.Context,
	in *github_com_openshift_origin_pkg_build_api_v1.BuildConfig,
	opts ...grpc.CallOption) (*github_com_openshift_origin_pkg_build_api_v1.BuildConfig, error) {
	return nil, errNotImplemented
}

func (u *UserResource) StartOsoBuild(ctx context.Context,
	in *github_com_openshift_origin_pkg_build_api_v1.Build,
	opts ...grpc.CallOption) (*github_com_openshift_origin_pkg_build_api_v1.Build, error) {
	return nil, errNotImplemented
}

func (u *UserResource) BuildDockerImage(ctx context.Context,
	in *osopb3.RawData,
	opts ...grpc.CallOption) (*osopb3.RawData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) RebuildDockerImage(ctx context.Context,
	in *osopb3.RawData,
	opts ...grpc.CallOption) (*osopb3.RawData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) CreateDockerImageBuild(ctx context.Context,
	in *osopb3.DockerBuildRequestData,
	opts ...grpc.CallOption) (*osopb3.DockerBuildResponseData, error) {
	return nil, errNotImplemented
}

func (u *UserResource) UpdateDockerImageBuild(ctx context.Context,
	in *osopb3.DockerBuildRequestData,
	opts ...grpc.CallOption) (*osopb3.DockerBuildResponseData, error) {
	return nil, errNotImplemented
}
