package vo

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/errors"
)

type MsgType int32

const (
	MsgTypeUnspecified MsgType = 0
	MsgText            MsgType = 1
	MsgImage           MsgType = 2
)

var validMsgTypeMap = map[MsgType]struct{}{
	MsgText:  {},
	MsgImage: {},
}

func IsValidMsgType(t MsgType) bool {
	_, ok := validMsgTypeMap[t]
	return ok
}

type MsgStatus int32

const (
	MsgStatusNormal MsgStatus = 1
	MsgStatusRecall MsgStatus = 2
)

const MaxTextContentLength = 500

type MsgContent struct {
	Type  MsgType
	Text  *MsgTextContent
	Image *MsgImageContent
}

func (c *MsgContent) Validate() error {
	if c == nil {
		return errors.ErrInvalidMsgContent
	}
	switch c.Type {
	case MsgText:
		return c.Text.Validate()
	case MsgImage:
		return c.Image.Validate()
	}
	return errors.ErrUnsupportedMsgType
}

type MsgTextContent struct {
	Content string
	Preview string
}

func (t *MsgTextContent) Validate() error {
	if t == nil {
		return errors.ErrInvalidMsgContent
	}
	if utf8.RuneCountInString(t.Content) > MaxTextContentLength {
		return xerror.ErrArgs.Msg("消息超长")
	}
	return nil
}

type MsgImageContent struct {
	Key     string
	Height  uint32
	Width   uint32
	Format  string
	Preview string
}

func (i *MsgImageContent) Validate() error {
	if i == nil {
		return errors.ErrInvalidMsgContent
	}
	return nil
}

type MsgExt struct {
	Recall *MsgExtRecall
}

type MsgExtRecall struct {
	RecallUid int64
	RecallAt  int64
}
