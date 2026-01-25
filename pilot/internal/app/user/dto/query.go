package dto

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/user/errors"
)

type ListUsersReq struct {
	Uids string `form:"uids"` // 多个用,分隔
}

type GetUserReq struct {
	Uid int64 `form:"uid"`
}

func (q *GetUserReq) Validate() error {
	if q.Uid == 0 {
		return errors.ErrInvalidUserId
	}
	return nil
}

type GetFanOrFollowingListReq struct {
	Uid   int64 `form:"uid"`
	Page  int32 `form:"page,optional,default=1"`
	Count int32 `form:"count,optional,default=20"`
}

func (q *GetFanOrFollowingListReq) Validate() error {
	if q.Uid == 0 {
		return errors.ErrInvalidUserId
	}
	return nil
}

type HoverReq struct {
	Uid int64 `form:"uid"`
}

func (q *HoverReq) Validate() error {
	if q.Uid == 0 {
		return errors.ErrInvalidUserId
	}
	return nil
}

type MentionUserReq struct {
	Search string `form:"search,optional"`
}

type SetNoteShowSettingReq struct {
	ShowNoteLikes bool `json:"show_note_likes"`
}
