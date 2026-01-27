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
	m.Content.ContentType = m.Type
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

type SendChatMsgCommand struct {
	ChatId  string         `json:"chat_id"`
	Type    int32          `json:"type"`
	Cid     string         `json:"cid"`
	Content *vo.MsgContent `json:"content"`
}

func (c *SendChatMsgCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}
	if c.ChatId == "" {
		return errors.ErrChatNotExists
	}
	return nil
}

func (c *SendChatMsgCommand) ToMsgReq() *MsgReq {
	req := &MsgReq{
		Type:    vo.MsgType(c.Type),
		Cid:     c.Cid,
		Content: c.Content,
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

type ListRecentChatsQuery struct {
	Uid    int64  `form:"-"`
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,default=30"`
}

func (q *ListRecentChatsQuery) Validate() error {
	if q == nil {
		return xerror.ErrNilArg
	}
	if q.Count <= 0 {
		q.Count = 30
	}
	if q.Count > 50 {
		q.Count = 50
	}
	return nil
}

type ListChatMsgsQuery struct {
	Uid    int64  `form:"-"`
	ChatId string `form:"chat_id"`
	Pos    int64  `form:"pos,optional"`
	Count  int32  `form:"count,default=50"`
}

func (q *ListChatMsgsQuery) Validate() error {
	if q == nil {
		return xerror.ErrNilArg
	}
	if q.ChatId == "" {
		return errors.ErrChatNotExists
	}
	if q.Count <= 0 {
		q.Count = 50
	}
	if q.Count > 100 {
		q.Count = 100
	}
	return nil
}
