package model

import (
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
)

type NoteImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type NoteImage struct {
	Url    string        `json:"url"`
	Type   int           `json:"type"`
	Meta   NoteImageMeta `json:"meta"`
	UrlPrv string        `json:"url_prv"`
}

type NoteImageList []*NoteImage

func (l NoteImageList) AsPb() []*notev1.NoteImage {
	images := make([]*notev1.NoteImage, 0, len(l))
	for _, img := range l {
		images = append(images, &notev1.NoteImage{
			Url:    img.Url,
			Type:   int32(img.Type),
			UrlPrv: img.UrlPrv,
			Meta: &notev1.NoteImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
				Format: img.Meta.Format,
			},
		})
	}
	return images
}

type NoteTag struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Ctime int64  `json:"ctime"`
}

func (t *NoteTag) AsPb() *notev1.NoteTag {
	if t == nil {
		return nil
	}

	return &notev1.NoteTag{
		Id:    t.Id,
		Name:  t.Name,
		Ctime: t.Ctime,
	}
}

func NoteTagListAsPb(tags []*NoteTag) []*notev1.NoteTag {
	if len(tags) == 0 {
		return nil
	}

	var r []*notev1.NoteTag
	for _, t := range tags {
		r = append(r, t.AsPb())
	}

	return r
}

type Note struct {
	NoteId   int64         `json:"note_id"`
	Title    string        `json:"title"`
	Desc     string        `json:"desc"`
	Privacy  int8          `json:"privacy,omitempty"`
	CreateAt int64         `json:"create_at,omitempty"`
	UpdateAt int64         `json:"update_at,omitempty"`
	Ip       string        `json:"ip"`
	Images   NoteImageList `json:"images"`
	Likes    int64         `json:"likes"`   // 点赞数
	Replies  int64         `json:"replies"` // 评论数
	Tags     []*NoteTag    `json:"tags,omitempty"`

	// unexported to user
	Owner int64 `json:"-"`
}

func (n *Note) AsSlice() []*Note {
	return []*Note{n}
}

func NoteFromDao(d *notedao.Note) *Note {
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
	n.Ip = xnet.BytesIpAsString(d.Ip)
	n.Owner = d.Owner

	return n
}

func NoteSliceFromDao(ds []*notedao.Note) []*Note {
	notes := make([]*Note, 0, len(ds))
	for _, n := range ds {
		notes = append(notes, NoteFromDao(n))
	}
	return notes
}

func (i *Note) AsPb() *notev1.NoteItem {
	res := &notev1.NoteItem{
		NoteId:   i.NoteId,
		Title:    i.Title,
		Desc:     i.Desc,
		Privacy:  int32(i.Privacy),
		CreateAt: i.CreateAt,
		UpdateAt: i.UpdateAt,
		Ip:       i.Ip,
		Images:   i.Images.AsPb(),
		Likes:    i.Likes,
		Replies:  i.Replies,
		Owner:    i.Owner,
	}

	if len(i.Tags) > 0 {
		for _, t := range i.Tags {
			if t != nil {
				res.Tags = append(res.Tags, t.AsPb())
			}
		}
	}

	return res
}

// 转换成pb并隐藏一些不公开的属性
func (i *Note) AsFeedPb() *notev1.FeedNoteItem {
	return &notev1.FeedNoteItem{
		NoteId:    i.NoteId,
		Title:     i.Title,
		Desc:      i.Desc,
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

type NoteExt struct {
	TagIds []int64
}

func (e *NoteExt) SetTagIds(s string) {
	e.TagIds = xslice.SplitInts[int64](s, ",")
}

type GetNoteReq struct {
	NoteId int64 `path:"note_id"`
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
