// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: v1/note.proto

package v1

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

type NoteItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NoteId   uint64       `protobuf:"varint,1,opt,name=note_id,json=noteId,proto3" json:"note_id,omitempty"`
	Title    string       `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Desc     string       `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
	Privacy  int32        `protobuf:"varint,4,opt,name=privacy,proto3" json:"privacy,omitempty"`
	CreateAt int64        `protobuf:"varint,5,opt,name=create_at,json=createAt,proto3" json:"create_at,omitempty"`
	UpdateAt int64        `protobuf:"varint,6,opt,name=update_at,json=updateAt,proto3" json:"update_at,omitempty"`
	Images   []*NoteImage `protobuf:"bytes,7,rep,name=images,proto3" json:"images,omitempty"`
	Likes    uint64       `protobuf:"varint,8,opt,name=likes,proto3" json:"likes,omitempty"`     // 点赞数量
	Replies  uint64       `protobuf:"varint,9,opt,name=replies,proto3" json:"replies,omitempty"` //评论数量
}

func (x *NoteItem) Reset() {
	*x = NoteItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_note_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NoteItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NoteItem) ProtoMessage() {}

func (x *NoteItem) ProtoReflect() protoreflect.Message {
	mi := &file_v1_note_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NoteItem.ProtoReflect.Descriptor instead.
func (*NoteItem) Descriptor() ([]byte, []int) {
	return file_v1_note_proto_rawDescGZIP(), []int{0}
}

func (x *NoteItem) GetNoteId() uint64 {
	if x != nil {
		return x.NoteId
	}
	return 0
}

func (x *NoteItem) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *NoteItem) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

func (x *NoteItem) GetPrivacy() int32 {
	if x != nil {
		return x.Privacy
	}
	return 0
}

func (x *NoteItem) GetCreateAt() int64 {
	if x != nil {
		return x.CreateAt
	}
	return 0
}

func (x *NoteItem) GetUpdateAt() int64 {
	if x != nil {
		return x.UpdateAt
	}
	return 0
}

func (x *NoteItem) GetImages() []*NoteImage {
	if x != nil {
		return x.Images
	}
	return nil
}

func (x *NoteItem) GetLikes() uint64 {
	if x != nil {
		return x.Likes
	}
	return 0
}

func (x *NoteItem) GetReplies() uint64 {
	if x != nil {
		return x.Replies
	}
	return 0
}

type NoteImageMeta struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Width  uint32 `protobuf:"varint,1,opt,name=width,proto3" json:"width,omitempty"`
	Height uint32 `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	Format string `protobuf:"bytes,3,opt,name=format,proto3" json:"format,omitempty"`
}

func (x *NoteImageMeta) Reset() {
	*x = NoteImageMeta{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_note_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NoteImageMeta) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NoteImageMeta) ProtoMessage() {}

func (x *NoteImageMeta) ProtoReflect() protoreflect.Message {
	mi := &file_v1_note_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NoteImageMeta.ProtoReflect.Descriptor instead.
func (*NoteImageMeta) Descriptor() ([]byte, []int) {
	return file_v1_note_proto_rawDescGZIP(), []int{1}
}

func (x *NoteImageMeta) GetWidth() uint32 {
	if x != nil {
		return x.Width
	}
	return 0
}

func (x *NoteImageMeta) GetHeight() uint32 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *NoteImageMeta) GetFormat() string {
	if x != nil {
		return x.Format
	}
	return ""
}

type NoteImage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url    string         `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Type   int32          `protobuf:"varint,2,opt,name=type,proto3" json:"type,omitempty"`
	UrlPrv string         `protobuf:"bytes,3,opt,name=url_prv,json=urlPrv,proto3" json:"url_prv,omitempty"`
	Meta   *NoteImageMeta `protobuf:"bytes,4,opt,name=meta,proto3" json:"meta,omitempty"`
}

func (x *NoteImage) Reset() {
	*x = NoteImage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_note_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NoteImage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NoteImage) ProtoMessage() {}

func (x *NoteImage) ProtoReflect() protoreflect.Message {
	mi := &file_v1_note_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NoteImage.ProtoReflect.Descriptor instead.
func (*NoteImage) Descriptor() ([]byte, []int) {
	return file_v1_note_proto_rawDescGZIP(), []int{2}
}

func (x *NoteImage) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *NoteImage) GetType() int32 {
	if x != nil {
		return x.Type
	}
	return 0
}

