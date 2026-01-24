package biz

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

const (
	maxTitleLen = 20
	maxDescLen  = 1000
	maxImageLen = 9
	maxTagCount = 10
)

type CreateNoteRequestBasic struct {
	Title    string         `json:"title"`
	Desc     string         `json:"desc"`
	Privacy  model.Privacy  `json:"privacy"`
	NoteType model.NoteType `json:"note_type"`
}

type CreateNoteRequestImage struct {
	FileId string `json:"file_id"`
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
}

type CreateNoteRequestVideo struct {
	FileId       string `json:"file_id"`
	TargetFileId string `json:"target_file_id"`
	CoverFileId  string `json:"cover_file_id"`
}

func (r *CreateNoteRequestVideo) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.FileId == "" || r.TargetFileId == "" {
		return global.ErrArgs.Msg("未指定视频资源")
	}
	if r.CoverFileId == "" {
		return global.ErrArgs.Msg("未指定封面")
	}

	return nil
}

type CreateNoteRequest struct {
	Basic   CreateNoteRequestBasic   `json:"basic"`
	Images  []CreateNoteRequestImage `json:"images"`
	Video   *CreateNoteRequestVideo  `json:"video,omitempty"`
	TagIds  []int64                  `json:"tag_ids"`
	AtUsers []*model.AtUser          `json:"at_users"`
}

func (r *CreateNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
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

	if r.Basic.Privacy != model.PrivacyPublic && r.Basic.Privacy != model.PrivacyPrivate {
		return global.ErrArgs.Msg("笔记参数权限不支持")
	}

	noteType := model.NoteType(r.Basic.NoteType)
	switch noteType {
	case model.AssetTypeImage:
		if len(r.Images) == 0 {
			return global.ErrArgs.Msg("至少需要包含一张照片")
		}
	case model.AssetTypeVideo:
		if r.Video == nil {
			return global.ErrArgs.Msg("未包含视频")
		}
	default:
		return global.ErrArgs.Msg("笔记资源类型不支持")
	}

	if len(r.TagIds) >= maxTagCount {
		return global.ErrArgs.Msg(fmt.Sprintf("笔记最多支持%d个标签", maxTagCount))
	}

	r.AtUsers = model.FilterInvalidAtUsers(r.AtUsers)

	return nil
}

type CreateNoteResponse struct {
	NoteId string `json:"note_id"`
}

type UpdateNoteRequest struct {
	NoteId int64 `json:"note_id"`
	CreateNoteRequest
}

func (r *UpdateNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.NoteId == 0 {
		return global.ErrArgs.Msg("笔记不存在")
	}

	return r.CreateNoteRequest.Validate()
}

type DeleteNoteRequest struct {
	NoteId int64 `json:"note_id"`
}

func (r *DeleteNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.NoteId <= 0 {
		return global.ErrArgs.Msg("笔记不存在")
	}

	return nil
}

type GetNoteReq struct {
	NoteId int64 `path:"note_id"`
}
