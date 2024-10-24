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
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: go.chromium.org/luci/bisection/proto/config/config.proto

package configpb

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

// Config is the service-wide configuration data for LUCI Bisection
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Service-wide configuration data for Gerrit integration
	GerritConfig *GerritConfig `protobuf:"bytes,1,opt,name=gerrit_config,json=gerritConfig,proto3" json:"gerrit_config,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetGerritConfig() *GerritConfig {
	if x != nil {
		return x.GerritConfig
	}
	return nil
}

// GerritConfig is the configuration data for Gerrit integration
type GerritConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether Gerrit API actions are enabled
	ActionsEnabled bool `protobuf:"varint,1,opt,name=actions_enabled,json=actionsEnabled,proto3" json:"actions_enabled,omitempty"`
	// Settings for creating reverts for culprit CLs
	CreateRevertSettings *GerritConfig_RevertActionSettings `protobuf:"bytes,2,opt,name=create_revert_settings,json=createRevertSettings,proto3" json:"create_revert_settings,omitempty"`
	// Settings for submitting reverts for culprit CLs
	SubmitRevertSettings *GerritConfig_RevertActionSettings `protobuf:"bytes,3,opt,name=submit_revert_settings,json=submitRevertSettings,proto3" json:"submit_revert_settings,omitempty"`
	// Maximum age of a culprit (sec) for its revert to be eligible
	// for the submit action.
	//
	// The age of a culprit is based on the time since the culprit was merged.
	// If a culprit is older than this limit, LUCI Bisection will skip
	// submitting its corresponding revert.
	MaxRevertibleCulpritAge int64 `protobuf:"varint,4,opt,name=max_revertible_culprit_age,json=maxRevertibleCulpritAge,proto3" json:"max_revertible_culprit_age,omitempty"`
}

func (x *GerritConfig) Reset() {
	*x = GerritConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GerritConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GerritConfig) ProtoMessage() {}

func (x *GerritConfig) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GerritConfig.ProtoReflect.Descriptor instead.
func (*GerritConfig) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescGZIP(), []int{1}
}

func (x *GerritConfig) GetActionsEnabled() bool {
	if x != nil {
		return x.ActionsEnabled
	}
	return false
}

func (x *GerritConfig) GetCreateRevertSettings() *GerritConfig_RevertActionSettings {
	if x != nil {
		return x.CreateRevertSettings
	}
	return nil
}

func (x *GerritConfig) GetSubmitRevertSettings() *GerritConfig_RevertActionSettings {
	if x != nil {
		return x.SubmitRevertSettings
	}
	return nil
}

func (x *GerritConfig) GetMaxRevertibleCulpritAge() int64 {
	if x != nil {
		return x.MaxRevertibleCulpritAge
	}
	return 0
}

// Settings for revert-related actions
type GerritConfig_RevertActionSettings struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether the action is enabled
	Enabled bool `protobuf:"varint,1,opt,name=enabled,proto3" json:"enabled,omitempty"`
	// The maximum number of times the action can be performed per day
	DailyLimit uint32 `protobuf:"varint,2,opt,name=daily_limit,json=dailyLimit,proto3" json:"daily_limit,omitempty"`
}

func (x *GerritConfig_RevertActionSettings) Reset() {
	*x = GerritConfig_RevertActionSettings{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GerritConfig_RevertActionSettings) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GerritConfig_RevertActionSettings) ProtoMessage() {}

func (x *GerritConfig_RevertActionSettings) ProtoReflect() protoreflect.Message {
	mi := &file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GerritConfig_RevertActionSettings.ProtoReflect.Descriptor instead.
func (*GerritConfig_RevertActionSettings) Descriptor() ([]byte, []int) {
	return file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescGZIP(), []int{1, 0}
}

func (x *GerritConfig_RevertActionSettings) GetEnabled() bool {
	if x != nil {
		return x.Enabled
	}
	return false
}

func (x *GerritConfig_RevertActionSettings) GetDailyLimit() uint32 {
	if x != nil {
		return x.DailyLimit
	}
	return 0
}

var File_go_chromium_org_luci_bisection_proto_config_config_proto protoreflect.FileDescriptor

var file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDesc = []byte{
	0x0a, 0x38, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f, 0x72,
	0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2f, 0x63, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x6c, 0x75, 0x63, 0x69,
	0x2e, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x22, 0x52, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x48, 0x0a, 0x0d, 0x67,
	0x65, 0x72, 0x72, 0x69, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x23, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47, 0x65, 0x72, 0x72, 0x69,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0c, 0x67, 0x65, 0x72, 0x72, 0x69, 0x74, 0x43,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xa7, 0x03, 0x0a, 0x0c, 0x47, 0x65, 0x72, 0x72, 0x69, 0x74,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x27, 0x0a, 0x0f, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0e, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12,
	0x6e, 0x0a, 0x16, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x5f, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74,
	0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x38, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47, 0x65, 0x72, 0x72, 0x69, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x52, 0x14, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12,
	0x6e, 0x0a, 0x16, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x5f, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74,
	0x5f, 0x73, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x38, 0x2e, 0x6c, 0x75, 0x63, 0x69, 0x2e, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x2e, 0x47, 0x65, 0x72, 0x72, 0x69, 0x74, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x2e, 0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x52, 0x14, 0x73, 0x75, 0x62, 0x6d, 0x69,
	0x74, 0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12,
	0x3b, 0x0a, 0x1a, 0x6d, 0x61, 0x78, 0x5f, 0x72, 0x65, 0x76, 0x65, 0x72, 0x74, 0x69, 0x62, 0x6c,
	0x65, 0x5f, 0x63, 0x75, 0x6c, 0x70, 0x72, 0x69, 0x74, 0x5f, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x17, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x69, 0x62,
	0x6c, 0x65, 0x43, 0x75, 0x6c, 0x70, 0x72, 0x69, 0x74, 0x41, 0x67, 0x65, 0x1a, 0x51, 0x0a, 0x14,
	0x52, 0x65, 0x76, 0x65, 0x72, 0x74, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x12, 0x1f,
	0x0a, 0x0b, 0x64, 0x61, 0x69, 0x6c, 0x79, 0x5f, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0a, 0x64, 0x61, 0x69, 0x6c, 0x79, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x42,
	0x36, 0x5a, 0x34, 0x67, 0x6f, 0x2e, 0x63, 0x68, 0x72, 0x6f, 0x6d, 0x69, 0x75, 0x6d, 0x2e, 0x6f,
	0x72, 0x67, 0x2f, 0x6c, 0x75, 0x63, 0x69, 0x2f, 0x62, 0x69, 0x73, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x3b, 0x63,
	0x6f, 0x6e, 0x66, 0x69, 0x67, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescOnce sync.Once
	file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescData = file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDesc
)

func file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescGZIP() []byte {
	file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescOnce.Do(func() {
		file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescData)
	})
	return file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDescData
}

var file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_go_chromium_org_luci_bisection_proto_config_config_proto_goTypes = []interface{}{
	(*Config)(nil),                            // 0: luci.bisection.config.Config
	(*GerritConfig)(nil),                      // 1: luci.bisection.config.GerritConfig
	(*GerritConfig_RevertActionSettings)(nil), // 2: luci.bisection.config.GerritConfig.RevertActionSettings
}
var file_go_chromium_org_luci_bisection_proto_config_config_proto_depIdxs = []int32{
	1, // 0: luci.bisection.config.Config.gerrit_config:type_name -> luci.bisection.config.GerritConfig
	2, // 1: luci.bisection.config.GerritConfig.create_revert_settings:type_name -> luci.bisection.config.GerritConfig.RevertActionSettings
	2, // 2: luci.bisection.config.GerritConfig.submit_revert_settings:type_name -> luci.bisection.config.GerritConfig.RevertActionSettings
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_go_chromium_org_luci_bisection_proto_config_config_proto_init() }
func file_go_chromium_org_luci_bisection_proto_config_config_proto_init() {
	if File_go_chromium_org_luci_bisection_proto_config_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Config); i {
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
		file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GerritConfig); i {
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
		file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GerritConfig_RevertActionSettings); i {
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
			RawDescriptor: file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_go_chromium_org_luci_bisection_proto_config_config_proto_goTypes,
		DependencyIndexes: file_go_chromium_org_luci_bisection_proto_config_config_proto_depIdxs,
		MessageInfos:      file_go_chromium_org_luci_bisection_proto_config_config_proto_msgTypes,
	}.Build()
	File_go_chromium_org_luci_bisection_proto_config_config_proto = out.File
	file_go_chromium_org_luci_bisection_proto_config_config_proto_rawDesc = nil
	file_go_chromium_org_luci_bisection_proto_config_config_proto_goTypes = nil
	file_go_chromium_org_luci_bisection_proto_config_config_proto_depIdxs = nil
}
