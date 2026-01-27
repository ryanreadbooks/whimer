package msg

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type CursorAndCountReq struct {
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,optional,default=20"`
}

func (r *CursorAndCountReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	return nil
}

type SysChatReq struct {
	ChatId string `form:"chat_id" json:"chat_id"`
}

func (r *SysChatReq) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.ChatId == "" {
		return xerror.ErrArgs.Msg("会话不存在")
	}

	return nil
}
