package model

import (
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
)

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}

type NoteItemImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type NoteItemImage struct {
	Url      string            `json:"url"`
	Type     int               `json:"type"`
	UrlPrv   string            `json:"url_prv"`
	Metadata NoteItemImageMeta `json:"metadata"`
}

type NoteItemImageList []*NoteItemImage

type Author struct {
	Uid      uint64 `json:"uid"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func NewAuthor(u *userv1.UserInfo) *Author {
	return &Author{
		Uid:      u.Uid,
		Nickname: u.Nickname,
		Avatar:   u.Avatar,
	}
}

type FeedNoteItem struct {
	NoteId   uint64            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	CreateAt int64             `json:"create_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    uint64            `json:"likes"` // 笔记总点赞数

	// 下面这些字段要额外设置
	Author   *Author     `json:"author"`   // 作者信息
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
			Url:    img.Url,
			Type:   int(img.Type),
			UrlPrv: img.UrlPrv,
			Metadata: NoteItemImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
				Format: img.Meta.Format,
			},
		})
	}

	return &FeedNoteItem{
		NoteId:   pb.NoteId,
		Title:    pb.Title,
		Desc:     pb.Desc,
		CreateAt: pb.CreatedAt,
		Images:   images,
		Likes:    pb.Likes,
	}
}
