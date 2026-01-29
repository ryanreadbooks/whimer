package entity

import (
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// 创作者笔记实体
type CreatorNote struct {
	Id         vo.NoteId
	OwnerId    int64
	Title      string
	Desc       string
	Privacy    vo.Visibility
	AssetType  vo.AssetType
	Status     vo.NoteStatus
	Type       vo.NoteType
	Ip         string
	CreateTime int64
	UpdateTime int64
	Images     []*NoteImage
	Videos     []*NoteVideo // 会生成多个视频文件
	AtUsers    mentionvo.AtUserList
	Tags       []*NoteTag

	Likes   int64
	Replies int64
}

type CreateNoteParams struct {
	Title     string
	Desc      string
	Privacy   vo.Visibility
	AssetType vo.AssetType

	Images []*NoteImage
	Videos *NoteVideo

	AtUsers mentionvo.AtUserList
	Tags    []*NoteTag
}

// 补充视频文件目标id
func (p *CreateNoteParams) SetVideoTargetFileId(fileId string) {
	p.Videos.SetTargetFileId(fileId)
}

type UpdateNoteParams struct {
	NoteId int64
	CreateNoteParams
}

// PageListNotesParams 分页查询参数
type PageListNotesParams struct {
	Page   int32
	Count  int32
	Status vo.NoteStatus
}

// PageListNotesResult 分页查询结果
type PageListNotesResult struct {
	Total int64
	Items []*CreatorNote
}
