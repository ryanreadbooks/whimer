package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

var _ repository.NoteSearchAdapter = &NoteSearchAdapterImpl{}

type NoteSearchAdapterImpl struct {
	docCli    searchv1.DocumentServiceClient
	searchCli searchv1.SearchServiceClient
}

func NewNoteSearchAdapterImpl(
	noteSearchCli searchv1.DocumentServiceClient,
	noteSearcher searchv1.SearchServiceClient,
) *NoteSearchAdapterImpl {
	return &NoteSearchAdapterImpl{
		docCli:    noteSearchCli,
		searchCli: noteSearcher,
	}
}

// 写入es
func (a *NoteSearchAdapterImpl) AddNote(ctx context.Context, note *entity.SearchNote) error {
	req := &searchv1.BatchAddNoteRequest{
		Notes: []*searchv1.Note{convert.EntitySearchNoteToPb(note)},
	}
	_, err := a.docCli.BatchAddNote(ctx, req)
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

// 从es中移除
func (a *NoteSearchAdapterImpl) DeleteNote(ctx context.Context, noteId notevo.NoteId) error {
	req := &searchv1.BatchDeleteNoteRequest{
		Ids: []string{noteId.String()},
	}

	_, err := a.docCli.BatchDeleteNote(ctx, req)
	if err != nil {
		return xerror.Wrap(err)
	}

	return err
}

// 搜索笔记
func (a *NoteSearchAdapterImpl) SearchNote(ctx context.Context, params *repository.SearchNoteParams) (*repository.SearchNoteResult, error) {
	filters := make([]*searchv1.NoteFilter, 0, len(params.Filters))
	for _, f := range params.Filters {
		filters = append(filters, &searchv1.NoteFilter{
			Type:  searchv1.NoteFilterType(searchv1.NoteFilterType_value[f.Type]),
			Value: f.Value,
		})
	}

	resp, err := a.searchCli.SearchNotes(ctx, &searchv1.SearchNotesRequest{
		Keyword:   params.Keyword,
		PageToken: params.PageToken,
		Count:     params.Count,
		Filters:   filters,
	})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	noteIds := make([]notevo.NoteId, 0, len(resp.GetNoteIds()))
	for _, n := range resp.GetNoteIds() {
		var id notevo.NoteId
		if err := id.UnmarshalText([]byte(n)); err == nil {
			noteIds = append(noteIds, id)
		}
	}

	return &repository.SearchNoteResult{
		NoteIds:   noteIds,
		NextToken: resp.GetNextToken(),
		HasNext:   resp.GetHasNext(),
		Total:     resp.GetTotal(),
	}, nil
}
