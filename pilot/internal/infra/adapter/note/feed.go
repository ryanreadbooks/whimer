package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"
)

type NoteFeedAdapterImpl struct {
	noteInteractServer notev1.NoteInteractServiceClient
	noteFeedServer     notev1.NoteFeedServiceClient
}

var _ repository.NoteFeedAdapter = (*NoteFeedAdapterImpl)(nil)

func NewNoteFeedAdapterImpl(
	noteFeedServer notev1.NoteFeedServiceClient,
	noteInteractServer notev1.NoteInteractServiceClient,
) *NoteFeedAdapterImpl {
	return &NoteFeedAdapterImpl{
		noteFeedServer:     noteFeedServer,
		noteInteractServer: noteInteractServer,
	}
}

func (a *NoteFeedAdapterImpl) RandomGet(ctx context.Context, count int32) ([]*entity.FeedNote, error) {
	resp, err := a.noteFeedServer.RandomGet(ctx,
		&notev1.RandomGetRequest{
			Count: count,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	items := resp.GetItems()
	if len(items) == 0 {
		return []*entity.FeedNote{}, nil
	}

	return convert.BatchPbFeedNotesToEntities(items), nil
}

func (a *NoteFeedAdapterImpl) GetNote(ctx context.Context, noteId int64) (*entity.FeedNote, *entity.FeedNoteExt, error) {
	resp, err := a.noteFeedServer.GetFeedNote(ctx,
		&notev1.GetFeedNoteRequest{
			NoteId: noteId,
		})
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	return convert.PbFeedNoteToEntity(resp.GetItem()), convert.PbFeedNoteExtToEntity(resp.GetExt()), nil
}

func (a *NoteFeedAdapterImpl) BatchGetNotes(ctx context.Context, noteIds []int64) (map[int64]*entity.FeedNote, error) {
	resp, err := a.noteFeedServer.BatchGetFeedNotes(ctx,
		&notev1.BatchGetFeedNotesRequest{
			NoteIds: noteIds,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	results := make(map[int64]*entity.FeedNote)
	resps := resp.GetResult()
	for noteId, resp := range resps {
		results[noteId] = convert.PbFeedNoteToEntity(resp.GetItem())
	}
	return results, nil
}

func (a *NoteFeedAdapterImpl) BatchCheckNoteExist(ctx context.Context, noteIds []int64) (map[int64]bool, error) {
	resp, err := a.noteFeedServer.BatchCheckFeedNoteExist(ctx,
		&notev1.BatchCheckFeedNoteExistRequest{
			NoteIds: noteIds,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}
	return resp.GetExistence(), nil
}

func (a *NoteFeedAdapterImpl) ListUserNote(ctx context.Context,
	uid int64, cursor int64, count int32) (
	[]*entity.FeedNote, *repository.CursorPageResult, error,
) {
	resp, err := a.noteFeedServer.ListFeedByUid(ctx,
		&notev1.ListFeedByUidRequest{
			Uid:    uid,
			Cursor: cursor,
			Count:  count,
		})
	if err != nil {
		return nil, nil, xerror.Wrap(err)
	}

	items := resp.GetItems()
	if len(items) == 0 {
		return []*entity.FeedNote{}, nil, nil
	}

	return convert.BatchPbFeedNotesToEntities(items), &repository.CursorPageResult{
		NextCursor: resp.GetNextCursor(),
		HasNext:    resp.GetHasNext(),
	}, nil
}

func (a *NoteFeedAdapterImpl) GetNoteAuthorUid(ctx context.Context, noteId int64) (int64, error) {
	resp, err := a.noteFeedServer.GetNoteAuthor(ctx,
		&notev1.GetNoteAuthorRequest{
			NoteId: noteId,
		})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetAuthor(), nil
}

func (a *NoteFeedAdapterImpl) ListUserLikedNote(ctx context.Context,
	uid int64, cursor string, count int32) (
	[]*entity.FeedNote, *repository.CursorPageResultV2, error,
) {
	resp, err := a.noteInteractServer.PageListUserLikedNote(ctx,
		&notev1.PageListUserLikedNoteRequest{
			Uid:    uid,
			Cursor: cursor,
			Count:  count,
		})
	if err != nil {
		return nil, &repository.CursorPageResultV2{}, xerror.Wrap(err)
	}

	items := resp.GetItems()
	if len(items) == 0 {
		return []*entity.FeedNote{}, &repository.CursorPageResultV2{}, nil
	}

	return convert.BatchPbFeedNotesToEntities(items),
		&repository.CursorPageResultV2{
			NextCursor: resp.GetNextCursor(),
			HasNext:    resp.GetHasNext(),
		}, nil
}

func (a *NoteFeedAdapterImpl) GetPublicPostedCount(ctx context.Context, uid int64) (int64, error) {
	resp, err := a.noteFeedServer.GetPublicPostedCount(ctx, &notev1.GetPublicPostedCountRequest{
		Uid: uid,
	})
	if err != nil {
		return 0, xerror.Wrap(err)
	}
	return resp.GetCount(), nil
}

func (a *NoteFeedAdapterImpl) GetUserRecentPost(ctx context.Context, uid int64, count int32) ([]*entity.RecentPost, error) {
	resp, err := a.noteFeedServer.GetUserRecentPost(ctx, &notev1.GetUserRecentPostRequest{
		Uid:   uid,
		Count: count,
	})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	posts := make([]*entity.RecentPost, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		if len(item.Images) > 0 {
			posts = append(posts, &entity.RecentPost{
				NoteId: vo.NoteId(item.NoteId),
				Cover:  convert.NewNoteImageItemUrlPrv(item.Images[0]),
				Type:   convert.PbNoteTypeToVoNoteType(item.NoteType),
			})
		}
	}

	return posts, nil
}
