package model

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type FollowReq struct {
	Target int64 `json:"target"`
	Action int8  `json:"action"`
}

func (r *FollowReq) Validate() error {
	if r.Target == 0 {
		return xerror.ErrArgs.Msg("非法用户id")
	}

	if r.Action != ActionFollow && r.Action != ActionUnFollow {
		return xerror.ErrArgs.Msg("不支持的操作")
	}

	return nil
}

type GetIsFollowingReq struct {
	Uid int64 `form:"uid"`
}

func (r *GetIsFollowingReq) Validate() error {
	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

type UpdateSettingReq struct {
	ShowFans    bool `json:"show_fans"`
	ShowFollows bool `json:"show_follows"`
}
