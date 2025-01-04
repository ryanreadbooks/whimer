package model

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/utils"
)

type AssetImageMeta struct {
	Width  uint32 `json:"w"`
	Height uint32 `json:"h"`
	Format string `json:"f"`
}

func NewAssetImageMeta(w, h uint32, format string) *AssetImageMeta {
	return &AssetImageMeta{
		Width:  w,
		Height: h,
		Format: format,
	}
}

func NewAssetImageMetaFromJson(s string) AssetImageMeta {
	var a AssetImageMeta
	_ = json.Unmarshal(utils.StringToBytes(s), &a)
	return a
}

func (r *AssetImageMeta) String() string {
	c, _ := json.Marshal(r)
	return string(c)
}
