// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: v1/notefeed.proto

package v1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

type RandomGetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Count int32 `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *RandomGetRequest) Reset() {
	*x = RandomGetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomGetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomGetRequest) ProtoMessage() {}

func (x *RandomGetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomGetRequest.ProtoReflect.Descriptor instead.
func (*RandomGetRequest) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{0}
}

func (x *RandomGetRequest) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type RandomGetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*NoteItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	Count int32       `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *RandomGetResponse) Reset() {
	*x = RandomGetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RandomGetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RandomGetResponse) ProtoMessage() {}

func (x *RandomGetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RandomGetResponse.ProtoReflect.Descriptor instead.
func (*RandomGetResponse) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{1}
}

func (x *RandomGetResponse) GetItems() []*NoteItem {
	if x != nil {
		return x.Items
	}
	return nil
}

func (x *RandomGetResponse) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

var File_v1_notefeed_proto protoreflect.FileDescriptor

var file_v1_notefeed_proto_rawDesc = []byte{
	0x0a, 0x11, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x66, 0x65, 0x65, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x73, 0x64, 0x6b, 0x2e, 0x76, 0x31,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x76,
	0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x33, 0x0a, 0x10,
	0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x1f, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42,
	0x09, 0xba, 0x48, 0x06, 0x1a, 0x04, 0x18, 0x1e, 0x20, 0x00, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x22, 0x56, 0x0a, 0x11, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x73, 0x64, 0x6b,
	0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74,
	0x65, 0x6d, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x32, 0x5d, 0x0a, 0x0f, 0x4e, 0x6f, 0x74,
	0x65, 0x46, 0x65, 0x65, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x4a, 0x0a, 0x09,
	0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x12, 0x1d, 0x2e, 0x6e, 0x6f, 0x74, 0x65,
	0x2e, 0x73, 0x64, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e,
	0x73, 0x64, 0x6b, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x9b, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d,
	0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x73, 0x64, 0x6b, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x4e, 0x6f,
	0x74, 0x65, 0x66, 0x65, 0x65, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2b, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x79, 0x61, 0x6e, 0x72, 0x65,
	0x61, 0x64, 0x62, 0x6f, 0x6f, 0x6b, 0x73, 0x2f, 0x77, 0x68, 0x69, 0x6d, 0x65, 0x72, 0x2f, 0x6e,
	0x6f, 0x74, 0x65, 0x2f, 0x73, 0x64, 0x6b, 0x2f, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x4e, 0x53, 0x58,
	0xaa, 0x02, 0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x2e, 0x53, 0x64, 0x6b, 0x2e, 0x56, 0x31, 0xca, 0x02,
	0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x5c, 0x53, 0x64, 0x6b, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x17, 0x4e,
	0x6f, 0x74, 0x65, 0x5c, 0x53, 0x64, 0x6b, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x4e, 0x6f, 0x74, 0x65, 0x3a, 0x3a, 0x53,
	0x64, 0x6b, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_notefeed_proto_rawDescOnce sync.Once
	file_v1_notefeed_proto_rawDescData = file_v1_notefeed_proto_rawDesc
)

func file_v1_notefeed_proto_rawDescGZIP() []byte {
	file_v1_notefeed_proto_rawDescOnce.Do(func() {
		file_v1_notefeed_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_notefeed_proto_rawDescData)
	})
	return file_v1_notefeed_proto_rawDescData
}

var file_v1_notefeed_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_v1_notefeed_proto_goTypes = []any{
	(*RandomGetRequest)(nil),  // 0: note.sdk.v1.RandomGetRequest
	(*RandomGetResponse)(nil), // 1: note.sdk.v1.RandomGetResponse
	(*NoteItem)(nil),          // 2: note.sdk.v1.NoteItem
}
var file_v1_notefeed_proto_depIdxs = []int32{
	2, // 0: note.sdk.v1.RandomGetResponse.items:type_name -> note.sdk.v1.NoteItem
	0, // 1: note.sdk.v1.NoteFeedService.RandomGet:input_type -> note.sdk.v1.RandomGetRequest
	1, // 2: note.sdk.v1.NoteFeedService.RandomGet:output_type -> note.sdk.v1.RandomGetResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_v1_notefeed_proto_init() }
func file_v1_notefeed_proto_init() {
	if File_v1_notefeed_proto != nil {
		return
	}
	file_v1_note_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_v1_notefeed_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RandomGetRequest); i {
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
		file_v1_notefeed_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*RandomGetResponse); i {
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
			RawDescriptor: file_v1_notefeed_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_v1_notefeed_proto_goTypes,
		DependencyIndexes: file_v1_notefeed_proto_depIdxs,
		MessageInfos:      file_v1_notefeed_proto_msgTypes,
	}.Build()
	File_v1_notefeed_proto = out.File
	file_v1_notefeed_proto_rawDesc = nil
	file_v1_notefeed_proto_goTypes = nil
	file_v1_notefeed_proto_depIdxs = nil
}
