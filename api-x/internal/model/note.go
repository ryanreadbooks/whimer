package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type NoteItemImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
}

type NoteItemImage struct {
	Url    string            `json:"url"`
	Type   int               `json:"type"`
	Meta   NoteItemImageMeta `json:"meta"`
	UrlPrv string            `json:"url_prv"`
}

type NoteItemImageList []*NoteItemImage

// 包含发起请求的用户和该笔记的交互记录
type Interaction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}

type NoteTag struct {
	Id   TagId  `json:"id"`
	Name string `json:"name"`
}

func NoteTagFromPb(t *notev1.NoteTag) *NoteTag {
	return &NoteTag{
		Id:   TagId(t.Id),
		Name: t.Name,
	}
}

func NoteTagsFromPbs(ts []*notev1.NoteTag) []*NoteTag {
	if len(ts) == 0 {
		return []*NoteTag{}
	}

	var r = make([]*NoteTag, 0, len(ts))
	for _, t := range ts {
		r = append(r, NoteTagFromPb(t))
	}
	return r
}

type AdminNoteItem struct {
	NoteId   NoteId            `json:"note_id"`
	Title    string            `json:"title"`
	Desc     string            `json:"desc"`
	Privacy  int8              `json:"privacy"`
	CreateAt int64             `json:"create_at"`
	UpdateAt int64             `json:"update_at"`
	Ip       string            `json:"-"`
	IpLoc    string            `json:"ip_loc"`
	Images   NoteItemImageList `json:"images"`
	Likes    int64             `json:"likes"`
	Replies  int64             `json:"replies"`
	Interact Interaction       `json:"interact"`
	TagList  []*NoteTag        `json:"tag_list,omitempty"`
}

func NewAdminNoteItemFromPb(pb *notev1.NoteItem) *AdminNoteItem {
	if pb == nil {
		return nil
	}

	images := make(NoteItemImageList, 0, len(pb.Images))
	for _, img := range pb.Images {
		images = append(images, &NoteItemImage{
			Url:    img.Url,
			Type:   int(img.Type),
			UrlPrv: img.UrlPrv,
			Meta: NoteItemImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
			},
		})
	}

	var tagList []*NoteTag = NoteTagsFromPbs(pb.GetTags())
	ctx := context.Background()
	ipLoc, _ := infra.Ip2Loc().Convert(ctx, pb.Ip)
	return &AdminNoteItem{
		NoteId:   NoteId(pb.NoteId),
		Title:    pb.Title,
		Desc:     pb.Desc,
		Privacy:  int8(pb.Privacy),
		CreateAt: pb.CreateAt,
		UpdateAt: pb.UpdateAt,
		Images:   images,
		Likes:    pb.Likes,
		Replies:  pb.Replies,
		TagList:  tagList,
		Ip:       pb.Ip,
		IpLoc:    ipLoc,
	}
}

// 点赞/取消点赞
type LikeReqAction uint8

const (
	LikeReqActionUndo LikeReqAction = 0
	LikeReqActionDo   LikeReqAction = 1
)
