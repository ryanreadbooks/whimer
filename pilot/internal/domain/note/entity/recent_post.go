package entity

import notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

// 最近发布
type RecentPost struct {
	NoteId notevo.NoteId
	Type   notevo.NoteType
	Cover  string
}
