package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/infra"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index"
)

type SearchService struct {
}

// page and count are ignored for now
func (s *SearchService) SearchNoteTags(ctx context.Context, text string, page, count int32) ([]*index.NoteTag, int64, error) {
	// 限制只能拿第一页的30条数据
	resp, total, err := infra.EsDao().NoteTagIndexer.Search(ctx, text, 1, 30)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "failed to search note tags").WithExtras(
			"text", text,
			"page", page,
			"count", count).WithCtx(ctx)
	}

	return resp, total, nil
}

func (s *SearchService) SearchNotes(ctx context.Context, in *searchv1.SearchNotesRequest) (*index.NoteIndexSearchResult, error) {
	// filter
	filters := make([]index.SearchNoteOption, 0, len(in.Filters))
	for _, filter := range in.Filters {
		filterValue := filter.Value
		switch filter.Type {
		case searchv1.NoteFilterType_filter_sort_rule:
			filters = append(filters, index.WithSearchNoteFilterNoteType(filterValue))
		case searchv1.NoteFilterType_filter_note_type:
			filters = append(filters, index.WithSearchNoteSortBy(filterValue))
		case searchv1.NoteFilterType_filter_note_pubtime:
			filters = append(filters, index.WithSearchNotePubTime(filterValue))
		}
	}

	res, err := infra.EsDao().NoteIndexer.Search(ctx,
		in.Keyword, in.PageToken, in.Count,
		filters...,
	)
	if err != nil {
		return nil, xerror.Wrapf(err, "failed to search notes").WithExtras(
			"keyword", in.Keyword,
			"page_token", in.PageToken,
			"count", in.Count).WithCtx(ctx)
	}

	return res, nil
}
