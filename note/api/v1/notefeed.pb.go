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

	Items []*FeedNoteItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
	Count int32           `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
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

func (x *RandomGetResponse) GetItems() []*FeedNoteItem {
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

type FeedNoteItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NoteId    uint64       `protobuf:"varint,1,opt,name=note_id,json=noteId,proto3" json:"note_id,omitempty"`
	Title     string       `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Desc      string       `protobuf:"bytes,3,opt,name=desc,proto3" json:"desc,omitempty"`
	CreatedAt int64        `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	Images    []*NoteImage `protobuf:"bytes,5,rep,name=images,proto3" json:"images,omitempty"`
	Likes     uint64       `protobuf:"varint,6,opt,name=likes,proto3" json:"likes,omitempty"`     // 点赞数量
	Author    int64        `protobuf:"varint,7,opt,name=author,proto3" json:"author,omitempty"`   // 笔记作者
	Replies   uint64       `protobuf:"varint,8,opt,name=replies,proto3" json:"replies,omitempty"` // 点赞数
	UpdatedAt int64        `protobuf:"varint,9,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

func (x *FeedNoteItem) Reset() {
	*x = FeedNoteItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FeedNoteItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FeedNoteItem) ProtoMessage() {}

func (x *FeedNoteItem) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FeedNoteItem.ProtoReflect.Descriptor instead.
func (*FeedNoteItem) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{2}
}

func (x *FeedNoteItem) GetNoteId() uint64 {
	if x != nil {
		return x.NoteId
	}
	return 0
}

func (x *FeedNoteItem) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *FeedNoteItem) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

func (x *FeedNoteItem) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *FeedNoteItem) GetImages() []*NoteImage {
	if x != nil {
		return x.Images
	}
	return nil
}

func (x *FeedNoteItem) GetLikes() uint64 {
	if x != nil {
		return x.Likes
	}
	return 0
}

func (x *FeedNoteItem) GetAuthor() int64 {
	if x != nil {
		return x.Author
	}
	return 0
}

func (x *FeedNoteItem) GetReplies() uint64 {
	if x != nil {
		return x.Replies
	}
	return 0
}

func (x *FeedNoteItem) GetUpdatedAt() int64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

type GetFeedNoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NoteId uint64 `protobuf:"varint,1,opt,name=note_id,json=noteId,proto3" json:"note_id,omitempty"` //笔记id
}

func (x *GetFeedNoteRequest) Reset() {
	*x = GetFeedNoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFeedNoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFeedNoteRequest) ProtoMessage() {}

func (x *GetFeedNoteRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFeedNoteRequest.ProtoReflect.Descriptor instead.
func (*GetFeedNoteRequest) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{3}
}

func (x *GetFeedNoteRequest) GetNoteId() uint64 {
	if x != nil {
		return x.NoteId
	}
	return 0
}

type GetFeedNoteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Item *FeedNoteItem `protobuf:"bytes,1,opt,name=item,proto3" json:"item,omitempty"`
}

func (x *GetFeedNoteResponse) Reset() {
	*x = GetFeedNoteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetFeedNoteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetFeedNoteResponse) ProtoMessage() {}

func (x *GetFeedNoteResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetFeedNoteResponse.ProtoReflect.Descriptor instead.
func (*GetFeedNoteResponse) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{4}
}

func (x *GetFeedNoteResponse) GetItem() *FeedNoteItem {
	if x != nil {
		return x.Item
	}
	return nil
}

type RecommendGetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid     int64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`                        // 用户id
	NeedNum int32 `protobuf:"varint,2,opt,name=need_num,json=needNum,proto3" json:"need_num,omitempty"` // 推荐条数
}

func (x *RecommendGetRequest) Reset() {
	*x = RecommendGetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecommendGetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecommendGetRequest) ProtoMessage() {}

func (x *RecommendGetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecommendGetRequest.ProtoReflect.Descriptor instead.
func (*RecommendGetRequest) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{5}
}

func (x *RecommendGetRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *RecommendGetRequest) GetNeedNum() int32 {
	if x != nil {
		return x.NeedNum
	}
	return 0
}

type RecommendGetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RecommendGetResponse) Reset() {
	*x = RecommendGetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecommendGetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecommendGetResponse) ProtoMessage() {}

func (x *RecommendGetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecommendGetResponse.ProtoReflect.Descriptor instead.
func (*RecommendGetResponse) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{6}
}

type GetUserRecentPostRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Uid   int64 `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Count int32 `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *GetUserRecentPostRequest) Reset() {
	*x = GetUserRecentPostRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserRecentPostRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserRecentPostRequest) ProtoMessage() {}

func (x *GetUserRecentPostRequest) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserRecentPostRequest.ProtoReflect.Descriptor instead.
func (*GetUserRecentPostRequest) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{7}
}

func (x *GetUserRecentPostRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *GetUserRecentPostRequest) GetCount() int32 {
	if x != nil {
		return x.Count
	}
	return 0
}

type GetUserRecentPostResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*FeedNoteItem `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *GetUserRecentPostResponse) Reset() {
	*x = GetUserRecentPostResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_v1_notefeed_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUserRecentPostResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUserRecentPostResponse) ProtoMessage() {}

