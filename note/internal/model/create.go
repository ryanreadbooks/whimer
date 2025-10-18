package model

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/global"
)

const (
	maxTitleLen = 20
	maxDescLen  = 1000
	maxImageLen = 9
	maxTagCount = 10
)

const (
	// 笔记可见范围
	PrivacyPublic  = global.PrivacyPublic
	PrivacyPrivate = global.PrivacyPrivate
)

type CreateNoteRequestBasic struct {
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Privacy int8   `json:"privacy"`
}

type CreateNoteRequestImage struct {
	FileId string `json:"file_id"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type CreateNoteRequest struct {
	Basic   CreateNoteRequestBasic   `json:"basic"`
	Images  []CreateNoteRequestImage `json:"images"`
	TagIds  []int64                  `json:"tag_ids"`
	AtUsers []*AtUser                `json:"at_users"`
}

func (r *CreateNoteRequest) Validate() error {
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

	if len(r.TagIds) >= maxTagCount {
		return global.ErrArgs.Msg(fmt.Sprintf("笔记最多支持%d个标签", maxTagCount))
	}

	r.AtUsers = FilterInvalidAtUsers(r.AtUsers)

	return nil
}

type CreateNoteResponse struct {
	NoteId string `json:"note_id"`
}
