// Copyright 2021 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: go.chromium.org/luci/luci_notify/api/service/v1/rpc.proto

package lucinotifypb

import prpc "go.chromium.org/luci/grpc/prpc"

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A request message for `QueryTreeClosers` RPC.
type QueryTreeClosersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Required. Only the tree closers that are associated with the builders in
	// the project will be returned.
	Project string `protobuf:"bytes,1,opt,name=project,proto3" json:"project,omitempty"`
	// Optional. The maximum number of tree closers to return.
	//
	// The service may return fewer than this value.
	// If unspecified, at most 100 tree closers will be returned.
	// The maximum value is 1000; values above 1000 will be coerced to 1000.
	PageSize int32 `protobuf:"varint,2,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	// Optional. A page token, received from a previous `QueryTreeClosers` call.
	// Provide this to retrieve the subsequent page.
	//
	// When paginating, all parameters provided to `QueryTreeClosers`, with the
	// exception of page_size and page_token, must match the call that provided
	// the page token.
	PageToken string `protobuf:"bytes,3,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
}

func (x *QueryTreeClosersRequest) Reset() {
	*x = QueryTreeClosersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryTreeClosersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryTreeClosersRequest) ProtoMessage() {}

func (x *QueryTreeClosersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryTreeClosersRequest.ProtoReflect.Descriptor instead.
func (*QueryTreeClosersRequest) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *QueryTreeClosersRequest) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

func (x *QueryTreeClosersRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *QueryTreeClosersRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

// A response message for `QueryTreeClosers` RPC.
type QueryTreeClosersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A list of builder IDs with their associated tree closers.
	BuilderTreeClosers []*QueryTreeClosersResponse_BuilderTreeClosers `protobuf:"bytes,1,rep,name=builder_tree_closers,json=builderTreeClosers,proto3" json:"builder_tree_closers,omitempty"`
	// A token that can be sent as `page_token` to retrieve the next page.
	// If this field is omitted, there are no subsequent pages.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
}

func (x *QueryTreeClosersResponse) Reset() {
	*x = QueryTreeClosersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryTreeClosersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryTreeClosersResponse) ProtoMessage() {}

func (x *QueryTreeClosersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryTreeClosersResponse.ProtoReflect.Descriptor instead.
func (*QueryTreeClosersResponse) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *QueryTreeClosersResponse) GetBuilderTreeClosers() []*QueryTreeClosersResponse_BuilderTreeClosers {
	if x != nil {
		return x.BuilderTreeClosers
	}
	return nil
}

func (x *QueryTreeClosersResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

type QueryTreeClosersResponse_BuilderTreeClosers struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The canonical ID of the builder (e.g. {project}/{bucket}/{builder}).
	BuilderId string `protobuf:"bytes,1,opt,name=builder_id,json=builderId,proto3" json:"builder_id,omitempty"`
	// A list of tree closer hosts that are associated with the builder.
	TreeCloserHosts []string `protobuf:"bytes,2,rep,name=tree_closer_hosts,json=treeCloserHosts,proto3" json:"tree_closer_hosts,omitempty"`
}

func (x *QueryTreeClosersResponse_BuilderTreeClosers) Reset() {
	*x = QueryTreeClosersResponse_BuilderTreeClosers{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryTreeClosersResponse_BuilderTreeClosers) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryTreeClosersResponse_BuilderTreeClosers) ProtoMessage() {}

func (x *QueryTreeClosersResponse_BuilderTreeClosers) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryTreeClosersResponse_BuilderTreeClosers.ProtoReflect.Descriptor instead.
func (*QueryTreeClosersResponse_BuilderTreeClosers) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescGZIP(), []int{1, 0}
}

func (x *QueryTreeClosersResponse_BuilderTreeClosers) GetBuilderId() string {
	if x != nil {
		return x.BuilderId
	}
	return ""
}

func (x *QueryTreeClosersResponse_BuilderTreeClosers) GetTreeCloserHosts() []string {
	if x != nil {
		return x.TreeCloserHosts
	}
	return nil
}

var File_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto protoreflect.FileDescriptor

var file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDesc = []byte{
	0x0a, 0x39, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x5f, 0x6e, 0x6f, 0x74, 0x69,
	0x66, 0x79, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x6c, 0x75, 0x63,
	0x69, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x76, 0x31, 0x22, 0x6f, 0x0a, 0x17, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x1d, 0x0a,
	0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x22, 0x92, 0x02, 0x0a,
	0x18, 0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x6d, 0x0a, 0x14, 0x62, 0x75, 0x69,
	0x6c, 0x64, 0x65, 0x72, 0x5f, 0x74, 0x72, 0x65, 0x65, 0x5f, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x72,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3b, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x6e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x72,
	0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x2e, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f,
	0x73, 0x65, 0x72, 0x73, 0x52, 0x12, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x54, 0x72, 0x65,
	0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x12, 0x26, 0x0a, 0x0f, 0x6e, 0x65, 0x78, 0x74,
	0x5f, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e,
	0x1a, 0x5f, 0x0a, 0x12, 0x42, 0x75, 0x69, 0x6c, 0x64, 0x65, 0x72, 0x54, 0x72, 0x65, 0x65, 0x43,
	0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x75, 0x69, 0x6c,
	0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x2a, 0x0a, 0x11, 0x74, 0x72, 0x65, 0x65, 0x5f, 0x63, 0x6c,
	0x6f, 0x73, 0x65, 0x72, 0x5f, 0x68, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09,
	0x52, 0x0f, 0x74, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x48, 0x6f, 0x73, 0x74,
	0x73, 0x32, 0x76, 0x0a, 0x0b, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73,
	0x12, 0x67, 0x0a, 0x10, 0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f,
	0x73, 0x65, 0x72, 0x73, 0x12, 0x27, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x6e, 0x6f, 0x74, 0x69,
	0x66, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x72, 0x65, 0x65, 0x43,
	0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e,
	0x6c, 0x75, 0x63, 0x69, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2e, 0x76, 0x31, 0x2e, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x54, 0x72, 0x65, 0x65, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x6f, 0x2e,
	0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63,
	0x69, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x5f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x6c, 0x75, 0x63,
	0x69, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescOnce sync.Once
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescData = file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDesc
)

