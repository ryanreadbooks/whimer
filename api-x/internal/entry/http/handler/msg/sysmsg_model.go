package msg

import (
	sysnotifymodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/sysnotify/model"
	usermodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/user/model"
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

type SystemMsgForMention struct {
	*sysnotifymodel.MentionedMsg
	User *usermodel.User `json:"user"`
}

type ListSysMsgMentionsResp struct {
	Msgs    []*SystemMsgForMention `json:"msgs"`
	HasNext bool                   `json:"has_next"`
}
