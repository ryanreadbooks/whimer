package model

import "github.com/ryanreadbooks/whimer/misc/xerror"

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
	NoteId uint64 `form:"note_id"`
	Source string `form:"source,optional"`
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
