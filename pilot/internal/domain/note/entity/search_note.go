package entity

import "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

type SearchNote struct {
	NoteId         vo.NoteId // note id
	Title          string
	Desc           string
	CreateAt       int64
	UpdateAt       int64
	AuthorUid      int64
	AuthorNickname string
	TagList        []*SearchedNoteTag
	AssetType      vo.AssetType
	Visibility     vo.Visibility
}
