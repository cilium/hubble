// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v4.23.4
// source: envoy/type/matcher/v3/filter_state.proto

package matcherv3

import (
	_ "github.com/cncf/xds/go/udpa/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// FilterStateMatcher provides a general interface for matching the filter state objects.
type FilterStateMatcher struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The filter state key to retrieve the object.
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// Types that are assignable to Matcher:
	//
	//	*FilterStateMatcher_StringMatch
	Matcher isFilterStateMatcher_Matcher `protobuf_oneof:"matcher"`
}

func (x *FilterStateMatcher) Reset() {
	*x = FilterStateMatcher{}
	if protoimpl.UnsafeEnabled {
		mi := &file_envoy_type_matcher_v3_filter_state_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FilterStateMatcher) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FilterStateMatcher) ProtoMessage() {}

func (x *FilterStateMatcher) ProtoReflect() protoreflect.Message {
	mi := &file_envoy_type_matcher_v3_filter_state_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FilterStateMatcher.ProtoReflect.Descriptor instead.
func (*FilterStateMatcher) Descriptor() ([]byte, []int) {
	return file_envoy_type_matcher_v3_filter_state_proto_rawDescGZIP(), []int{0}
}

func (x *FilterStateMatcher) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (m *FilterStateMatcher) GetMatcher() isFilterStateMatcher_Matcher {
	if m != nil {
		return m.Matcher
	}
	return nil
}

func (x *FilterStateMatcher) GetStringMatch() *StringMatcher {
	if x, ok := x.GetMatcher().(*FilterStateMatcher_StringMatch); ok {
		return x.StringMatch
	}
	return nil
}

type isFilterStateMatcher_Matcher interface {
	isFilterStateMatcher_Matcher()
}

type FilterStateMatcher_StringMatch struct {
	// Matches the filter state object as a string value.
	StringMatch *StringMatcher `protobuf:"bytes,2,opt,name=string_match,json=stringMatch,proto3,oneof"`
}

func (*FilterStateMatcher_StringMatch) isFilterStateMatcher_Matcher() {}

var File_envoy_type_matcher_v3_filter_state_proto protoreflect.FileDescriptor

var file_envoy_type_matcher_v3_filter_state_proto_rawDesc = []byte{
	0x0a, 0x28, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x61, 0x74,
	0x63, 0x68, 0x65, 0x72, 0x2f, 0x76, 0x33, 0x2f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x65, 0x6e, 0x76, 0x6f,
	0x79, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76,
	0x33, 0x1a, 0x22, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x61,
	0x74, 0x63, 0x68, 0x65, 0x72, 0x2f, 0x76, 0x33, 0x2f, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x75, 0x64, 0x70, 0x61, 0x2f, 0x61, 0x6e, 0x6e, 0x6f,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x01,
	0x0a, 0x12, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x4d, 0x61, 0x74,
	0x63, 0x68, 0x65, 0x72, 0x12, 0x19, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x49, 0x0a, 0x0c, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x5f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x24, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x2e, 0x74, 0x79,
	0x70, 0x65, 0x2e, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x2e, 0x76, 0x33, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x48, 0x00, 0x52, 0x0b, 0x73,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x42, 0x0e, 0x0a, 0x07, 0x6d, 0x61,
	0x74, 0x63, 0x68, 0x65, 0x72, 0x12, 0x03, 0xf8, 0x42, 0x01, 0x42, 0x89, 0x01, 0x0a, 0x23, 0x69,
	0x6f, 0x2e, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2e, 0x65, 0x6e, 0x76,
	0x6f, 0x79, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x2e,
	0x76, 0x33, 0x42, 0x10, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x46, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f, 0x67, 0x6f,
	0x2d, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2d, 0x70, 0x6c, 0x61, 0x6e, 0x65, 0x2f, 0x65,
	0x6e, 0x76, 0x6f, 0x79, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x2f, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65,
	0x72, 0x2f, 0x76, 0x33, 0x3b, 0x6d, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72, 0x76, 0x33, 0xba, 0x80,
	0xc8, 0xd1, 0x06, 0x02, 0x10, 0x02, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_envoy_type_matcher_v3_filter_state_proto_rawDescOnce sync.Once
	file_envoy_type_matcher_v3_filter_state_proto_rawDescData = file_envoy_type_matcher_v3_filter_state_proto_rawDesc
)

func file_envoy_type_matcher_v3_filter_state_proto_rawDescGZIP() []byte {
	file_envoy_type_matcher_v3_filter_state_proto_rawDescOnce.Do(func() {
		file_envoy_type_matcher_v3_filter_state_proto_rawDescData = protoimpl.X.CompressGZIP(file_envoy_type_matcher_v3_filter_state_proto_rawDescData)
	})
	return file_envoy_type_matcher_v3_filter_state_proto_rawDescData
}

var file_envoy_type_matcher_v3_filter_state_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_envoy_type_matcher_v3_filter_state_proto_goTypes = []interface{}{
	(*FilterStateMatcher)(nil), // 0: envoy.type.matcher.v3.FilterStateMatcher
	(*StringMatcher)(nil),      // 1: envoy.type.matcher.v3.StringMatcher
}
var file_envoy_type_matcher_v3_filter_state_proto_depIdxs = []int32{
	1, // 0: envoy.type.matcher.v3.FilterStateMatcher.string_match:type_name -> envoy.type.matcher.v3.StringMatcher
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_envoy_type_matcher_v3_filter_state_proto_init() }
func file_envoy_type_matcher_v3_filter_state_proto_init() {
	if File_envoy_type_matcher_v3_filter_state_proto != nil {
		return
	}
	file_envoy_type_matcher_v3_string_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_envoy_type_matcher_v3_filter_state_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FilterStateMatcher); i {
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
	file_envoy_type_matcher_v3_filter_state_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*FilterStateMatcher_StringMatch)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_envoy_type_matcher_v3_filter_state_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_envoy_type_matcher_v3_filter_state_proto_goTypes,
		DependencyIndexes: file_envoy_type_matcher_v3_filter_state_proto_depIdxs,
		MessageInfos:      file_envoy_type_matcher_v3_filter_state_proto_msgTypes,
	}.Build()
	File_envoy_type_matcher_v3_filter_state_proto = out.File
	file_envoy_type_matcher_v3_filter_state_proto_rawDesc = nil
	file_envoy_type_matcher_v3_filter_state_proto_goTypes = nil
	file_envoy_type_matcher_v3_filter_state_proto_depIdxs = nil
}