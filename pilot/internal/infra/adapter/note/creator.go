package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

// implement domain/note/repository/creator
var _ repository.NoteCreatorAdapter = &CreatorAdapterImpl{}

type CreatorAdapterImpl struct {
	cli notev1.NoteCreatorServiceClient
}

func NewCreatorAdapterImpl(cli notev1.NoteCreatorServiceClient) *CreatorAdapterImpl {
	return &CreatorAdapterImpl{cli: cli}
}

// 获取笔记
func (a *CreatorAdapterImpl) GetNote(ctx context.Context, noteId int64) (*entity.CreatorNote, error) {
	req := &notev1.GetNoteRequest{NoteId: noteId}
	resp, err := a.cli.GetNote(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return convert.PbNoteToCreatorEntity(resp.GetNote()), nil
}

// 新建笔记
func (a *CreatorAdapterImpl) CreateNote(ctx context.Context, params *entity.CreateNoteParams) (int64, error) {
	req := convert.EntityCreateNoteParamsAsPb(params)
	resp, err := a.cli.CreateNote(ctx, req)
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetNoteId(), nil
}

// 更新笔记
func (a *CreatorAdapterImpl) UpdateNote(ctx context.Context, params *entity.UpdateNoteParams) (int64, error) {
	req := &notev1.UpdateNoteRequest{
		NoteId: params.NoteId,
		Note:   convert.EntityCreateNoteParamsAsPb(&params.CreateNoteParams),
	}
	resp, err := a.cli.UpdateNote(ctx, req)
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetNoteId(), nil
}

// 删除笔记
func (a *CreatorAdapterImpl) DeleteNote(ctx context.Context, noteId int64) error {
	req := &notev1.DeleteNoteRequest{NoteId: noteId}
	_, err := a.cli.DeleteNote(ctx, req)
	if err != nil {
		return xerror.Wrap(err)
	}
	return nil
}

// 分页获取笔记列表
func (a *CreatorAdapterImpl) PageListNotes(ctx context.Context, params *entity.PageListNotesParams) (*entity.PageListNotesResult, error) {
	req := &notev1.PageListNoteRequest{
		Page:           params.Page,
		Count:          params.Count,
		LifeCycleState: convert.NoteStatusToLifeCycleState(params.Status),
	}
	resp, err := a.cli.PageListNote(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &entity.PageListNotesResult{
		Total: resp.GetTotal(),
		Items: convert.BatchPbNotesToCreatorEntities(resp.GetItems()),
	}, nil
}
