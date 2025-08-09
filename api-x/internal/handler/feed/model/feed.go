package model

import (
	imodel "github.com/ryanreadbooks/whimer/api-x/internal/model"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
)

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
	Followed  bool `json:"followed"`  // 用户是否关注了笔记作者
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
	Uid      int64  `json:"uid"`
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

// 浏览返回的笔记结构
type FeedNoteItem struct {
	NoteId   imodel.NoteId     `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Images   NoteItemImageList `json:"images"`
	Likes    int64             `json:"likes"` // 笔记总点赞数

	// 下面这些字段要单独设置 不从note grpc接口中拿
	Author   *Author     `json:"author"`   // 作者信息
	Comments int64       `json:"comments"` // 笔记总评论数
	Interact Interaction `json:"interact"` // 当前请求的用户与该笔记的交互记录，比如点赞、评论、收藏等动作
}

// 笔记详情返回的结构 更加详细的信息
type FullFeedNoteItem struct {
	*FeedNoteItem

	// 更多信息
	TagList []*imodel.NoteTag `json:"tag_list,omitempty"`
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
		NoteId:   imodel.NoteId(pb.NoteId),
		Title:    pb.Title,
		Desc:     pb.Desc,
		CreateAt: pb.CreatedAt,
		UpdateAt: pb.UpdatedAt,
		Images:   images,
		Likes:    pb.Likes,
	}
}
