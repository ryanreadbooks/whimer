package model

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type UidReq struct {
	Uid int64 `form:"uid"`
}

func (r *UidReq) Validate() error {
	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

type ListInfosReq struct {
	Uids string `form:"uids"` // 多个用,分隔
}

type GetUserReq struct {
	Uid int64 `form:"uid"`
}

type HoverReq struct {
	Uid int64 `form:"uid"`
}

func (r *HoverReq) Validate() error {
	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

type GetFanOrFollowingListReq struct {
	Uid   int64 `form:"uid"`
	Page  int32 `form:"page,optional,default=1"`
	Count int32 `form:"count,optional,default=20"`
}

func (r *GetFanOrFollowingListReq) Validate() error {
	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	return nil
}

type UserWithFollowStatus struct {
	User     *User          `json:"user"`
	Relation RelationStatus `json:"relation"`
}

type GetFanOrFollowingListResp struct {
	Items []*UserWithFollowStatus `json:"items"`
	Total int64                   `json:"total"`
}
