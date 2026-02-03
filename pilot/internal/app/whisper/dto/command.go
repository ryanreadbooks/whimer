package dto

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

// MsgReq 发送消息请求
type MsgReq struct {
	Type    vo.MsgType
	Cid     string
	Content *vo.MsgContent
}

func (m *MsgReq) SetContentType() {
	m.Content.Type = m.Type
}

func (m *MsgReq) Validate(_ context.Context) error {
	if m == nil {
		return xerror.ErrNilArg
	}
	if !vo.IsValidMsgType(m.Type) {
		return errors.ErrUnsupportedMsgType
	}
	if err := m.Content.Validate(); err != nil {
		return err
	}
	return nil
}

type CreateP2PChatCommand struct {
	Uid    int64  `json:"-"`
	Target int64  `json:"target"`
	Type   string `json:"type"`
}

func (c *CreateP2PChatCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}
	if c.Target == 0 {
		return errors.ErrUserNotFound
	}
	if !vo.IsValidChatType(c.Type) {
		return errors.ErrUnsupportedChatType
	}
	return nil
}

func (c *CreateP2PChatCommand) IsP2P() bool {
	return vo.ChatType(c.Type) == vo.P2PChat
}

type CreateChatResult struct {
	ChatId string `json:"chat_id"`
}

// SendMsgContent 发送消息请求的内容（不含preview）
type SendMsgContent struct {
	Text  *SendMsgTextContent  `json:"text,omitempty,optional"`
	Image *SendMsgImageContent `json:"image,omitempty,optional"`
}

// SendMsgTextContent 发送文本消息内容（不含preview）
type SendMsgTextContent struct {
	Content string `json:"content"`
}

// SendMsgImageContent 发送图片消息内容（不含preview）
type SendMsgImageContent struct {
	Key    string `json:"key"`
	Height uint32 `json:"height"`
	Width  uint32 `json:"width"`
	Format string `json:"format"`
}

type SendChatMsgCommand struct {
	ChatId  string          `json:"chat_id"`
	Type    MsgType         `json:"type"`
	Cid     string          `json:"cid"`
	Content *SendMsgContent `json:"content"`
}

func (c *SendChatMsgCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}
	if c.ChatId == "" {
		return errors.ErrChatNotExists
	}
	if !c.Type.IsValid() {
		return errors.ErrUnsupportedMsgType
	}
	if c.Content == nil {
		return errors.ErrInvalidMsgContent
	}
	if c.Type == MsgTypeText && c.Content.Text == nil {
		return errors.ErrInvalidMsgContent
	}
	if c.Type == MsgTypeImage && c.Content.Image == nil {
		return errors.ErrInvalidMsgContent
	}

	return nil
}

func (c *SendChatMsgCommand) ToMsgReq() *MsgReq {
	var content *vo.MsgContent
	if c.Content != nil {
		content = &vo.MsgContent{}
		if c.Content.Text != nil {
			content.Text = &vo.MsgTextContent{
				Content: c.Content.Text.Content,
			}
		}
		if c.Content.Image != nil {
			content.Image = &vo.MsgImageContent{
				Key:    c.Content.Image.Key,
				Height: c.Content.Image.Height,
				Width:  c.Content.Image.Width,
				Format: c.Content.Image.Format,
			}
		}
	}

	req := &MsgReq{
		Type:    c.Type.ToVO(),
		Cid:     c.Cid,
		Content: content,
	}
	req.SetContentType()
	return req
}

type SendChatMsgResult struct {
	MsgId string `json:"msg_id"`
}

type RecallChatMsgCommand struct {
	ChatId string `json:"chat_id"`
	MsgId  string `json:"msg_id"`
}

func (c *RecallChatMsgCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}
	if c.ChatId == "" {
		return errors.ErrChatNotExists
	}
	if c.MsgId == "" {
		return errors.ErrChatMsgNotExists
	}
	return nil
}

type ClearChatUnreadCommand struct {
	ChatId string `json:"chat_id"`
}

func (c *ClearChatUnreadCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}
	if c.ChatId == "" {
		return errors.ErrChatNotExists
	}
	return nil
}
