// Code generated by protoc-gen-go.
// source: paas/ci/openshift/service.proto
// DO NOT EDIT!

/*
Package openshift is a generated protocol buffer package.

It is generated from these files:
	paas/ci/openshift/service.proto

It has these top-level messages:
	EnterWorkspaceRequest
	EnterWorkspaceResponse
	LeaveWorkspaceRequest
	LeaveWorkspaceResponse
	CreateProjectRequest
	CreateProjectResponse
	LookupProjectsRequest
	LookupProjectsResponse
	OpenProjectRequest
	OpenProjectResponse
	DeleteProjectRequest
	DeleteProjectResponse
	BuildDockerImageRequest
	GitRepo
*/
package openshift

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type EnterWorkspaceRequest struct {
	Username    string `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
	Credentials string `protobuf:"bytes,2,opt,name=credentials" json:"credentials,omitempty"`
}

func (m *EnterWorkspaceRequest) Reset()                    { *m = EnterWorkspaceRequest{} }
func (m *EnterWorkspaceRequest) String() string            { return proto.CompactTextString(m) }
func (*EnterWorkspaceRequest) ProtoMessage()               {}
func (*EnterWorkspaceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type EnterWorkspaceResponse struct {
	Identifier string `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
}

func (m *EnterWorkspaceResponse) Reset()                    { *m = EnterWorkspaceResponse{} }
func (m *EnterWorkspaceResponse) String() string            { return proto.CompactTextString(m) }
func (*EnterWorkspaceResponse) ProtoMessage()               {}
func (*EnterWorkspaceResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type LeaveWorkspaceRequest struct {
	Username   string `protobuf:"bytes,1,opt,name=username" json:"username,omitempty"`
	Identifier string `protobuf:"bytes,2,opt,name=identifier" json:"identifier,omitempty"`
}

func (m *LeaveWorkspaceRequest) Reset()                    { *m = LeaveWorkspaceRequest{} }
func (m *LeaveWorkspaceRequest) String() string            { return proto.CompactTextString(m) }
func (*LeaveWorkspaceRequest) ProtoMessage()               {}
func (*LeaveWorkspaceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type LeaveWorkspaceResponse struct {
	Flag int64 `protobuf:"varint,1,opt,name=flag" json:"flag,omitempty"`
}

func (m *LeaveWorkspaceResponse) Reset()                    { *m = LeaveWorkspaceResponse{} }
func (m *LeaveWorkspaceResponse) String() string            { return proto.CompactTextString(m) }
func (*LeaveWorkspaceResponse) ProtoMessage()               {}
func (*LeaveWorkspaceResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type CreateProjectRequest struct {
	Identifier  string `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
	ProjectID   string `protobuf:"bytes,2,opt,name=projectID" json:"projectID,omitempty"`
	ProjectName string `protobuf:"bytes,3,opt,name=projectName" json:"projectName,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description" json:"description,omitempty"`
}

func (m *CreateProjectRequest) Reset()                    { *m = CreateProjectRequest{} }
func (m *CreateProjectRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateProjectRequest) ProtoMessage()               {}
func (*CreateProjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type CreateProjectResponse struct {
	ApiVersion  string `protobuf:"bytes,1,opt,name=apiVersion" json:"apiVersion,omitempty"`
	Kind        string `protobuf:"bytes,2,opt,name=kind" json:"kind,omitempty"`
	ProjectJSON []byte `protobuf:"bytes,3,opt,name=projectJSON,proto3" json:"projectJSON,omitempty"`
}

func (m *CreateProjectResponse) Reset()                    { *m = CreateProjectResponse{} }
func (m *CreateProjectResponse) String() string            { return proto.CompactTextString(m) }
func (*CreateProjectResponse) ProtoMessage()               {}
func (*CreateProjectResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type LookupProjectsRequest struct {
	Identifier string `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
}

func (m *LookupProjectsRequest) Reset()                    { *m = LookupProjectsRequest{} }
func (m *LookupProjectsRequest) String() string            { return proto.CompactTextString(m) }
func (*LookupProjectsRequest) ProtoMessage()               {}
func (*LookupProjectsRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type LookupProjectsResponse struct {
	ApiVersion   string `protobuf:"bytes,1,opt,name=apiVersion" json:"apiVersion,omitempty"`
	Kind         string `protobuf:"bytes,2,opt,name=kind" json:"kind,omitempty"`
	ProjectsJSON []byte `protobuf:"bytes,3,opt,name=projectsJSON,proto3" json:"projectsJSON,omitempty"`
}

func (m *LookupProjectsResponse) Reset()                    { *m = LookupProjectsResponse{} }
func (m *LookupProjectsResponse) String() string            { return proto.CompactTextString(m) }
func (*LookupProjectsResponse) ProtoMessage()               {}
func (*LookupProjectsResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type OpenProjectRequest struct {
	Identifier string `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
	ProjectID  string `protobuf:"bytes,2,opt,name=projectID" json:"projectID,omitempty"`
}

func (m *OpenProjectRequest) Reset()                    { *m = OpenProjectRequest{} }
func (m *OpenProjectRequest) String() string            { return proto.CompactTextString(m) }
func (*OpenProjectRequest) ProtoMessage()               {}
func (*OpenProjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type OpenProjectResponse struct {
	ApiVersion  string `protobuf:"bytes,1,opt,name=apiVersion" json:"apiVersion,omitempty"`
	Kind        string `protobuf:"bytes,2,opt,name=kind" json:"kind,omitempty"`
	ProjectJSON []byte `protobuf:"bytes,3,opt,name=projectJSON,proto3" json:"projectJSON,omitempty"`
}

func (m *OpenProjectResponse) Reset()                    { *m = OpenProjectResponse{} }
func (m *OpenProjectResponse) String() string            { return proto.CompactTextString(m) }
func (*OpenProjectResponse) ProtoMessage()               {}
func (*OpenProjectResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type DeleteProjectRequest struct {
	Identifier string `protobuf:"bytes,1,opt,name=identifier" json:"identifier,omitempty"`
	ProjectID  string `protobuf:"bytes,2,opt,name=projectID" json:"projectID,omitempty"`
}

func (m *DeleteProjectRequest) Reset()                    { *m = DeleteProjectRequest{} }
func (m *DeleteProjectRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteProjectRequest) ProtoMessage()               {}
func (*DeleteProjectRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

type DeleteProjectResponse struct {
	Flag int64 `protobuf:"varint,1,opt,name=flag" json:"flag,omitempty"`
}

func (m *DeleteProjectResponse) Reset()                    { *m = DeleteProjectResponse{} }
func (m *DeleteProjectResponse) String() string            { return proto.CompactTextString(m) }
func (*DeleteProjectResponse) ProtoMessage()               {}
func (*DeleteProjectResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

type BuildDockerImageRequest struct {
	DockerfileContent string `protobuf:"bytes,1,opt,name=dockerfileContent" json:"dockerfileContent,omitempty"`
	ContextArchive    string `protobuf:"bytes,2,opt,name=contextArchive" json:"contextArchive,omitempty"`
	Gitrepo           string `protobuf:"bytes,3,opt,name=gitrepo" json:"gitrepo,omitempty"`
	BinaryStream      string `protobuf:"bytes,4,opt,name=binaryStream" json:"binaryStream,omitempty"`
}

func (m *BuildDockerImageRequest) Reset()                    { *m = BuildDockerImageRequest{} }
func (m *BuildDockerImageRequest) String() string            { return proto.CompactTextString(m) }
func (*BuildDockerImageRequest) ProtoMessage()               {}
func (*BuildDockerImageRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type GitRepo struct {
	Site        string `protobuf:"bytes,1,opt,name=site" json:"site,omitempty"`
	Uri         string `protobuf:"bytes,2,opt,name=uri" json:"uri,omitempty"`
	Ref         string `protobuf:"bytes,3,opt,name=ref" json:"ref,omitempty"`
	ContextPath string `protobuf:"bytes,4,opt,name=contextPath" json:"contextPath,omitempty"`
	DockerFile  string `protobuf:"bytes,5,opt,name=dockerFile" json:"dockerFile,omitempty"`
}

func (m *GitRepo) Reset()                    { *m = GitRepo{} }
func (m *GitRepo) String() string            { return proto.CompactTextString(m) }
func (*GitRepo) ProtoMessage()               {}
func (*GitRepo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func init() {
	proto.RegisterType((*EnterWorkspaceRequest)(nil), "paas.ci.openshift.service.EnterWorkspaceRequest")
	proto.RegisterType((*EnterWorkspaceResponse)(nil), "paas.ci.openshift.service.EnterWorkspaceResponse")
	proto.RegisterType((*LeaveWorkspaceRequest)(nil), "paas.ci.openshift.service.LeaveWorkspaceRequest")
	proto.RegisterType((*LeaveWorkspaceResponse)(nil), "paas.ci.openshift.service.LeaveWorkspaceResponse")
	proto.RegisterType((*CreateProjectRequest)(nil), "paas.ci.openshift.service.CreateProjectRequest")
	proto.RegisterType((*CreateProjectResponse)(nil), "paas.ci.openshift.service.CreateProjectResponse")
	proto.RegisterType((*LookupProjectsRequest)(nil), "paas.ci.openshift.service.LookupProjectsRequest")
	proto.RegisterType((*LookupProjectsResponse)(nil), "paas.ci.openshift.service.LookupProjectsResponse")
	proto.RegisterType((*OpenProjectRequest)(nil), "paas.ci.openshift.service.OpenProjectRequest")
	proto.RegisterType((*OpenProjectResponse)(nil), "paas.ci.openshift.service.OpenProjectResponse")
	proto.RegisterType((*DeleteProjectRequest)(nil), "paas.ci.openshift.service.DeleteProjectRequest")
	proto.RegisterType((*DeleteProjectResponse)(nil), "paas.ci.openshift.service.DeleteProjectResponse")
	proto.RegisterType((*BuildDockerImageRequest)(nil), "paas.ci.openshift.service.BuildDockerImageRequest")
	proto.RegisterType((*GitRepo)(nil), "paas.ci.openshift.service.GitRepo")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion2

// Client API for SimpleService service

type SimpleServiceClient interface {
	EnterWorkspace(ctx context.Context, in *EnterWorkspaceRequest, opts ...grpc.CallOption) (*EnterWorkspaceResponse, error)
	LeaveWorkspace(ctx context.Context, in *LeaveWorkspaceRequest, opts ...grpc.CallOption) (*LeaveWorkspaceResponse, error)
	CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*CreateProjectResponse, error)
	LookupProjects(ctx context.Context, in *LookupProjectsRequest, opts ...grpc.CallOption) (*LookupProjectsResponse, error)
	OpenProject(ctx context.Context, in *OpenProjectRequest, opts ...grpc.CallOption) (*OpenProjectResponse, error)
	DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*DeleteProjectResponse, error)
}

type simpleServiceClient struct {
	cc *grpc.ClientConn
}

func NewSimpleServiceClient(cc *grpc.ClientConn) SimpleServiceClient {
	return &simpleServiceClient{cc}
}

func (c *simpleServiceClient) EnterWorkspace(ctx context.Context, in *EnterWorkspaceRequest, opts ...grpc.CallOption) (*EnterWorkspaceResponse, error) {
	out := new(EnterWorkspaceResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/EnterWorkspace", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *simpleServiceClient) LeaveWorkspace(ctx context.Context, in *LeaveWorkspaceRequest, opts ...grpc.CallOption) (*LeaveWorkspaceResponse, error) {
	out := new(LeaveWorkspaceResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/LeaveWorkspace", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *simpleServiceClient) CreateProject(ctx context.Context, in *CreateProjectRequest, opts ...grpc.CallOption) (*CreateProjectResponse, error) {
	out := new(CreateProjectResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/CreateProject", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *simpleServiceClient) LookupProjects(ctx context.Context, in *LookupProjectsRequest, opts ...grpc.CallOption) (*LookupProjectsResponse, error) {
	out := new(LookupProjectsResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/LookupProjects", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *simpleServiceClient) OpenProject(ctx context.Context, in *OpenProjectRequest, opts ...grpc.CallOption) (*OpenProjectResponse, error) {
	out := new(OpenProjectResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/OpenProject", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *simpleServiceClient) DeleteProject(ctx context.Context, in *DeleteProjectRequest, opts ...grpc.CallOption) (*DeleteProjectResponse, error) {
	out := new(DeleteProjectResponse)
	err := grpc.Invoke(ctx, "/paas.ci.openshift.service.SimpleService/DeleteProject", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for SimpleService service

type SimpleServiceServer interface {
	EnterWorkspace(context.Context, *EnterWorkspaceRequest) (*EnterWorkspaceResponse, error)
	LeaveWorkspace(context.Context, *LeaveWorkspaceRequest) (*LeaveWorkspaceResponse, error)
	CreateProject(context.Context, *CreateProjectRequest) (*CreateProjectResponse, error)
	LookupProjects(context.Context, *LookupProjectsRequest) (*LookupProjectsResponse, error)
	OpenProject(context.Context, *OpenProjectRequest) (*OpenProjectResponse, error)
	DeleteProject(context.Context, *DeleteProjectRequest) (*DeleteProjectResponse, error)
}

func RegisterSimpleServiceServer(s *grpc.Server, srv SimpleServiceServer) {
	s.RegisterService(&_SimpleService_serviceDesc, srv)
}

func _SimpleService_EnterWorkspace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EnterWorkspaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).EnterWorkspace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/EnterWorkspace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).EnterWorkspace(ctx, req.(*EnterWorkspaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SimpleService_LeaveWorkspace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LeaveWorkspaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).LeaveWorkspace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/LeaveWorkspace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).LeaveWorkspace(ctx, req.(*LeaveWorkspaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SimpleService_CreateProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).CreateProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/CreateProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).CreateProject(ctx, req.(*CreateProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SimpleService_LookupProjects_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LookupProjectsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).LookupProjects(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/LookupProjects",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).LookupProjects(ctx, req.(*LookupProjectsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SimpleService_OpenProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OpenProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).OpenProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/OpenProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).OpenProject(ctx, req.(*OpenProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SimpleService_DeleteProject_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteProjectRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SimpleServiceServer).DeleteProject(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/paas.ci.openshift.service.SimpleService/DeleteProject",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SimpleServiceServer).DeleteProject(ctx, req.(*DeleteProjectRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _SimpleService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "paas.ci.openshift.service.SimpleService",
	HandlerType: (*SimpleServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "EnterWorkspace",
			Handler:    _SimpleService_EnterWorkspace_Handler,
		},
		{
			MethodName: "LeaveWorkspace",
			Handler:    _SimpleService_LeaveWorkspace_Handler,
		},
		{
			MethodName: "CreateProject",
			Handler:    _SimpleService_CreateProject_Handler,
		},
		{
			MethodName: "LookupProjects",
			Handler:    _SimpleService_LookupProjects_Handler,
		},
		{
			MethodName: "OpenProject",
			Handler:    _SimpleService_OpenProject_Handler,
		},
		{
			MethodName: "DeleteProject",
			Handler:    _SimpleService_DeleteProject_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}

var fileDescriptor0 = []byte{
	// 633 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xb4, 0x56, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x56, 0xfa, 0x43, 0xe9, 0xf4, 0x47, 0xb0, 0x34, 0x25, 0x44, 0x08, 0xaa, 0x3d, 0x20, 0x24,
	0x8a, 0x53, 0xe0, 0x00, 0x57, 0xda, 0x02, 0x2a, 0x42, 0x6d, 0x95, 0xf0, 0x23, 0x71, 0xdb, 0x3a,
	0x93, 0x66, 0x89, 0xe3, 0x5d, 0x76, 0xd7, 0x11, 0x9c, 0x79, 0x09, 0x5e, 0x82, 0x17, 0xe1, 0xa9,
	0xf0, 0xda, 0x9b, 0xc4, 0x76, 0x5c, 0x93, 0x8a, 0x72, 0x9b, 0xfd, 0x3c, 0x3b, 0xf3, 0xcd, 0xcc,
	0xce, 0x27, 0xc3, 0x7d, 0xc9, 0x98, 0x6e, 0xf9, 0xbc, 0x25, 0x24, 0x86, 0xba, 0xcf, 0x7b, 0xa6,
	0xa5, 0x51, 0x8d, 0xb8, 0x8f, 0x9e, 0x54, 0xc2, 0x08, 0x72, 0xc7, 0x3a, 0x78, 0x3e, 0xf7, 0x26,
	0x0e, 0x9e, 0x73, 0xa0, 0x1f, 0xa0, 0xfe, 0x2a, 0x34, 0xa8, 0x3e, 0x09, 0x35, 0xd0, 0x92, 0xf9,
	0xd8, 0xc6, 0xaf, 0x11, 0x6a, 0x43, 0x9a, 0x70, 0x3d, 0x8a, 0x9d, 0x42, 0x36, 0xc4, 0x46, 0x6d,
	0xa7, 0xf6, 0x70, 0xb5, 0x3d, 0x39, 0x93, 0x1d, 0x58, 0xf3, 0x15, 0x76, 0x31, 0x34, 0x9c, 0x05,
	0xba, 0xb1, 0x90, 0x7c, 0xce, 0x42, 0xf4, 0x05, 0x6c, 0x17, 0xc3, 0x6a, 0x29, 0x42, 0x8d, 0xe4,
	0x1e, 0x00, 0x4f, 0xdc, 0x7a, 0x1c, 0x95, 0x8b, 0x9c, 0x41, 0x68, 0x07, 0xea, 0xef, 0x90, 0x8d,
	0xf0, 0x52, 0x84, 0xf2, 0x41, 0x17, 0x66, 0x82, 0xee, 0xc2, 0x76, 0x31, 0xa8, 0xa3, 0x43, 0x60,
	0xa9, 0x17, 0xb0, 0xf3, 0x24, 0xe2, 0x62, 0x3b, 0xb1, 0xe9, 0xcf, 0x1a, 0x6c, 0x1d, 0x28, 0x64,
	0x06, 0x4f, 0x95, 0xf8, 0x82, 0xbe, 0x19, 0x53, 0xf8, 0x0b, 0x77, 0x72, 0x17, 0x56, 0x65, 0x7a,
	0xe3, 0xe8, 0xd0, 0xb1, 0x98, 0x02, 0xb6, 0x6b, 0xee, 0x70, 0x6c, 0x6b, 0x58, 0x4c, 0xbb, 0x96,
	0x81, 0xac, 0x47, 0x17, 0xb5, 0xaf, 0xb8, 0x34, 0x5c, 0x84, 0x8d, 0xa5, 0xd4, 0x23, 0x03, 0xd1,
	0x21, 0xd4, 0x0b, 0xcc, 0xa6, 0x6d, 0x65, 0x92, 0x7f, 0x44, 0xa5, 0xed, 0x4d, 0x47, 0x6d, 0x8a,
	0xd8, 0x3a, 0x07, 0x3c, 0xec, 0x3a, 0x56, 0x89, 0x9d, 0x21, 0xf4, 0xb6, 0x73, 0x72, 0x9c, 0x10,
	0x5a, 0x6f, 0x67, 0x21, 0xfa, 0x3c, 0x1e, 0x86, 0x10, 0x83, 0x48, 0xba, 0x74, 0x7a, 0xce, 0x4e,
	0x50, 0x19, 0x37, 0xbc, 0x70, 0xf1, 0x1f, 0x88, 0x52, 0x58, 0x77, 0xac, 0x74, 0x86, 0x69, 0x0e,
	0xa3, 0x6d, 0x20, 0x27, 0xf1, 0xeb, 0xbe, 0xca, 0x89, 0xd1, 0x01, 0xdc, 0xca, 0xc5, 0xfc, 0xaf,
	0xbd, 0x7e, 0x0f, 0x5b, 0x87, 0x18, 0xe0, 0xd5, 0x3e, 0x3a, 0xfa, 0x08, 0xea, 0x85, 0xa8, 0x15,
	0x0f, 0xff, 0x57, 0x0d, 0x6e, 0xef, 0x47, 0x3c, 0xe8, 0x1e, 0x0a, 0x7f, 0x80, 0xea, 0x68, 0xc8,
	0xce, 0x27, 0xeb, 0xb7, 0x0b, 0x37, 0xbb, 0x09, 0xda, 0xe3, 0x01, 0x1e, 0x88, 0x78, 0xb9, 0x43,
	0xe3, 0xd8, 0xcc, 0x7e, 0x20, 0x0f, 0x60, 0xd3, 0xb7, 0xe6, 0x37, 0xf3, 0x52, 0xf9, 0x7d, 0x3e,
	0x42, 0xc7, 0xac, 0x80, 0x92, 0x06, 0xac, 0x9c, 0x73, 0xa3, 0x50, 0x0a, 0xb7, 0x0f, 0xe3, 0xa3,
	0x9d, 0xf9, 0x19, 0x0f, 0x99, 0xfa, 0xde, 0x89, 0xcf, 0x6c, 0xe8, 0x96, 0x21, 0x87, 0xd1, 0x1f,
	0x35, 0x58, 0x79, 0xc3, 0xe3, 0x9a, 0x62, 0xff, 0xb8, 0x1e, 0xcd, 0xcd, 0x58, 0x1a, 0x12, 0x9b,
	0xdc, 0x80, 0xc5, 0x48, 0x71, 0x97, 0xda, 0x9a, 0x16, 0x51, 0xd8, 0x73, 0xb9, 0xac, 0x99, 0x68,
	0x59, 0xca, 0xe9, 0x94, 0x99, 0xfe, 0x78, 0xe7, 0x32, 0x90, 0x1d, 0x40, 0x5a, 0xe0, 0xeb, 0xb8,
	0xc0, 0xc6, 0x72, 0x3a, 0x80, 0x29, 0xf2, 0xf4, 0xf7, 0x32, 0x6c, 0x74, 0xf8, 0x50, 0x06, 0xd8,
	0x49, 0x45, 0x95, 0x44, 0xb0, 0x99, 0x57, 0x3f, 0xb2, 0xe7, 0x5d, 0x28, 0xc1, 0x5e, 0xa9, 0xfe,
	0x36, 0x9f, 0x5c, 0xe2, 0x86, 0x1b, 0x69, 0x9c, 0x36, 0xaf, 0x72, 0x95, 0x69, 0x4b, 0x55, 0xb6,
	0x32, 0xed, 0x05, 0x12, 0xaa, 0x60, 0x23, 0xa7, 0x49, 0xa4, 0x55, 0x11, 0xa3, 0x4c, 0x57, 0x9b,
	0x7b, 0xf3, 0x5f, 0xc8, 0x94, 0x9a, 0xd3, 0x97, 0xea, 0x52, 0xcb, 0x34, 0xac, 0xba, 0xd4, 0x72,
	0xf1, 0x0a, 0x60, 0x2d, 0x23, 0x08, 0xe4, 0x71, 0x45, 0x84, 0x59, 0x31, 0x6a, 0x7a, 0xf3, 0xba,
	0x4f, 0x1b, 0x9b, 0xdb, 0xdd, 0xca, 0xc6, 0x96, 0x69, 0x47, 0x65, 0x63, 0x4b, 0x65, 0x61, 0x7f,
	0xed, 0xf3, 0xea, 0xc4, 0xf5, 0xec, 0x5a, 0xf2, 0xfb, 0xf0, 0xec, 0x4f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x66, 0x00, 0xc5, 0xe4, 0x61, 0x08, 0x00, 0x00,
}