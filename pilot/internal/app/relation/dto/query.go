package dto

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/relation/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
)

type FollowCommand struct {
	Target int64 `json:"target"`
	Action int8  `json:"action"`
}

func (r *FollowCommand) Validate() error {
	if r.Target == 0 {
		return errors.ErrInvalidUserId
	}

	if !vo.FollowAction(r.Action).IsValid() {
		return errors.ErrInvalidAction
	}

	return nil
}

type CheckFollowingQuery struct {
	Uid int64 `form:"uid"`
}

func (r *CheckFollowingQuery) Validate() error {
	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

type UpdateSettingsCommand struct {
	ShowFans    bool `json:"show_fans"`
	ShowFollows bool `json:"show_follows"`
}
