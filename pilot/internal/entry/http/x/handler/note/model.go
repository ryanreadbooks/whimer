package note

import (

	feedmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type ListReq struct {
	Cursor int64 `form:"cursor,optional"`
	Count  int32 `form:"count,optional"`
}

func (r *ListReq) Validate() error {
	if r.Count == 0 {
		r.Count = 15
	}
	if r.Count >= 15 {
		r.Count = 15
	}

	return nil
}

type PageListReq struct {
	Page   int32  `form:"page,optional"`
	Count  int32  `form:"count,default=15"`
	Status string `form:"status,default=published"` // published, auditing, banned
}

func (r *PageListReq) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Count >= 15 {
		r.Count = 15
	}

	return nil
}

type AdminPageListRes struct {
	Items []*model.AdminNoteItem `json:"items"`
	Total int64                  `json:"total"`
}

func NewPageListResFromPb(pb *notev1.PageListNoteResponse) *AdminPageListRes {
	if pb == nil {
		return nil
	}

	items := make([]*model.AdminNoteItem, 0, len(pb.Items))
	for _, item := range pb.Items {
		items = append(items, model.NewAdminNoteItemFromPb(item))
	}

	return &AdminPageListRes{
		Items: items,
		Total: int64(pb.Total),
	}
}

type GetLikedNoteRequest struct {
	Uid    int64  `form:"uid"`
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,optional"`
}

func (r *GetLikedNoteRequest) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	r.Count = min(r.Count, 20)
	if r.Count <= 0 {
		r.Count = 10
	}

	return nil
}

type GetLikedNoteResponse struct {
	Items      []*feedmodel.FeedNoteItem `json:"items"`
	NextCursor string                    `json:"next_cursor"`
	HasNext    bool                      `json:"has_next"`
}
