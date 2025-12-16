package model

import (
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type Note struct {
	NoteId   int64         `json:"note_id"`
	Title    string        `json:"title"`
	Desc     string        `json:"desc"`
	Privacy  Privacy       `json:"privacy,omitempty"`
	Type     NoteType      `json:"type"`
	State    NoteState     `json:"state"` // 笔记状态
	CreateAt int64         `json:"create_at,omitempty"`
	UpdateAt int64         `json:"update_at,omitempty"`
	Ip       string        `json:"ip"`
	Images   NoteImageList `json:"images"`
	Videos   *NoteVideo    `json:"videos"`  // 包含多种编码的视频资源
	Likes    int64         `json:"likes"`   // 点赞数
	Replies  int64         `json:"replies"` // 评论数

	// ext字段
	Tags    []*NoteTag `json:"tags,omitempty"`
	AtUsers []*AtUser  `json:"at_users,omitempty"`

	// unexported to user
	Owner int64 `json:"-"`
}

func (n *Note) AsSlice() []*Note {
	return []*Note{n}
}

func (i *Note) AsPb() *notev1.NoteItem {
	res := &notev1.NoteItem{
		NoteId:   i.NoteId,
		Title:    i.Title,
		Desc:     i.Desc,
		Privacy:  int32(i.Privacy),
		State:    notev1.NoteState(i.State),
		NoteType: notev1.NoteAssetType(i.Type),
		CreateAt: i.CreateAt,
		UpdateAt: i.UpdateAt,
		Ip:       i.Ip,
		Images:   i.Images.AsPb(),
		Likes:    i.Likes,
		Replies:  i.Replies,
		Owner:    i.Owner,
	}

	// note tags
	res.Tags = NoteTagListAsPb(i.Tags)
	// at_users
	res.AtUsers = AtUsersAsPb(i.AtUsers)

	return res
}

// 转换成pb并隐藏一些不公开的属性
func (i *Note) AsFeedPb() *notev1.FeedNoteItem {
	return &notev1.FeedNoteItem{
		NoteId:    i.NoteId,
		Title:     i.Title,
		Desc:      i.Desc,
		NoteType:  notev1.NoteAssetType(i.Type),
		CreatedAt: i.CreateAt,
		UpdatedAt: i.UpdateAt,
		Images:    i.Images.AsPb(),
		Ip:        i.Ip,
		Likes:     i.Likes,
		Author:    i.Owner,
		Replies:   i.Replies,
	}
}

type Notes struct {
	Items []*Note `json:"items"`
}

func (n *Notes) GetIds() []int64 {
	r := make([]int64, 0, len(n.Items))
	for _, item := range n.Items {
		r = append(r, item.NoteId)
	}
	return r
}

func PbFeedNoteItemsFromNotes(ns *Notes) []*notev1.FeedNoteItem {
	items := make([]*notev1.FeedNoteItem, 0, len(ns.Items))
	for _, item := range ns.Items {
		items = append(items, item.AsFeedPb())
	}

	return items
}

// 每个用户和笔记的交互情况
type UserInteraction struct {
	Liked bool // 是否点赞
}

func (u *UserInteraction) AsPb() *notev1.NoteInteraction {
	return &notev1.NoteInteraction{
		Liked: u.Liked,
	}
}

type LikeStatus struct {
	NoteId int64
	Liked  bool
}

type InteractStatus struct {
	NoteId    int64
	Liked     bool
	Starred   bool
	Commented bool
}
