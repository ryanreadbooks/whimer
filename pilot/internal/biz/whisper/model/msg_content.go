package model

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type MsgContent struct {
	contentType MsgType `json:"-"` // see [Msg.SetContentType]

	Text *MsgTextContent `json:"text,omitempty"`
}

func (c *MsgContent) Validate() error {
	if c == nil {
		return errors.ErrInvalidMsgContent
	}

	switch c.contentType {
	case MsgText:
		return c.Text.Validate()
	}

	return errors.ErrUnsupportedMsgType
}

type MsgTextContent struct {
	Content string `json:"content"`
	Preview string `json:"preview,optional"`
}

func (t *MsgTextContent) AsReqPb() *userchatv1.MsgReq_Text {
	return &userchatv1.MsgReq_Text{
		Text: &pbmsg.MsgContentText{
			Content: t.Content,
		},
	}
}

func (t *MsgTextContent) Validate() error {
	if t == nil {
		return errors.ErrInvalidMsgContent
	}

	l := utf8.RuneCountInString(t.Content)
	if l > MaxTextContentLength {
		return xerror.ErrArgs.Msg("消息超长")
	}

	return nil
}

type MsgImageContent struct {
	Key     string `json:"key"`
	Height  uint32 `json:"height"`
	Width   uint32 `json:"width"`
	Format  string `json:"format"`
	Preview string `json:"preview,optional"` // 预览文本
}

func (i *MsgImageContent) Validate() error {
	if i == nil {
		return errors.ErrInvalidMsgContent
	}

	return nil
}

func (i *MsgImageContent) AsReqPb() *userchatv1.MsgReq_Image {
	return &userchatv1.MsgReq_Image{
		Image: &pbmsg.MsgContentImage{
			Key:    i.Key,
			Height: i.Height,
			Width:  i.Width,
			Format: i.Format,
		},
	}
}
