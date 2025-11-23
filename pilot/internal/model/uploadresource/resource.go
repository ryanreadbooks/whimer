package uploadresource

import "github.com/ryanreadbooks/whimer/misc/oss/keygen"

// 定义每类资源对应的桶名称和key前缀
type Metadata struct {
	Bucket        string `json:"bucket"`
	Prefix        string `json:"prefix"`
	PrependBucket bool   `json:"prepend_bucket,optional,default=true"`
	PrependPrefix bool   `json:"prepend_prefix,optional"`
	Stringer      string `json:"stringer,optional"`
}

func (m *Metadata) GetStringer() keygen.Stringer {
	if m.Stringer == "random_v7" {
		return keygen.RandomStringerV7{}
	}

	return keygen.RandomStringer{}
}

type Type string

const (
	NoteImage    Type = "note_image"
	CommentImage Type = "comment_image"
)

var (
	validType = map[Type]struct{}{
		NoteImage:    {},
		CommentImage: {},
	}
)

func CheckValid(s string) bool {
	_, ok := validType[Type(s)]
	return ok
}
