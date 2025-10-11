package msg

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
)

var (
	ErrUserNotFound = xerror.ErrArgs.Msg("用户不存在")
	ErrChatNotFound = xerror.ErrArgs.Msg("会话不存在")
)

type ListChatsReq struct {
	Seq   int64 `form:"seq,optional"`
	Count int   `form:"count,optional"`
}

type CreateChatReq struct {
	Target int64 `json:"target"`
}

func (r *CreateChatReq) Validate() error {
	if r.Target == 0 {
		return ErrUserNotFound
	}

	return nil
}

type ListMsgsReq struct {
	ChatId int64 `form:"chat_id"`
	Seq    int64 `form:"seq,optional"`
	Count  int   `form:"count,optional"`
}

func (r *ListMsgsReq) Validate() error {
	if r.ChatId <= 0 {
		return ErrChatNotFound
	}

	return nil
}

type SendMsgReq struct {
	ChatId   int64  `json:"chat_id"`
	Receiver int64  `json:"receiver"`
	MsgType  int32  `json:"msg_type"`
	Content  string `json:"content"`
}

func (r *SendMsgReq) Validate() error {
	if r.Receiver == 0 {
		return ErrUserNotFound
	}

	if r.ChatId <= 0 {
		return ErrChatNotFound
	}

	contentLen := utf8.RuneCountInString(r.Content)

	// msgtype + content check
	switch r.MsgType {
	case int32(pbmsg.MsgType_MSG_TYPE_TEXT):
		if contentLen > 500 {
			return xerror.ErrArgs.Msg("消息长度太长")
		}

	default:
		return xerror.ErrArgs.Msg("不支持的消息类型")
	}

	return nil
}

type DeleteChatReq struct {
	ChatId int64 `json:"chat_id"`
}

func (r *DeleteChatReq) Validate() error {
	if r.ChatId <= 0 {
		return ErrChatNotFound
	}

	return nil
}

type DeleteMsgReq struct {
	ChatId int64 `json:"chat_id"`
	MsgId  int64 `json:"msg_id"`
}

func (r *DeleteMsgReq) Validate() error {
	if r.ChatId <= 0 {
		return ErrChatNotFound
	}

	if r.MsgId <= 0 {
		return xerror.ErrArgs.Msg("消息不存在")
	}

	return nil
}


type GetChatReq struct {
	Id int64 `form:"id"`
}
func (r *GetChatReq) Validate() error {
	if r.Id == 0 {
		return xerror.ErrArgs.Msg("会话不存在")
	}

	return nil
}
