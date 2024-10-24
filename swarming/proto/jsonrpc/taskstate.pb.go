// Copyright 2019 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This proto file describes the external scheduler plugin API.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: go.chromium.org/luci/swarming/proto/jsonrpc/taskstate.proto

package jsonrpc

import (
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

// TaskState defines the TaskState enum used by the swarming json rpc interface.
// This is not to be confused with the new proto rpc interface, which has its
// own incompatible definition of TaskState.
//
// If you make any modifications, please keep comments in sync and make
// corresponding modifications swarming_rpcs.py:TaskState
type TaskState int32

const (
	// Invalid state, do not use.
	TaskState_INVALID TaskState = 0
	// The task is currently running. This is in fact 3 phases: the initial
	// overhead to fetch input files, the actual task running, and the tear down
	// overhead to archive output files to the server.
	TaskState_RUNNING TaskState = 16
	// The task is currently pending. This means that no bot reaped the task. It
	// will stay in this state until either a task reaps it or the expiration
	// elapsed. The task pending expiration is specified as
	// TaskSlice.expiration_secs, one per task slice.
	TaskState_PENDING TaskState = 32
	// The task is not pending anymore, and never ran due to lack of capacity. This
	// means that other higher priority tasks ran instead and that not enough bots
	// were available to run this task for TaskSlice.expiration_secs seconds.
	TaskState_EXPIRED TaskState = 48
	// The task ran for longer than the allowed time in
	// TaskProperties.execution_timeout_secs or TaskProperties.io_timeout_secs.
	// This means the bot forcefully killed the task process as described in the
	// graceful termination dance in the documentation.
	TaskState_TIMED_OUT TaskState = 64
	// The task ran but the bot had an internal failure, unrelated to the task
	// itself. It can be due to the server being unavailable to get task update,
	// the host on which the bot is running crashing or rebooting, etc.
	TaskState_BOT_DIED TaskState = 80
	// The task never ran, and was manually cancelled via the 'cancel' API before
	// it was reaped.
	TaskState_CANCELED TaskState = 96
	// The task ran and completed normally. The task process exit code may be 0 or
	// another value.
	TaskState_COMPLETED TaskState = 112
	// The task ran but was manually killed via the 'cancel' API. This means the
	// bot forcefully killed the task process as described in the graceful
	// termination dance in the documentation.
	TaskState_KILLED TaskState = 128
	// The task was never set to PENDING and was immediately refused, as the server
	// determined that there is no bot capacity to run this task. This happens
	// because no bot exposes a superset of the requested task dimensions.
	//
	// Set TaskSlice.wait_for_capacity to True to force the server to keep the task
	// slice pending even in this case. Generally speaking, the task will
	// eventually switch to EXPIRED, as there's no bot to run it. That said, there
	// are situations where it is known that in some not-too-distant future a wild
	// bot will appear that will be able to run this task.
	TaskState_NO_RESOURCE TaskState = 256
	// The task encounted an error caused by the client. This means that
	// rerunning the task with the same parameters will not change the result
	TaskState_CLIENT_ERROR TaskState = 512
)

// Enum value maps for TaskState.
var (
	TaskState_name = map[int32]string{
		0:   "INVALID",
		16:  "RUNNING",
		32:  "PENDING",
		48:  "EXPIRED",
		64:  "TIMED_OUT",
		80:  "BOT_DIED",
		96:  "CANCELED",
		112: "COMPLETED",
		128: "KILLED",
		256: "NO_RESOURCE",
		512: "CLIENT_ERROR",
	}
	TaskState_value = map[string]int32{
		"INVALID":      0,
		"RUNNING":      16,
		"PENDING":      32,
		"EXPIRED":      48,
		"TIMED_OUT":    64,
		"BOT_DIED":     80,
		"CANCELED":     96,
		"COMPLETED":    112,
		"KILLED":       128,
		"NO_RESOURCE":  256,
		"CLIENT_ERROR": 512,
	}
)

func (x TaskState) Enum() *TaskState {
	p := new(TaskState)
	*p = x
	return p
}

func (x TaskState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TaskState) Descriptor() protoreflect.EnumDescriptor {
	return file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_enumTypes[0].Descriptor()
}

func (TaskState) Type() protoreflect.EnumType {
	return &file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_enumTypes[0]
}

func (x TaskState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TaskState.Descriptor instead.
func (TaskState) EnumDescriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescGZIP(), []int{0}
}

var File_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto protoreflect.FileDescriptor

var file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDesc = []byte{
	0x0a, 0x3b, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x73, 0x77, 0x61, 0x72, 0x6d, 0x69, 0x6e, 0x67, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6a, 0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x2f, 0x74, 0x61,
	0x73, 0x6b, 0x73, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6a,
	0x73, 0x6f, 0x6e, 0x72, 0x70, 0x63, 0x2a, 0xab, 0x01, 0x0a, 0x09, 0x54, 0x61, 0x73, 0x6b, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x10,
	0x00, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e, 0x4e, 0x49, 0x4e, 0x47, 0x10, 0x10, 0x12, 0x0b,
	0x0a, 0x07, 0x50, 0x45, 0x4e, 0x44, 0x49, 0x4e, 0x47, 0x10, 0x20, 0x12, 0x0b, 0x0a, 0x07, 0x45,
	0x58, 0x50, 0x49, 0x52, 0x45, 0x44, 0x10, 0x30, 0x12, 0x0d, 0x0a, 0x09, 0x54, 0x49, 0x4d, 0x45,
	0x44, 0x5f, 0x4f, 0x55, 0x54, 0x10, 0x40, 0x12, 0x0c, 0x0a, 0x08, 0x42, 0x4f, 0x54, 0x5f, 0x44,
	0x49, 0x45, 0x44, 0x10, 0x50, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x41, 0x4e, 0x43, 0x45, 0x4c, 0x45,
	0x44, 0x10, 0x60, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4d, 0x50, 0x4c, 0x45, 0x54, 0x45, 0x44,
	0x10, 0x70, 0x12, 0x0b, 0x0a, 0x06, 0x4b, 0x49, 0x4c, 0x4c, 0x45, 0x44, 0x10, 0x80, 0x01, 0x12,
	0x10, 0x0a, 0x0b, 0x4e, 0x4f, 0x5f, 0x52, 0x45, 0x53, 0x4f, 0x55, 0x52, 0x43, 0x45, 0x10, 0x80,
	0x02, 0x12, 0x11, 0x0a, 0x0c, 0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x5f, 0x45, 0x52, 0x52, 0x4f,
	0x52, 0x10, 0x80, 0x04, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d,
	0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x73, 0x77, 0x61,
	0x72, 0x6d, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6a, 0x73, 0x6f, 0x6e,
	0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescOnce sync.Once
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescData = file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDesc
)

func file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescGZIP() []byte {
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescOnce.Do(func() {
		file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescData = protoimpl.X.CompressGZIP(file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescData)
	})
	return file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDescData
}

var file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_goTypes = []interface{}{
	(TaskState)(0), // 0: jsonrpc.TaskState
}
var file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_init() }
func file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_init() {
	if File_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_goTypes,
		DependencyIndexes: file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_depIdxs,
		EnumInfos:         file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_enumTypes,
	}.Build()
	File_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto = out.File
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_rawDesc = nil
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_goTypes = nil
	file_go_chromium_org_luci_swarming_proto_jsonrpc_taskstate_proto_depIdxs = nil
}
