package entity

import (
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// 信息流中的笔记基础信息 用于笔记列表、笔记详情等场景
type FeedNote struct {
	Id        notevo.NoteId
	Title     string
	Desc      string
	CreateAt  int64
	UpdateAt  int64
	Images    []*NoteImage
	Videos    []*NoteVideo
	Likes     int64
	Comments  int64
	Ip        string
	Type      notevo.NoteType
	AuthorUid int64
}

type FeedNoteExt struct {
	Tags    []*NoteTag
	AtUsers mentionvo.AtUserList
}
