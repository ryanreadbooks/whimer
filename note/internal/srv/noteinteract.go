package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type NoteInteractSrv struct {
	Ctx *Service

	noteBiz         *biz.NoteBiz
	noteInteractBiz *biz.NoteInteractBiz
	noteEventBiz    *biz.NoteEventBiz
}

func NewNoteInteractSrv(ctx *Service, biz *biz.Biz) *NoteInteractSrv {
	s := &NoteInteractSrv{
		Ctx:             ctx,
		noteBiz:         biz.Note,
		noteInteractBiz: biz.Interact,
		noteEventBiz:    biz.NoteEvent,
	}

	return s
}

// 点赞笔记
func (s *NoteInteractSrv) LikeNote(ctx context.Context, in *notev1.LikeNoteRequest) (*notev1.LikeNoteResponse, error) {
	var (
		opUid = metadata.Uid(ctx)
		err   error
	)

	if opUid != in.Uid {
		return nil, xerror.ErrPermission
	}

	note, err := s.noteBiz.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", in.NoteId).WithCtx(ctx)
	}

	// 非公开笔记 其它人不能点赞
	if note.Privacy != model.PrivacyPublic && opUid != note.Owner {
		return nil, global.ErrNoteNotPublic
	}

	if note.State != model.NoteStatePublished { // 未发布的笔记不能点赞
		return nil, global.ErrNoteNotFound
	}

	op := biz.DoLike
	if in.Operation == notev1.LikeNoteRequest_OPERATION_UNDO_LIKE {
		op = biz.UnDoLike
	}
	err = s.noteInteractBiz.LikeNote(ctx, opUid, in.NoteId, op)
	if err != nil {
		return nil, xerror.Wrapf(err, "srv like note failed").WithExtras("noteId", in.NoteId).WithCtx(ctx)
	}

	isLiked := op == biz.DoLike
	if err := s.noteEventBiz.NoteLiked(ctx, in.NoteId, opUid, note.Owner, isLiked); err != nil {
		xlog.Msg("srv like note publish event failed").
			Err(err).
			Extras("note_id", in.NoteId, "user_id", opUid, "is_liked", isLiked).
			Errorx(ctx)
	}

	return &notev1.LikeNoteResponse{}, nil
}

// 获取笔记点赞数量
func (s *NoteInteractSrv) GetNoteLikes(ctx context.Context, noteId int64) (int64, error) {
	uid := metadata.Uid(ctx)

	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return 0, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy != model.PrivacyPublic && uid != note.Owner {
		return 0, global.ErrNoteNotPublic
	}

	return s.noteInteractBiz.GetNoteLikes(ctx, noteId)
}

func (s *NoteInteractSrv) CheckUserLikeStatus(ctx context.Context, uid int64, noteId int64) (bool, error) {
	return s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
}

func (s *NoteInteractSrv) BatchCheckUserLikeStatus(ctx context.Context, req map[int64][]int64) (
	map[int64][]*model.LikeStatus, error,
) {
	// 批量查找就不检查noteId是否存在
	uidsCheckStatus, err := s.noteInteractBiz.BatchCheckUserLikeStatus(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "batch check user like status failed").WithExtra("req", req).WithCtx(ctx)
	}

	return uidsCheckStatus, nil
}

// 获取笔记的交互信息
func (s *NoteInteractSrv) GetNoteInteraction(ctx context.Context, noteId int64) (*model.UserInteraction, error) {
	uid := metadata.Uid(ctx)

	// 非公开笔记非作者不能获取互动状态
	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy != model.PrivacyPublic && uid != note.Owner {
		return nil, global.ErrNoteNotPublic
	}

	liked, err := s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "common check user like status failed")
	}

	return &model.UserInteraction{Liked: liked}, nil
}

// 列出用户点赞过的笔记
func (s *NoteInteractSrv) PageListUserLikedNoteIds(ctx context.Context, in *notev1.PageListUserLikedNoteRequest) (
	[]int64, model.PageResultV2, error,
) {
	pageRes := model.PageResultV2{}
	resp, err := dep.GetCounter().PageGetUserRecord(ctx,
		&counterv1.PageGetUserRecordRequest{
			BizCode:  global.NoteLikeBizcode,
			Uid:      in.Uid,
			Cursor:   in.Cursor,
			Count:    in.Count,
			SortRule: counterv1.SortRule_SORT_RULE_DESC,
		})
	if err != nil {
		return nil, pageRes, xerror.Wrapf(err, "counter page get user record failed").WithCtx(ctx)
	}

	noteIds := make([]int64, 0, len(resp.Items))
	for _, item := range resp.Items {
		noteIds = append(noteIds, item.Oid)
	}
	pageRes.NextCursor = resp.NextCursor
	pageRes.HasNext = resp.HasNext

	return noteIds, pageRes, nil
}
