package model

import (
	"context"
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type MsgType int32

const (
	MsgText  MsgType = MsgType(pbmsg.MsgType_MSG_TYPE_TEXT)
	MsgImage MsgType = MsgType(pbmsg.MsgType_MSG_TYPE_IMAGE)
)

var (
	validMsgTypeMap = map[MsgType]struct{}{
		MsgText:  {},
		MsgImage: {},
	}
)

func IsValidMsgType(t MsgType) bool {
	_, ok := validMsgTypeMap[t]
	return ok
}

type Msg struct {
	Type    MsgType     `json:"type"`
	Cid     string      `json:"cid"`
	Content *MsgContent `json:"content"`
}

func (m *Msg) SetContentType() {
	m.Content.contentType = m.Type
}

func (m *Msg) Validate(_ context.Context) error {
	if m == nil {
		return xerror.ErrNilArg
	}

	if !IsValidMsgType(m.Type) {
		return errors.ErrUnsupportedMsgType
	}

	// check content
	if err := m.Content.Validate(); err != nil {
		return err
	}

	return nil
}

// assign msg content as pb format for pbMsg
func AssignPbMsgContent(msg *Msg, pbMsg *userchatv1.MsgReq) error {
	switch msg.Type {
	case MsgText:
		pbMsg.Content = msg.Content.Text.AsReqPb()
		return nil
	}

	return errors.ErrUnsupportedMsgType
}

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
