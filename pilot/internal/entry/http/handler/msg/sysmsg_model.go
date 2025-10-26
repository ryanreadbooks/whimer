package msg

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	sysnotifymodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
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

type SystemMsgForMention struct {
	*sysnotifymodel.MentionedMsg
	User *usermodel.User `json:"user"` // 谁@
}

type ListSysMsgMentionsResp struct {
	ChatId  string                 `json:"chat_id"`
	Msgs    []*SystemMsgForMention `json:"msgs"`
	HasNext bool                   `json:"has_next"`
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

type SystemMsgForReply struct {
	*sysnotifymodel.ReplyMsg
	User *usermodel.User `json:"user"` // 谁回复
}

type ListSysMsgRepliesResp struct {
	ChatId  string               `json:"chat_id"`
	Msgs    []*SystemMsgForReply `json:"msgs"`
	HasNext bool                 `json:"has_next"`
}
