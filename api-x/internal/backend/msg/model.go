package msg

import "github.com/ryanreadbooks/whimer/misc/xerror"

type ListChatsReq struct {
	Seq   int64 `form:"seq,optional"`
	Count int   `form:"count,optional"`
}

type CreateChatReq struct {
	Target int64 `json:"target"`
}

func (r *CreateChatReq) Validate() error {
	if r.Target == 0 {
		return xerror.ErrArgs.Msg("用户不存在")
	}

	return nil
}
