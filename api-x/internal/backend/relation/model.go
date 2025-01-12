package relation

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	ActionFollow   = 1
	ActionUnFollow = 2
)

type FollowReq struct {
	Target uint64 `json:"target"`
	Action int8   `json:"action"`
}

func (r *FollowReq) Validate() error {
	if r.Target <= 0 {
		return xerror.ErrArgs.Msg("非法用户id")
	}

	if r.Action != ActionFollow && r.Action != ActionUnFollow {
		return xerror.ErrArgs.Msg("不支持的操作")
	}

	return nil
}
