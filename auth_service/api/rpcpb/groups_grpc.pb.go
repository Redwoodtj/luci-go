// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package rpcpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GroupsClient is the client API for Groups service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GroupsClient interface {
	// ListGroups returns all the groups currently stored in LUCI Auth service
	// datastore. The groups will be returned in alphabetical order based on their
	// ID.
	ListGroups(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListGroupsResponse, error)
	// GetGroup returns information about an individual group, given the name.
	GetGroup(ctx context.Context, in *GetGroupRequest, opts ...grpc.CallOption) (*AuthGroup, error)
}

type groupsClient struct {
	cc grpc.ClientConnInterface
}

func NewGroupsClient(cc grpc.ClientConnInterface) GroupsClient {
	return &groupsClient{cc}
}

func (c *groupsClient) ListGroups(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*ListGroupsResponse, error) {
	out := new(ListGroupsResponse)
	err := c.cc.Invoke(ctx, "/auth.service.Groups/ListGroups", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *groupsClient) GetGroup(ctx context.Context, in *GetGroupRequest, opts ...grpc.CallOption) (*AuthGroup, error) {
	out := new(AuthGroup)
	err := c.cc.Invoke(ctx, "/auth.service.Groups/GetGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GroupsServer is the server API for Groups service.
// All implementations must embed UnimplementedGroupsServer
// for forward compatibility
type GroupsServer interface {
	// ListGroups returns all the groups currently stored in LUCI Auth service
	// datastore. The groups will be returned in alphabetical order based on their
	// ID.
	ListGroups(context.Context, *emptypb.Empty) (*ListGroupsResponse, error)
	// GetGroup returns information about an individual group, given the name.
	GetGroup(context.Context, *GetGroupRequest) (*AuthGroup, error)
	mustEmbedUnimplementedGroupsServer()
}

// UnimplementedGroupsServer must be embedded to have forward compatible implementations.
type UnimplementedGroupsServer struct {
}

func (UnimplementedGroupsServer) ListGroups(context.Context, *emptypb.Empty) (*ListGroupsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListGroups not implemented")
}
func (UnimplementedGroupsServer) GetGroup(context.Context, *GetGroupRequest) (*AuthGroup, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGroup not implemented")
}
func (UnimplementedGroupsServer) mustEmbedUnimplementedGroupsServer() {}

// UnsafeGroupsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GroupsServer will
// result in compilation errors.
type UnsafeGroupsServer interface {
	mustEmbedUnimplementedGroupsServer()
}

func RegisterGroupsServer(s grpc.ServiceRegistrar, srv GroupsServer) {
	s.RegisterService(&Groups_ServiceDesc, srv)
}

func _Groups_ListGroups_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupsServer).ListGroups(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.service.Groups/ListGroups",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupsServer).ListGroups(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Groups_GetGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GroupsServer).GetGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/auth.service.Groups/GetGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GroupsServer).GetGroup(ctx, req.(*GetGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Groups_ServiceDesc is the grpc.ServiceDesc for Groups service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Groups_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.service.Groups",
	HandlerType: (*GroupsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListGroups",
			Handler:    _Groups_ListGroups_Handler,
		},
		{
			MethodName: "GetGroup",
			Handler:    _Groups_GetGroup_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "go.chromium.org/luci/auth_service/api/rpcpb/groups.proto",
}