func file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescGZIP() []byte {
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescOnce.Do(func() {
		file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescData)
	})
	return file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDescData
}

var file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_goTypes = []interface{}{
	(*QueryTreeClosersRequest)(nil),                     // 0: luci.notify.v1.QueryTreeClosersRequest
	(*QueryTreeClosersResponse)(nil),                    // 1: luci.notify.v1.QueryTreeClosersResponse
	(*QueryTreeClosersResponse_BuilderTreeClosers)(nil), // 2: luci.notify.v1.QueryTreeClosersResponse.BuilderTreeClosers
}
var file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_depIdxs = []int32{
	2, // 0: luci.notify.v1.QueryTreeClosersResponse.builder_tree_closers:type_name -> luci.notify.v1.QueryTreeClosersResponse.BuilderTreeClosers
	0, // 1: luci.notify.v1.TreeClosers.QueryTreeClosers:input_type -> luci.notify.v1.QueryTreeClosersRequest
	1, // 2: luci.notify.v1.TreeClosers.QueryTreeClosers:output_type -> luci.notify.v1.QueryTreeClosersResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_init() }
func file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_init() {
	if File_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryTreeClosersRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryTreeClosersResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryTreeClosersResponse_BuilderTreeClosers); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_goTypes,
		DependencyIndexes: file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_depIdxs,
		MessageInfos:      file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_msgTypes,
	}.Build()
	File_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto = out.File
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_rawDesc = nil
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_goTypes = nil
	file_go_chromium_org_luci_luci_notify_api_service_v1_rpc_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TreeClosersClient is the client API for TreeClosers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TreeClosersClient interface {
	// Retrieves tree closers associated with the builders from the specified
	// project.
	QueryTreeClosers(ctx context.Context, in *QueryTreeClosersRequest, opts ...grpc.CallOption) (*QueryTreeClosersResponse, error)
}
type treeClosersPRPCClient struct {
	client *prpc.Client
}

func NewTreeClosersPRPCClient(client *prpc.Client) TreeClosersClient {
	return &treeClosersPRPCClient{client}
}

func (c *treeClosersPRPCClient) QueryTreeClosers(ctx context.Context, in *QueryTreeClosersRequest, opts ...grpc.CallOption) (*QueryTreeClosersResponse, error) {
	out := new(QueryTreeClosersResponse)
	err := c.client.Call(ctx, "luci.notify.v1.TreeClosers", "QueryTreeClosers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type treeClosersClient struct {
	cc grpc.ClientConnInterface
}

func NewTreeClosersClient(cc grpc.ClientConnInterface) TreeClosersClient {
	return &treeClosersClient{cc}
}

func (c *treeClosersClient) QueryTreeClosers(ctx context.Context, in *QueryTreeClosersRequest, opts ...grpc.CallOption) (*QueryTreeClosersResponse, error) {
	out := new(QueryTreeClosersResponse)
	err := c.cc.Invoke(ctx, "/luci.notify.v1.TreeClosers/QueryTreeClosers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TreeClosersServer is the server API for TreeClosers service.
type TreeClosersServer interface {
	// Retrieves tree closers associated with the builders from the specified
	// project.
	QueryTreeClosers(context.Context, *QueryTreeClosersRequest) (*QueryTreeClosersResponse, error)
}

// UnimplementedTreeClosersServer can be embedded to have forward compatible implementations.
type UnimplementedTreeClosersServer struct {
}

func (*UnimplementedTreeClosersServer) QueryTreeClosers(context.Context, *QueryTreeClosersRequest) (*QueryTreeClosersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTreeClosers not implemented")
}

func RegisterTreeClosersServer(s prpc.Registrar, srv TreeClosersServer) {
	s.RegisterService(&_TreeClosers_serviceDesc, srv)
}

func _TreeClosers_QueryTreeClosers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryTreeClosersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TreeClosersServer).QueryTreeClosers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/luci.notify.v1.TreeClosers/QueryTreeClosers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TreeClosersServer).QueryTreeClosers(ctx, req.(*QueryTreeClosersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TreeClosers_serviceDesc = grpc.ServiceDesc{
	ServiceName: "luci.notify.v1.TreeClosers",
	HandlerType: (*TreeClosersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "QueryTreeClosers",
			Handler:    _TreeClosers_QueryTreeClosers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "go.chromium.org/luci/luci_notify/api/service/v1/rpc.proto",
}