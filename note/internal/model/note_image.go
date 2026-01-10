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

	bucket string `json:"-"` // 非全场景必须字段 用到时手动Set
}

func (i *NoteImage) SetBucket(bucket string) {
	if i == nil {
		return
	}
	i.bucket = bucket
}

func (i *NoteImage) GetBucket() string {
	if i == nil {
		return ""
	}
	return i.bucket
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

func (l NoteImageList) SetBucket(bucket string) {
	for _, img := range l {
		img.SetBucket(bucket)
	}
}
