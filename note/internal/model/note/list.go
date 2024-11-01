package creator

import (
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type ItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type ItemImageList []*ItemImage

func (l ItemImageList) AsPb() []*notev1.NoteImage {
	images := make([]*notev1.NoteImage, 0, len(l))
	for _, img := range l {
		images = append(images, &notev1.NoteImage{
			Url:  img.Url,
			Type: int32(img.Type),
		})
	}
	return images
}

type Item struct {
	NoteId   uint64        `json:"note_id"`
	Title    string        `json:"title"`
	Desc     string        `json:"desc"`
	Privacy  int8          `json:"privacy"`
	CreateAt int64         `json:"create_at"`
	UpdateAt int64         `json:"update_at"`
	Images   ItemImageList `json:"images"`
	Likes    uint64        `json:"likes"`
}

func (i *Item) AsPb() *notev1.NoteItem {
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

type BatchNoteItem struct {
	Items []*Item `json:"items"`
}

type GetNoteReq struct {
	NoteId uint64 `path:"note_id"`
}
