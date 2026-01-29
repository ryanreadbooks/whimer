package model

import (
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type NoteImageMeta struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

// 笔记图片资源
type NoteImage struct {
	Key  string        `json:"url"`
	Type int           `json:"type"`
	Meta NoteImageMeta `json:"meta"`
}
type NoteImageList []*NoteImage

func (l NoteImageList) AsPb() []*notev1.NoteImage {
	images := make([]*notev1.NoteImage, 0, len(l))
	for _, img := range l {
		images = append(images, &notev1.NoteImage{
			Key:  img.Key,
			Type: int32(img.Type),
			Meta: &notev1.NoteImageMeta{
				Width:  img.Meta.Width,
				Height: img.Meta.Height,
				Format: img.Meta.Format,
			},
		})
	}
	return images
}
