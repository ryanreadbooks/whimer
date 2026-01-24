package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"
)

type InteractAdapterImpl struct {
	cli notev1.NoteInteractServiceClient
}

func NewInteractAdapterImpl(c notev1.NoteInteractServiceClient) *InteractAdapterImpl {
	return &InteractAdapterImpl{
		cli: c,
	}
}

var _ repository.NoteLikesAdapter = &InteractAdapterImpl{}

func (a *InteractAdapterImpl) GetLikeStatus(ctx context.Context, p *repository.GetLikeStatusParams) (
	*repository.GetLikeStatusResult, error,
) {
	req := &notev1.CheckUserLikeStatusRequest{
		NoteId: p.NoteId,
		Uid:    p.Uid,
	}

	resp, err := a.cli.CheckUserLikeStatus(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &repository.GetLikeStatusResult{Liked: resp.GetLiked()}, nil
}

func (a *InteractAdapterImpl) BatchGetLikeStatus(
	ctx context.Context, p *repository.BatchGetLikeStatusParams,
) (*repository.BatchGetLikeStatusResult, error) {
	mappings := make(map[int64]*notev1.NoteIdList)
	mappings[p.Uid] = &notev1.NoteIdList{
		NoteIds: p.NoteIds,
	}
	req := &notev1.BatchCheckUserLikeStatusRequest{
		Mappings: mappings,
	}
	resp, err := a.cli.BatchCheckUserLikeStatus(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	result := resp.GetResults()[p.Uid]
	liked := make(map[int64]bool, len(result.GetList()))
	for _, lst := range result.GetList() {
		liked[lst.GetNoteId()] = lst.GetLiked()
	}

	return &repository.BatchGetLikeStatusResult{
		Liked: liked,
	}, nil
}

func (a *InteractAdapterImpl) LikeNote(ctx context.Context, p *repository.LikeNoteParams) error {
	_, err := a.cli.LikeNote(ctx, &notev1.LikeNoteRequest{
		Uid:       p.Uid,
		NoteId:    p.NoteId,
		Operation: convert.LikeActionAsPbLikeOperation(p.Action),
	})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (a *InteractAdapterImpl) GetLikeCount(ctx context.Context, noteId int64) (int64, error) {
	resp, err := a.cli.GetNoteLikes(ctx, &notev1.GetNoteLikesRequest{
		NoteId: noteId,
	})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetLikes(), nil
}
