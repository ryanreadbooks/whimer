package model

import "github.com/ryanreadbooks/whimer/misc/xerror"

type FeedRecommandRequest struct {
	NeedNum int    `form:"need_num"`
	Source  string `form:"source"`
}

func (r *FeedRecommandRequest) Validate() error {
	const (
		maxNeed = 20
	)

	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NeedNum > maxNeed {
		return xerror.ErrInvalidArgs.Msg("不能拿这么多")
	}

	return nil
}

type FeedRecommendResponse struct {
}

type FeedDetailRequest struct {
	NoteId string `form:"note_id"`
	Source string `form:"source"`
}

type FeedDetailResponse struct {
}
