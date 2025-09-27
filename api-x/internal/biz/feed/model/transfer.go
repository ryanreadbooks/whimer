package model

import (
	imodel "github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	CategoryHomeRecommend = "home_recommend"
)

var (
	validCategories = map[string]struct{}{
		CategoryHomeRecommend: {},
	}
)

type FeedRecommendRequest struct {
	NeedNum  int    `form:"need_num"`
	Platform string `form:"platform,optional"`
	Category string `form:"category,optional"`
}

func (r *FeedRecommendRequest) Validate() error {
	const (
		maxNeed = 20
	)

	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NeedNum > maxNeed {
		return xerror.ErrInvalidArgs.Msg("不能拿这么多")
	}

	if r.Category == "" {
		r.Category = CategoryHomeRecommend
	}

	if _, ok := validCategories[r.Category]; !ok {
		return xerror.ErrInvalidArgs.Msg("不支持的信息分类")
	}

	return nil
}

type FeedDetailRequest struct {
	NoteId imodel.NoteId `form:"note_id" path:"note_id"`
	Source string        `form:"source,optional"`
}

func (r *FeedDetailRequest) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NoteId == 0 {
		return xerror.ErrArgs.Msg("笔记不存在")
	}

	return nil
}

type FeedByUserRequest struct {
	Uid    int64         `form:"uid"`
	Cursor imodel.NoteId `form:"cursor,optional"`
	Count  int32         `form:"count,optional"`
}

func (r *FeedByUserRequest) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	if r.Count > 20 {
		r.Count = 20
	}
	if r.Count <= 0 {
		r.Count = 10
	}

	return nil
}

type PageResult struct {
	NextCursor int64
	HasNext    bool
}

type FeedByUserResponse struct {
	Items      []*FeedNoteItem `json:"items"`
	NextCursor imodel.NoteId   `json:"next_cursor"`
	HasNext    bool            `json:"has_next"`
}
