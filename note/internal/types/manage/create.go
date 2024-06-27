package manage

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/global"
)

const (
	maxTitleLen = 20
	maxDescLen  = 1000

	PrivacyPublic  = 1
	PrivacyPrivate = 2
)

type CreateReq struct {
	Basic struct {
		Title   string `json:"title"`
		Desc    string `json:"desc"`
		Privacy int    `json:"privacy"`
	} `json:"basic"`
	Images []struct {
		Mime   string `json:"mime,optional"`
		FileId string `json:"file_id"`
		Height int    `json:"height,optional"`
		Weight int    `json:"weight,optional"`
	} `json:"images"`
}

func (r *CreateReq) Validate() error {
	if r == nil {
		return global.ErrArgs.Msg("请求为空")
	}

	if len(r.Images) == 0 {
		return global.ErrArgs.Msg("至少需要包含一张照片")
	}

	title := []rune(r.Basic.Title)
	if len(title) > 20 {
		return global.ErrArgs.Msg(fmt.Sprintf("笔记标题最多%d个字符", maxTitleLen))
	}
	desc := []rune(r.Basic.Desc)
	if len(desc) > maxDescLen {
		return global.ErrArgs.Msg(fmt.Sprintf("笔记描述最多%d个字符", maxTitleLen))
	}

	if r.Basic.Privacy != PrivacyPublic && r.Basic.Privacy != PrivacyPrivate {
		return global.ErrArgs.Msg("笔记权限未知")
	}

	return nil
}

type CreateRes struct {
	NoteId string `json:"note_id"`
}
