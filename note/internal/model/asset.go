package model

import (
	"encoding/json"
)

type AssetImageMeta struct {
	Width  uint32            `json:"w"`
	Height uint32            `json:"h"`
	Format string            `json:"f"`
	Extra  map[string]string `json:"ext,omitempty"`
}

func NewAssetImageMeta(w, h uint32, format string) *AssetImageMeta {
	return &AssetImageMeta{
		Width:  w,
		Height: h,
		Format: format,
	}
}

func NewAssetImageMetaFromJson(s []byte) AssetImageMeta {
	var a AssetImageMeta
	_ = json.Unmarshal(s, &a)
	return a
}

func (r *AssetImageMeta) String() string {
	c, _ := json.Marshal(r)
	return string(c)
}

func (r *AssetImageMeta) Bytes() []byte {
	c, _ := json.Marshal(r)
	return c
}

type AssetPreviewEventMetadata struct {
	Key         string `json:"key"`
	Width       uint32 `json:"width"`
	Height      uint32 `json:"height"`
	ContentType string `json:"content_type"`
}

// kafka中消息
type AssetPreviewEvent struct {
	Preview AssetPreviewEventMetadata `json:"preview"` // 预览信息
	Default AssetPreviewEventMetadata `json:"default"` // 原始信息
}
