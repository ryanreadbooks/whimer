package creator

import "github.com/ryanreadbooks/whimer/note/sdk"

type ListResItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type ListResItemImageList []*ListResItemImage

func (l ListResItemImageList) AsPb() []*sdk.GetNoteResImage {
	images := make([]*sdk.GetNoteResImage, 0, len(l))
	for _, img := range l {
		images = append(images, &sdk.GetNoteResImage{
			Url:  img.Url,
			Type: int32(img.Type),
		})
	}
	return images
}

type ListResItem struct {
	NoteId   uint64               `json:"note_id"`
	Title    string               `json:"title"`
	Desc     string               `json:"desc"`
	Privacy  int8                 `json:"privacy"`
	CreateAt int64                `json:"create_at"`
	UpdateAt int64                `json:"update_at"`
	Images   ListResItemImageList `json:"images"`
}

func (i *ListResItem) AsPb() *sdk.NoteItem {
	return &sdk.NoteItem{
		NoteId:   i.NoteId,
		Title:    i.Title,
		Desc:     i.Desc,
		Privacy:  int32(i.Privacy),
		CreateAt: i.CreateAt,
		UpdateAt: i.UpdateAt,
		Images:   i.Images.AsPb(),
	}
}

type ListRes struct {
	Items []*ListResItem `json:"items"`
}

type GetNoteReq struct {
	NoteId uint64 `path:"note_id"`
}
