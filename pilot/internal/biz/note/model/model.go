package model

import (
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type NoteStatus string

const (
	NoteStatusPublished = "published"
	NoteStatusAuditing  = "auditing"
	NoteStatusRejected  = "rejected"
	NoteStatusBanned    = "banned"
)

func NoteStatusAsPb(status NoteStatus) notev1.NoteLifeCycleState {
	switch status {
	case NoteStatusPublished:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_PUBLISHED
	case NoteStatusAuditing:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_AUDITING
	case NoteStatusRejected:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_REJECTED
	case NoteStatusBanned:
		return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_BANNED
	}
	return notev1.NoteLifeCycleState_LIFE_CYCLE_STATE_PUBLISHED
}

// LikeNoteReq 点赞笔记请求
type LikeNoteReq struct {
	NoteId imodel.NoteId
	Action imodel.LikeReqAction
}

// GetNoteLikeCountReq 获取笔记点赞数请求
type GetNoteLikeCountReq struct {
	NoteId imodel.NoteId
}

// GetNoteLikeCountRes 获取笔记点赞数响应
type GetNoteLikeCountRes struct {
	NoteId imodel.NoteId
	Count  int64
}

// AddTagReq 添加标签请求
type AddTagReq struct {
	Name string
}

// AddTagRes 添加标签响应
type AddTagRes struct {
	TagId imodel.TagId
}

// SearchTagsReq 搜索标签请求
type SearchTagsReq struct {
	Name string
}

// SearchedTag 搜索到的标签
type SearchedTag struct {
	Id   string
	Name string
}

// ListLikedNotesReq 获取点赞笔记请求
type ListLikedNotesReq struct {
	Uid    int64
	Cursor string
	Count  int32
}

// CreateNoteRes 创建笔记响应
type CreateNoteRes struct {
	NoteId imodel.NoteId
}

// UpdateNoteReq 更新笔记请求
type UpdateNoteReq struct {
	NoteId imodel.NoteId
	Note   *notev1.CreateNoteRequest
}

// UpdateNoteRes 更新笔记响应
type UpdateNoteRes struct {
	NoteId imodel.NoteId
}

// DeleteNoteReq 删除笔记请求
type DeleteNoteReq struct {
	NoteId imodel.NoteId
}

// PageListNotesReq 分页列出笔记请求
type PageListNotesReq struct {
	Page  int32
	Count int32
}

// GetNoteReq 获取笔记请求
type GetNoteReq struct {
	NoteId imodel.NoteId
}

// CheckNoteImagesReq 检查笔记图片请求
type CheckNoteImagesReq struct {
	Images []NoteImage
}

// NoteImage 笔记图片
type NoteImage struct {
	FileId string
	Width  uint32
	Height uint32
	Format string
}

func NoteImageFromPb(pbimg *notev1.NoteImage) NoteImage {
	return NoteImage{
		FileId: pbimg.GetKey(),
		Width:  pbimg.GetMeta().GetWidth(),
		Height: pbimg.GetMeta().GetHeight(),
		Format: pbimg.GetMeta().GetFormat(),
	}
}

func NoteImagesFromPbs(pbimgs []*notev1.NoteImage) []NoteImage {
	images := make([]NoteImage, 0, len(pbimgs))
	for _, pbimg := range pbimgs {
		images = append(images, NoteImageFromPb(pbimg))
	}
	return images
}

// 笔记视频
type NoteVideo struct {
	FileId string
}
