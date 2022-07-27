// Copyright 2022 The LUCI Authors.
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
// 	protoc-gen-go v1.28.0
// 	protoc        v3.17.3
// source: go.chromium.org/luci/cipd/api/cipd/v1/verification_log.proto

package api

import (
	_ "go.chromium.org/luci/common/bq/pb"
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

// VerificationLogEntry defines a schema for `verification` BigQuery table.
//
// It records details about hash verification jobs, in particular to collect
// information for https://crbug.com/1261988.
//
// This is a best effort log populated using in-memory buffers. Some entries may
// be dropped if a process crashes before it flushes the buffer.
//
// Field types must be compatible with BigQuery Storage Write API, see
// https://cloud.google.com/bigquery/docs/write-api#data_type_conversions
type VerificationLogEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OperationId        int64  `protobuf:"varint,1,opt,name=operation_id,json=operationId,proto3" json:"operation_id,omitempty"`                       // matches UploadOperation.ID
	InitiatedBy        string `protobuf:"bytes,2,opt,name=initiated_by,json=initiatedBy,proto3" json:"initiated_by,omitempty"`                        // e.g. "user:someone@example.com"
	TempGsPath         string `protobuf:"bytes,3,opt,name=temp_gs_path,json=tempGsPath,proto3" json:"temp_gs_path,omitempty"`                         // the GS object in the staging area being verified
	ExpectedInstanceId string `protobuf:"bytes,4,opt,name=expected_instance_id,json=expectedInstanceId,proto3" json:"expected_instance_id,omitempty"` // may be empty if not known
	VerifiedInstanceId string `protobuf:"bytes,5,opt,name=verified_instance_id,json=verifiedInstanceId,proto3" json:"verified_instance_id,omitempty"` // always populated on success
	Submitted          int64  `protobuf:"varint,6,opt,name=submitted,proto3" json:"submitted,omitempty"`                                              // microseconds since epoch
	Started            int64  `protobuf:"varint,7,opt,name=started,proto3" json:"started,omitempty"`                                                  // microseconds since epoch
	Finished           int64  `protobuf:"varint,8,opt,name=finished,proto3" json:"finished,omitempty"`                                                // microseconds since epoch
	ServiceVersion     string `protobuf:"bytes,9,opt,name=service_version,json=serviceVersion,proto3" json:"service_version,omitempty"`               // GAE service version e.g. "4123-abcdef"
	ProcessId          string `protobuf:"bytes,10,opt,name=process_id,json=processId,proto3" json:"process_id,omitempty"`                             // identifier of the concrete backend process
	TraceId            string `protobuf:"bytes,11,opt,name=trace_id,json=traceId,proto3" json:"trace_id,omitempty"`                                   // Cloud Trace ID of the request
	FileSize           int64  `protobuf:"varint,12,opt,name=file_size,json=fileSize,proto3" json:"file_size,omitempty"`                               // total file size in bytes
	VerificationSpeed  int64  `protobuf:"varint,13,opt,name=verification_speed,json=verificationSpeed,proto3" json:"verification_speed,omitempty"`    // file_size / duration, in bytes per second
	Outcome            string `protobuf:"bytes,14,opt,name=outcome,proto3" json:"outcome,omitempty"`                                                  // see cas.UploadStatus enum
	Error              string `protobuf:"bytes,15,opt,name=error,proto3" json:"error,omitempty"`                                                      // error message, if any
}

func (x *VerificationLogEntry) Reset() {
	*x = VerificationLogEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VerificationLogEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VerificationLogEntry) ProtoMessage() {}

func (x *VerificationLogEntry) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VerificationLogEntry.ProtoReflect.Descriptor instead.
func (*VerificationLogEntry) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescGZIP(), []int{0}
}

func (x *VerificationLogEntry) GetOperationId() int64 {
	if x != nil {
		return x.OperationId
	}
	return 0
}

func (x *VerificationLogEntry) GetInitiatedBy() string {
	if x != nil {
		return x.InitiatedBy
	}
	return ""
}

func (x *VerificationLogEntry) GetTempGsPath() string {
	if x != nil {
		return x.TempGsPath
	}
	return ""
}

func (x *VerificationLogEntry) GetExpectedInstanceId() string {
	if x != nil {
		return x.ExpectedInstanceId
	}
	return ""
}

func (x *VerificationLogEntry) GetVerifiedInstanceId() string {
	if x != nil {
		return x.VerifiedInstanceId
	}
	return ""
}

func (x *VerificationLogEntry) GetSubmitted() int64 {
	if x != nil {
		return x.Submitted
	}
	return 0
}

func (x *VerificationLogEntry) GetStarted() int64 {
	if x != nil {
		return x.Started
	}
	return 0
}

func (x *VerificationLogEntry) GetFinished() int64 {
	if x != nil {
		return x.Finished
	}
	return 0
}

func (x *VerificationLogEntry) GetServiceVersion() string {
	if x != nil {
		return x.ServiceVersion
	}
	return ""
}

