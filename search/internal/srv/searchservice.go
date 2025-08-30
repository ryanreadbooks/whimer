package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/search/internal/infra"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index"
)

type SearchService struct{
}

// page and count are ignored for now
func (s *SearchService) SearchNoteTags(ctx context.Context, text string, page, count int32) ([]*index.NoteTag, int64, error) {
	// 限制只能拿第一页的30条数据
	return infra.EsDao().NoteTagIndexer.Search(ctx, text, 1, 30)
}

func (s *SearchService) SearchNotes(ctx context.Context, keyword, pageToken string, count int32) (*index.NoteIndexSearchResult, error) {
	return infra.EsDao().NoteIndexer.Search(ctx, keyword, pageToken, count)
}