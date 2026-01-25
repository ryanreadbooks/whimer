package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// 搜索笔记过滤器
type NoteSearchFilter struct {
	Type  string
	Value string
}

// 搜索笔记参数
type SearchNoteParams struct {
	Keyword   string
	PageToken string
	Count     int32
	Filters   []*NoteSearchFilter
}

// 搜索笔记结果
type SearchNoteResult struct {
	NoteIds   []vo.NoteId
	NextToken string
	HasNext   bool
	Total     int64
}

type NoteSearchAdapter interface {
	// 将笔记写入搜索存储中
	AddNote(ctx context.Context, note *entity.SearchNote) error
	// 将笔记从搜索存储中删除
	DeleteNote(ctx context.Context, noteId vo.NoteId) error
	// 搜索笔记
	SearchNote(ctx context.Context, params *SearchNoteParams) (*SearchNoteResult, error)
}
