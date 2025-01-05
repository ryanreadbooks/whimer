package model

type NotePreviewAssetMetadata struct {
	Key         string `json:"key"`
	Width       uint32 `json:"width"`
	Height      uint32 `json:"height"`
	ContentType string `json:"content_type"`
}

// 笔记资源图片预览图资源表示
type NotePreviewAsset struct {
	Preview NotePreviewAssetMetadata `json:"preview"` // 预览信息
	Default NotePreviewAssetMetadata `json:"default"` // 原始信息
}
