package vo

import (
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	kB = 1024
	mB = 1024 * kB
)

// ObjectType 对象类型
type ObjectType string

const (
	ObjectTypeNoteImage      ObjectType = "note_image"
	ObjectTypeNoteVideo      ObjectType = "note_video"
	ObjectTypeNoteVideoCover ObjectType = "note_video_cover"
	ObjectTypeCommentImage   ObjectType = "comment_image"
)

func (t ObjectType) String() string {
	return string(t)
}

func (t ObjectType) IsValid() bool {
	_, ok := objectTypeMetadataMap[t]
	return ok
}

// PermitSize 返回允许的最大字节数
func (t ObjectType) PermitSize() int64 {
	switch t {
	case ObjectTypeNoteImage:
		return 10 * mB
	case ObjectTypeNoteVideo:
		return 500 * mB
	case ObjectTypeNoteVideoCover:
		return 2 * mB
	case ObjectTypeCommentImage:
		return 10 * mB
	}
	return 0
}

var (
	allowedImageType = []string{"image/jpeg", "image/png", "image/webp"}
	allowedVideoType = []string{"video/mp4"}
)

// PermitContentType 返回允许的内容类型
func (t ObjectType) PermitContentType() []string {
	switch t {
	case ObjectTypeNoteImage, ObjectTypeCommentImage, ObjectTypeNoteVideoCover:
		return allowedImageType
	case ObjectTypeNoteVideo:
		return allowedVideoType
	}
	return nil
}

// CheckContent 检查对象内容是否符合规则
func (t ObjectType) CheckContent(content []byte, totalSize int64) error {
	mimeType := mimetype.Detect(content).String()
	permitSize := t.PermitSize()
	permitType := t.PermitContentType()
	if slices.Contains(permitType, mimeType) && totalSize <= permitSize {
		return nil
	}
	return xerror.ErrArgs.Msg("对象格式非法")
}

// Metadata 返回对象类型的存储元数据
func (t ObjectType) Metadata() ObjectTypeMetadata {
	if m, ok := objectTypeMetadataMap[t]; ok {
		return m
	}
	return ObjectTypeMetadata{}
}

// ObjectTypeMetadata 对象类型的存储元数据
type ObjectTypeMetadata struct {
	Bucket        string
	Prefix        string
	PrependBucket bool
	PrependPrefix bool
	PrefixSegment string
	stringer      keygen.Stringer
}

func (m ObjectTypeMetadata) GetStringer() keygen.Stringer {
	if m.stringer != nil {
		return m.stringer
	}
	return keygen.RandomStringer{}
}

// AllObjectTypes 返回所有支持的对象类型
func AllObjectTypes() []ObjectType {
	return []ObjectType{
		ObjectTypeNoteImage,
		ObjectTypeNoteVideo,
		ObjectTypeNoteVideoCover,
		ObjectTypeCommentImage,
	}
}

// objectTypeMetadataMap 所有对象类型的元数据定义
var objectTypeMetadataMap = map[ObjectType]ObjectTypeMetadata{
	ObjectTypeNoteImage: {
		Bucket:        "nota",
		Prefix:        "assets",
		PrependBucket: true,
		PrependPrefix: false,
	},
	ObjectTypeNoteVideo: {
		Bucket:        "videos",
		Prefix:        "note/cosmic",
		PrependBucket: true,
		PrependPrefix: true,
		PrefixSegment: "note",
	},
	ObjectTypeNoteVideoCover: {
		Bucket:        "nota",
		Prefix:        "cover",
		PrependBucket: true,
		PrependPrefix: true,
	},
	ObjectTypeCommentImage: {
		Bucket:        "pics",
		Prefix:        "cmt_inline",
		PrependBucket: true,
		PrependPrefix: true,
		stringer:      keygen.RandomStringerV7{},
	},
}

// ObjectInfo 通用对象信息
type ObjectInfo struct {
	FileId string
}

// ObjectMeta 对象元数据（存储配置相关）
type ObjectMeta struct {
	Bucket        string
	Prefix        string
	PrefixSegment string
}

// UploadTicket 上传凭证
type UploadTicket struct {
	FileIds      []string
	Bucket       string
	AccessKey    string
	SecretKey    string
	SessionToken string
	ExpireAt     int64
	UploadAddr   string
}

// PostPolicyTicket Post Policy 上传凭证
type PostPolicyTicket struct {
	FileId     string
	UploadAddr string
	Form       map[string]string
}

type UploadTicketDeprecated struct {
	FileIds     []string `json:"file_ids"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}
