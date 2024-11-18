package model

import (
	"github.com/ryanreadbooks/whimer/note/internal/infra/dao"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type NoteImageList []*NoteImage

func (l NoteImageList) AsPb() []*notev1.NoteImage {
	images := make([]*notev1.NoteImage, 0, len(l))
	for _, img := range l {
		images = append(images, &notev1.NoteImage{
			Url:  img.Url,
			Type: int32(img.Type),
		})
	}
	return images
}

type Note struct {
	NoteId          uint64          `json:"note_id"`
	Title           string          `json:"title"`
	Desc            string          `json:"desc"`
	Privacy         int8            `json:"privacy,omitempty"`
	CreateAt        int64           `json:"create_at,omitempty"`
	UpdateAt        int64           `json:"update_at,omitempty"`
	Images          NoteImageList   `json:"images"`
	Likes           uint64          `json:"likes"`
	UserInteraction UserInteraction `json:"user_interaction,omitempty"`

	// unexported to user
	Owner uint64 `json:"-"`
}

func (n *Note) AsSlice() []*Note {
	return []*Note{n}
}

func NoteFromDao(d *dao.Note) *Note {
	n := &Note{}
	if d == nil {
		return n
	}
	n.NoteId = d.Id
	n.Title = d.Title
	n.Desc = d.Desc
	n.Privacy = d.Privacy
	n.CreateAt = d.CreateAt
	n.UpdateAt = d.UpdateAt
	n.Owner = d.Owner

	return n
}

func NoteSliceFromDao(ds []*dao.Note) []*Note {
	notes := make([]*Note, 0, len(ds))
	for _, n := range ds {
		notes = append(notes, NoteFromDao(n))
	}
	return notes
}

func (i *Note) AsPb() *notev1.NoteItem {
	return &notev1.NoteItem{
		NoteId:   i.NoteId,
		Title:    i.Title,
		Desc:     i.Desc,
		Privacy:  int32(i.Privacy),
		CreateAt: i.CreateAt,
		UpdateAt: i.UpdateAt,
		Images:   i.Images.AsPb(),
		Likes:    i.Likes,
	}
}

// 转换成pb并隐藏一些不公开的属性
func (i *Note) AsFeedPb() *notev1.FeedNoteItem {
	return &notev1.FeedNoteItem{
		NoteId:      i.NoteId,
		Title:       i.Title,
		Desc:        i.Desc,
		CreatedAt:   i.CreateAt,
		Images:      i.Images.AsPb(),
		Likes:       i.Likes,
		Interaction: i.UserInteraction.AsPb(),
	}
}

type Notes struct {
	Items []*Note `json:"items"`
}

type GetNoteReq struct {
	NoteId uint64 `path:"note_id"`
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
