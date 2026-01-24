package uploadresource

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
)

// Metadata 定义每类资源对应的桶名称和key前缀（配置解析用）
type Metadata struct {
	Bucket        string `json:"bucket"`
	Prefix        string `json:"prefix"`
	PrependBucket bool   `json:"prepend_bucket,optional,default=true"`
	PrependPrefix bool   `json:"prepend_prefix,optional"`
	Stringer      string `json:"stringer,optional"`
	PrefixSegment string `json:"prefix_segment,optional"`
}

func (m *Metadata) GetStringer() keygen.Stringer {
	if m.Stringer == "random_v7" {
		return keygen.RandomStringerV7{}
	}
	return keygen.RandomStringer{}
}

// Type 类型别名，用于配置解析和 infra 层
type Type = storagevo.ObjectType

// 类型常量别名，保持向后兼容
const (
	NoteImage      = storagevo.ObjectTypeNoteImage
	NoteVideo      = storagevo.ObjectTypeNoteVideo
	NoteVideoCover = storagevo.ObjectTypeNoteVideoCover
	CommentImage   = storagevo.ObjectTypeCommentImage
)

// CheckValid 检查类型是否有效
func CheckValid(s string) bool {
	return Type(s).IsValid()
}