func (x *NoteImage) GetUrlPrv() string {
	if x != nil {
		return x.UrlPrv
	}
	return ""
}

func (x *NoteImage) GetMeta() *NoteImageMeta {
	if x != nil {
		return x.Meta
	}
	return nil
}

var File_v1_note_proto protoreflect.FileDescriptor

var file_v1_note_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0b, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x22, 0x81, 0x02, 0x0a,
	0x08, 0x4e, 0x6f, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x74,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x6e, 0x6f, 0x74, 0x65,
	0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x72, 0x69, 0x76, 0x61, 0x63, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x70,
	0x72, 0x69, 0x76, 0x61, 0x63, 0x79, 0x12, 0x1b, 0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x41, 0x74, 0x12, 0x1b, 0x0a, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x61, 0x74,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x74,
	0x12, 0x2e, 0x0a, 0x06, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x16, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4e,
	0x6f, 0x74, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x06, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x73, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x05, 0x6c, 0x69, 0x6b, 0x65, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x65,
	0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x65, 0x73,
	0x22, 0x55, 0x0a, 0x0d, 0x4e, 0x6f, 0x74, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x74,
	0x61, 0x12, 0x14, 0x0a, 0x05, 0x77, 0x69, 0x64, 0x74, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x05, 0x77, 0x69, 0x64, 0x74, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x22, 0x7a, 0x0a, 0x09, 0x4e, 0x6f, 0x74, 0x65, 0x49,
	0x6d, 0x61, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x72,
	0x6c, 0x5f, 0x70, 0x72, 0x76, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x72, 0x6c,
	0x50, 0x72, 0x76, 0x12, 0x2e, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x4e, 0x6f, 0x74, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x04, 0x6d,
	0x65, 0x74, 0x61, 0x42, 0x97, 0x01, 0x0a, 0x0f, 0x63, 0x6f, 0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x65,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x42, 0x09, 0x4e, 0x6f, 0x74, 0x65, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x72, 0x79, 0x61, 0x6e, 0x72, 0x65, 0x61, 0x64, 0x62, 0x6f, 0x6f, 0x6b, 0x73, 0x2f, 0x77,
	0x68, 0x69, 0x6d, 0x65, 0x72, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76,
	0x31, 0xa2, 0x02, 0x03, 0x4e, 0x41, 0x58, 0xaa, 0x02, 0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x2e, 0x41,
	0x70, 0x69, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x5c, 0x41, 0x70, 0x69,
	0x5c, 0x56, 0x31, 0xe2, 0x02, 0x17, 0x4e, 0x6f, 0x74, 0x65, 0x5c, 0x41, 0x70, 0x69, 0x5c, 0x56,
	0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d,
	0x4e, 0x6f, 0x74, 0x65, 0x3a, 0x3a, 0x41, 0x70, 0x69, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_v1_note_proto_rawDescOnce sync.Once
	file_v1_note_proto_rawDescData = file_v1_note_proto_rawDesc
)

func file_v1_note_proto_rawDescGZIP() []byte {
	file_v1_note_proto_rawDescOnce.Do(func() {
		file_v1_note_proto_rawDescData = protoimpl.X.CompressGZIP(file_v1_note_proto_rawDescData)
	})
	return file_v1_note_proto_rawDescData
}

var file_v1_note_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_v1_note_proto_goTypes = []any{
	(*NoteItem)(nil),      // 0: note.api.v1.NoteItem
	(*NoteImageMeta)(nil), // 1: note.api.v1.NoteImageMeta
	(*NoteImage)(nil),     // 2: note.api.v1.NoteImage
}
var file_v1_note_proto_depIdxs = []int32{
	2, // 0: note.api.v1.NoteItem.images:type_name -> note.api.v1.NoteImage
	1, // 1: note.api.v1.NoteImage.meta:type_name -> note.api.v1.NoteImageMeta
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_v1_note_proto_init() }
func file_v1_note_proto_init() {
	if File_v1_note_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_v1_note_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*NoteItem); i {
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
		file_v1_note_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*NoteImageMeta); i {
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
		file_v1_note_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*NoteImage); i {
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
			RawDescriptor: file_v1_note_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_v1_note_proto_goTypes,
		DependencyIndexes: file_v1_note_proto_depIdxs,
		MessageInfos:      file_v1_note_proto_msgTypes,
	}.Build()
	File_v1_note_proto = out.File
	file_v1_note_proto_rawDesc = nil
	file_v1_note_proto_goTypes = nil
	file_v1_note_proto_depIdxs = nil
}