func (x *GetUserRecentPostResponse) ProtoReflect() protoreflect.Message {
	mi := &file_v1_notefeed_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUserRecentPostResponse.ProtoReflect.Descriptor instead.
func (*GetUserRecentPostResponse) Descriptor() ([]byte, []int) {
	return file_v1_notefeed_proto_rawDescGZIP(), []int{8}
}

func (x *GetUserRecentPostResponse) GetItems() []*FeedNoteItem {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_v1_notefeed_proto protoreflect.FileDescriptor

var file_v1_notefeed_proto_rawDesc = []byte{
	0x0a, 0x11, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x66, 0x65, 0x65, 0x64, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76,
	0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0d, 0x76,
	0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x15, 0x76, 0x31,
	0x2f, 0x6e, 0x6f, 0x74, 0x65, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x61, 0x63, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x33, 0x0a, 0x10, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x42, 0x09, 0xba, 0x48, 0x06, 0x1a, 0x04, 0x18, 0x1e, 0x20,
	0x00, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x22, 0x5a, 0x0a, 0x11, 0x52, 0x61, 0x6e, 0x64,
	0x6f, 0x6d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a,
	0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6e,
	0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x65, 0x65, 0x64, 0x4e,
	0x6f, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x12, 0x14,
	0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x22, 0x87, 0x02, 0x0a, 0x0c, 0x46, 0x65, 0x65, 0x64, 0x4e, 0x6f, 0x74,
	0x65, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x74, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x6e, 0x6f, 0x74, 0x65, 0x49, 0x64, 0x12, 0x14,
	0x0a, 0x05, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74,
	0x69, 0x74, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x64, 0x65, 0x73, 0x63, 0x12, 0x1d, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x2e, 0x0a, 0x06, 0x69, 0x6d, 0x61, 0x67, 0x65,
	0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4e, 0x6f, 0x74, 0x65, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52,
	0x06, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x73,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x61,
	0x75, 0x74, 0x68, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x65, 0x73,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x65, 0x73, 0x12,
	0x1d, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x36,
	0x0a, 0x12, 0x47, 0x65, 0x74, 0x46, 0x65, 0x65, 0x64, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x07, 0x6e, 0x6f, 0x74, 0x65, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x07, 0xba, 0x48, 0x04, 0x32, 0x02, 0x20, 0x00, 0x52, 0x06,
	0x6e, 0x6f, 0x74, 0x65, 0x49, 0x64, 0x22, 0x44, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x46, 0x65, 0x65,
	0x64, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a,
	0x04, 0x69, 0x74, 0x65, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6e, 0x6f,
	0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x65, 0x65, 0x64, 0x4e, 0x6f,
	0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x04, 0x69, 0x74, 0x65, 0x6d, 0x22, 0x4d, 0x0a, 0x13,
	0x52, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x03, 0x75, 0x69, 0x64, 0x12, 0x24, 0x0a, 0x08, 0x6e, 0x65, 0x65, 0x64, 0x5f, 0x6e, 0x75,
	0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x42, 0x09, 0xba, 0x48, 0x06, 0x1a, 0x04, 0x18, 0x1e,
	0x20, 0x00, 0x52, 0x07, 0x6e, 0x65, 0x65, 0x64, 0x4e, 0x75, 0x6d, 0x22, 0x16, 0x0a, 0x14, 0x52,
	0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x4b, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65,
	0x63, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69,
	0x64, 0x12, 0x1d, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x42, 0x07, 0xba, 0x48, 0x04, 0x1a, 0x02, 0x18, 0x05, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x22, 0x4c, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x63, 0x65, 0x6e,
	0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2f, 0x0a,
	0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x6e,
	0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x65, 0x65, 0x64, 0x4e,
	0x6f, 0x74, 0x65, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x32, 0xe8,
	0x02, 0x0a, 0x0f, 0x4e, 0x6f, 0x74, 0x65, 0x46, 0x65, 0x65, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x4a, 0x0a, 0x09, 0x52, 0x61, 0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x12,
	0x1d, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61,
	0x6e, 0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e,
	0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x52, 0x61, 0x6e,
	0x64, 0x6f, 0x6d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x50,
	0x0a, 0x0b, 0x47, 0x65, 0x74, 0x46, 0x65, 0x65, 0x64, 0x4e, 0x6f, 0x74, 0x65, 0x12, 0x1f, 0x2e,
	0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x46,
	0x65, 0x65, 0x64, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x20,
	0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74,
	0x46, 0x65, 0x65, 0x64, 0x4e, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x53, 0x0a, 0x0c, 0x52, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x47, 0x65, 0x74,
	0x12, 0x20, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x52,
	0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x21, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31,
	0x2e, 0x52, 0x65, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x64, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x62, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x25, 0x2e, 0x6e, 0x6f, 0x74,
	0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72,
	0x52, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x26, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e,
	0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x63, 0x65, 0x6e, 0x74, 0x50, 0x6f, 0x73,
	0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x9b, 0x01, 0x0a, 0x0f, 0x63, 0x6f,
	0x6d, 0x2e, 0x6e, 0x6f, 0x74, 0x65, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x42, 0x0d, 0x4e,
	0x6f, 0x74, 0x65, 0x66, 0x65, 0x65, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2b,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x79, 0x61, 0x6e, 0x72,
	0x65, 0x61, 0x64, 0x62, 0x6f, 0x6f, 0x6b, 0x73, 0x2f, 0x77, 0x68, 0x69, 0x6d, 0x65, 0x72, 0x2f,
	0x6e, 0x6f, 0x74, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x4e, 0x41,
	0x58, 0xaa, 0x02, 0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x2e, 0x41, 0x70, 0x69, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x0b, 0x4e, 0x6f, 0x74, 0x65, 0x5c, 0x41, 0x70, 0x69, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x17,
	0x4e, 0x6f, 0x74, 0x65, 0x5c, 0x41, 0x70, 0x69, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x4e, 0x6f, 0x74, 0x65, 0x3a, 0x3a,
	0x41, 0x70, 0x69, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_v1_notefeed_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_v1_notefeed_proto_goTypes = []any{
	(*RandomGetRequest)(nil),          // 0: note.api.v1.RandomGetRequest
	(*RandomGetResponse)(nil),         // 1: note.api.v1.RandomGetResponse
	(*FeedNoteItem)(nil),              // 2: note.api.v1.FeedNoteItem
	(*GetFeedNoteRequest)(nil),        // 3: note.api.v1.GetFeedNoteRequest
	(*GetFeedNoteResponse)(nil),       // 4: note.api.v1.GetFeedNoteResponse
	(*RecommendGetRequest)(nil),       // 5: note.api.v1.RecommendGetRequest
	(*RecommendGetResponse)(nil),      // 6: note.api.v1.RecommendGetResponse
	(*GetUserRecentPostRequest)(nil),  // 7: note.api.v1.GetUserRecentPostRequest
	(*GetUserRecentPostResponse)(nil), // 8: note.api.v1.GetUserRecentPostResponse
	(*NoteImage)(nil),                 // 9: note.api.v1.NoteImage
}
var file_v1_notefeed_proto_depIdxs = []int32{
	2, // 0: note.api.v1.RandomGetResponse.items:type_name -> note.api.v1.FeedNoteItem
	9, // 1: note.api.v1.FeedNoteItem.images:type_name -> note.api.v1.NoteImage
	2, // 2: note.api.v1.GetFeedNoteResponse.item:type_name -> note.api.v1.FeedNoteItem
	2, // 3: note.api.v1.GetUserRecentPostResponse.items:type_name -> note.api.v1.FeedNoteItem
	0, // 4: note.api.v1.NoteFeedService.RandomGet:input_type -> note.api.v1.RandomGetRequest
	3, // 5: note.api.v1.NoteFeedService.GetFeedNote:input_type -> note.api.v1.GetFeedNoteRequest
	5, // 6: note.api.v1.NoteFeedService.RecommendGet:input_type -> note.api.v1.RecommendGetRequest
	7, // 7: note.api.v1.NoteFeedService.GetUserRecentPost:input_type -> note.api.v1.GetUserRecentPostRequest
	1, // 8: note.api.v1.NoteFeedService.RandomGet:output_type -> note.api.v1.RandomGetResponse
	4, // 9: note.api.v1.NoteFeedService.GetFeedNote:output_type -> note.api.v1.GetFeedNoteResponse
	6, // 10: note.api.v1.NoteFeedService.RecommendGet:output_type -> note.api.v1.RecommendGetResponse
	8, // 11: note.api.v1.NoteFeedService.GetUserRecentPost:output_type -> note.api.v1.GetUserRecentPostResponse
	8, // [8:12] is the sub-list for method output_type
	4, // [4:8] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_v1_notefeed_proto_init() }
func file_v1_notefeed_proto_init() {
	if File_v1_notefeed_proto != nil {
		return
	}
	file_v1_note_proto_init()
	file_v1_noteinteract_proto_init()
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
		file_v1_notefeed_proto_msgTypes[2].Exporter = func(v any, i int) any {
			switch v := v.(*FeedNoteItem); i {
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
		file_v1_notefeed_proto_msgTypes[3].Exporter = func(v any, i int) any {
			switch v := v.(*GetFeedNoteRequest); i {
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
		file_v1_notefeed_proto_msgTypes[4].Exporter = func(v any, i int) any {
			switch v := v.(*GetFeedNoteResponse); i {
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
		file_v1_notefeed_proto_msgTypes[5].Exporter = func(v any, i int) any {
			switch v := v.(*RecommendGetRequest); i {
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
		file_v1_notefeed_proto_msgTypes[6].Exporter = func(v any, i int) any {
			switch v := v.(*RecommendGetResponse); i {
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
		file_v1_notefeed_proto_msgTypes[7].Exporter = func(v any, i int) any {
			switch v := v.(*GetUserRecentPostRequest); i {
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
		file_v1_notefeed_proto_msgTypes[8].Exporter = func(v any, i int) any {
			switch v := v.(*GetUserRecentPostResponse); i {
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
			NumMessages:   9,
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
