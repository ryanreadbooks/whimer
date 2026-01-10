package uploadresource

import (
	"slices"

	"github.com/gabriel-vasile/mimetype"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

// 定义每类资源对应的桶名称和key前缀
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

const (
	kB = 1024
	mB = 1024 * kB
)

type Type string

const (
	NoteImage      Type = "note_image"
	NoteVideo      Type = "note_video"
	NoteVideoCover Type = "note_video_cover"
	CommentImage   Type = "comment_image"
)

func (t Type) PermitSize() int64 {
	switch t {
	case NoteImage:
		return 10 * mB
	case NoteVideo:
		return 500 * mB
	case NoteVideoCover:
		return 2 * mB
	case CommentImage:
		return 10 * mB
	}

	return 0
}

var (
	allowedImageType = []string{"image/jpeg", "image/png", "image/webp"}
	allowedVideoType = []string{"video/mp4"}
)

func (t Type) PermitContentType() []string {
	switch t {
	case NoteImage, CommentImage, NoteVideoCover:
		return allowedImageType
	case NoteVideo:
		return allowedVideoType
	}

	return nil
}

func (t Type) Check(b []byte, total int64) error {
	mimeType := mimetype.Detect(b).String()
	permitSize := t.PermitSize()
	permitType := t.PermitContentType()
	if slices.Contains(permitType, mimeType) && total <= permitSize {
		return nil
	}

	return xerror.ErrArgs.Msg("资源格式非法")
}

var (
	validType = map[Type]struct{}{
		NoteImage:      {},
		NoteVideo:      {},
		CommentImage:   {},
		NoteVideoCover: {},
	}
)

func CheckValid(s string) bool {
	_, ok := validType[Type(s)]
	return ok
}
