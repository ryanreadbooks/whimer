package note

import (
	"context"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/note"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type InteractAdapterImpl struct {
	noteInteractCli notev1.NoteInteractServiceClient
	noteStatCache   *notecache.StatStore
}

func NewInteractAdapterImpl(
	noteInteractCli notev1.NoteInteractServiceClient,
	noteStatCache *notecache.StatStore,
) *InteractAdapterImpl {
	return &InteractAdapterImpl{
		noteInteractCli: noteInteractCli,
		noteStatCache:   noteStatCache,
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

	resp, err := a.noteInteractCli.CheckUserLikeStatus(ctx, req)
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
	resp, err := a.noteInteractCli.BatchCheckUserLikeStatus(ctx, req)
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
	_, err := a.noteInteractCli.LikeNote(ctx, &notev1.LikeNoteRequest{
		Uid:       p.Uid,
		NoteId:    p.NoteId,
		Operation: convert.LikeActionAsPbLikeOperation(p.Action),
	})
	if err != nil {
		return xerror.Wrap(err)
	}
	inc := 0
	switch p.Action {
	case notevo.LikeActionDo:
		inc = 1
	case notevo.LikeActionUndo:
		inc = -1
	}

	if inc != 0 {
		// 同步点赞增量到缓存
		if err := a.noteStatCache.Add(ctx, notecache.NoteLikeCountStat, notecache.NoteStatRepr{
			NoteId: notevo.NoteId(p.NoteId).String(),
			Inc:    int64(inc),
		}); err != nil {
			// log only
			xlog.Msg("note stat add like count failed").Err(err).
				Extras("note_id", p.NoteId, "inc", inc).Errorx(ctx)
		}
	}

	return nil
}

func (a *InteractAdapterImpl) GetLikeCount(ctx context.Context, noteId int64) (int64, error) {
	resp, err := a.noteInteractCli.GetNoteLikes(ctx, &notev1.GetNoteLikesRequest{
		NoteId: noteId,
	})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetLikes(), nil
}

func (a *InteractAdapterImpl) BatchCheckUserLikeStatus(ctx context.Context,
	mappings map[int64][]int64,
) (map[int64]map[int64]bool, error) {
	reqMappings := make(map[int64]*notev1.NoteIdList)
	for uid, noteIds := range mappings {
		reqMappings[uid] = &notev1.NoteIdList{NoteIds: noteIds}
	}

	resp, err := a.noteInteractCli.BatchCheckUserLikeStatus(ctx,
		&notev1.BatchCheckUserLikeStatusRequest{
			Mappings: reqMappings,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	result := make(map[int64]map[int64]bool)
	for uid, statusList := range resp.GetResults() {
		result[uid] = make(map[int64]bool)
		for _, item := range statusList.GetList() {
			result[uid][item.GetNoteId()] = item.GetLiked()
		}
	}

	return result, nil
}
