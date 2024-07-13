package creator

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/global"
)

const (
	maxTitleLen = 20
	maxDescLen  = 1000
	maxImageLen = 9

	PrivacyPublic  = global.PrivacyPublic
	PrivacyPrivate = global.PrivacyPrivate
)

type CreateReq struct {
	Basic struct {
		Title   string `json:"title"`
		Desc    string `json:"desc"`
		Privacy int    `json:"privacy"`
	} `json:"basic"`
	Images []struct {
		FileId string `json:"file_id"`
	} `json:"images"`
}

func (r *CreateReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if len(r.Images) == 0 {
		return global.ErrArgs.Msg("至少需要包含一张照片")
	}

	if len(r.Images) > maxImageLen {
		return global.ErrArgs.Msg(fmt.Sprintf("最多上传%d张图片", maxImageLen))
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
		return global.ErrArgs.Msg("笔记参数权限不支持")
	}

	return nil
}

type CreateRes struct {
	NoteId string `json:"note_id"`
}