func (x *VerificationLogEntry) GetProcessId() string {
	if x != nil {
		return x.ProcessId
	}
	return ""
}

func (x *VerificationLogEntry) GetTraceId() string {
	if x != nil {
		return x.TraceId
	}
	return ""
}

func (x *VerificationLogEntry) GetFileSize() int64 {
	if x != nil {
		return x.FileSize
	}
	return 0
}

func (x *VerificationLogEntry) GetVerificationSpeed() int64 {
	if x != nil {
		return x.VerificationSpeed
	}
	return 0
}

func (x *VerificationLogEntry) GetOutcome() string {
	if x != nil {
		return x.Outcome
	}
	return ""
}

func (x *VerificationLogEntry) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto protoreflect.FileDescriptor

var file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDesc = []byte{
	0x0a, 0x3c, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x63, 0x69, 0x70, 0x64, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x63, 0x69, 0x70, 0x64, 0x2f, 0x76, 0x31, 0x2f, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04,
	0x63, 0x69, 0x70, 0x64, 0x1a, 0x2f, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75,
	0x6d, 0x2e, 0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2f, 0x62, 0x71, 0x2f, 0x70, 0x62, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc8, 0x04, 0x0a, 0x14, 0x56, 0x65, 0x72, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x6f, 0x67, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x21,
	0x0a, 0x0c, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x6f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x12, 0x21, 0x0a, 0x0c, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x62,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x61, 0x74,
	0x65, 0x64, 0x42, 0x79, 0x12, 0x20, 0x0a, 0x0c, 0x74, 0x65, 0x6d, 0x70, 0x5f, 0x67, 0x73, 0x5f,
	0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x65, 0x6d, 0x70,
	0x47, 0x73, 0x50, 0x61, 0x74, 0x68, 0x12, 0x30, 0x0a, 0x14, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74,
	0x65, 0x64, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x65, 0x64, 0x49, 0x6e,
	0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x76, 0x65, 0x72, 0x69,
	0x66, 0x69, 0x65, 0x64, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x65, 0x64,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x2d, 0x0a, 0x09, 0x73, 0x75,
	0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x42, 0x0f, 0xe2,
	0xbc, 0x24, 0x0b, 0x0a, 0x09, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x54, 0x41, 0x4d, 0x50, 0x52, 0x09,
	0x73, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x74, 0x65, 0x64, 0x12, 0x29, 0x0a, 0x07, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x65, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x42, 0x0f, 0xe2, 0xbc, 0x24, 0x0b,
	0x0a, 0x09, 0x54, 0x49, 0x4d, 0x45, 0x53, 0x54, 0x41, 0x4d, 0x50, 0x52, 0x07, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x65, 0x64, 0x12, 0x2b, 0x0a, 0x08, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65, 0x64,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x03, 0x42, 0x0f, 0xe2, 0xbc, 0x24, 0x0b, 0x0a, 0x09, 0x54, 0x49,
	0x4d, 0x45, 0x53, 0x54, 0x41, 0x4d, 0x50, 0x52, 0x08, 0x66, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x65,
	0x64, 0x12, 0x27, 0x0a, 0x0f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x76, 0x65, 0x72,
	0x73, 0x69, 0x6f, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72,
	0x6f, 0x63, 0x65, 0x73, 0x73, 0x5f, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x70, 0x72, 0x6f, 0x63, 0x65, 0x73, 0x73, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x74, 0x72, 0x61,
	0x63, 0x65, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x73, 0x69, 0x7a,
	0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x12, 0x2d, 0x0a, 0x12, 0x76, 0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x73, 0x70, 0x65, 0x65, 0x64, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x03, 0x52, 0x11, 0x76,
	0x65, 0x72, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x70, 0x65, 0x65, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x6f, 0x75, 0x74, 0x63, 0x6f, 0x6d, 0x65, 0x18, 0x0e, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x6f, 0x75, 0x74, 0x63, 0x6f, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72,
	0x72, 0x6f, 0x72, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72,
	0x42, 0x2b, 0x5a, 0x29, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e,
	0x6f, 0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x63, 0x69, 0x70, 0x64, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x63, 0x69, 0x70, 0x64, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescOnce sync.Once
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescData = file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDesc
)

func file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescGZIP() []byte {
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescOnce.Do(func() {
		file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescData = protoimpl.X.CompressGZIP(file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescData)
	})
	return file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDescData
}

var file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_goTypes = []interface{}{
	(*VerificationLogEntry)(nil), // 0: cipd.VerificationLogEntry
}
var file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_init() }
func file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_init() {
	if File_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VerificationLogEntry); i {
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
			RawDescriptor: file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_goTypes,
		DependencyIndexes: file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_depIdxs,
		MessageInfos:      file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_msgTypes,
	}.Build()
	File_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto = out.File
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_rawDesc = nil
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_goTypes = nil
	file_go_chromium_org_luci_cipd_api_cipd_v1_verification_log_proto_depIdxs = nil
}