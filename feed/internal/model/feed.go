package model

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
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
	Platform string `form:"platform,omitempty"`
	Category string `form:"category"`
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

	if _, ok := validCategories[r.Category]; !ok {
		return xerror.ErrInvalidArgs.Msg("不支持的信息分类")
	}

	return nil
}

type FeedRecommendResponse struct {
}

type FeedDetailRequest struct {
	NoteId uint64 `form:"note_id"`
	Source string `form:"source"`
}

type FeedDetailResponse struct {
}

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}

type NoteItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type NoteItemImageList []*NoteItemImage

type FeedNoteItem struct {
	NoteId   uint64            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	CreateAt int64             `json:"create_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    uint64            `json:"likes"`  // 笔记总点赞数
	Author   uint64            `json:"author"` // 作者

	// 下面这两个字段要额外设置
	Comments uint64      `json:"comments"` // 笔记总评论数
	Interact Interaction `json:"interact"` // 当前请求的用户与该笔记的交互记录，比如点赞、评论、收藏等动作
}

func NewFeedNoteItemFromPb(pb *notev1.FeedNoteItem) *FeedNoteItem {
	if pb == nil {
		return nil
	}

	images := make(NoteItemImageList, 0, len(pb.Images))
	for _, img := range pb.Images {
		images = append(images, &NoteItemImage{
			Url:  img.Url,
			Type: int(img.Type),
		})
	}

	return &FeedNoteItem{
		NoteId:   pb.NoteId,
		Title:    pb.Title,
		Desc:     pb.Desc,
		CreateAt: pb.CreatedAt,
		Images:   images,
		Likes:    pb.Likes,
		Author:   pb.Author,
	}
}
