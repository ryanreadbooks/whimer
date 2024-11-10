package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteInteractSrv struct {
	Ctx *Service

	noteBiz         biz.NoteBiz
	noteInteractBiz biz.NoteInteractBiz
}

func NewNoteInteractSrv(ctx *Service, biz biz.Biz) *NoteInteractSrv {
	s := &NoteInteractSrv{
		Ctx:             ctx,
		noteBiz:         biz.Note,
		noteInteractBiz: biz.Interact,
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

	var op = biz.DoLike
	if in.Operation == notev1.LikeNoteRequest_OPERATION_UNDO_LIKE {
		op = biz.UnDoLike
	}
	err = s.noteInteractBiz.LikeNote(ctx, opUid, in.NoteId, op)
	if err != nil {
		return nil, xerror.Wrapf(err, "srv like note failed").WithExtras("noteId", in.NoteId).WithCtx(ctx)
	}

	return &notev1.LikeNoteResponse{}, nil
}

// 获取笔记点赞数量
func (s *NoteInteractSrv) GetNoteLikes(ctx context.Context, noteId uint64) (uint64, error) {
	return s.noteInteractBiz.GetNoteLikes(ctx, noteId)
}

func (s *NoteInteractSrv) CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error) {
	return s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
}

func (s *NoteInteractSrv) GetNoteInteraction(ctx context.Context, noteId uint64) (*model.UserInteraction, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	liked, err := s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "common check user like status failed")
	}

	return &model.UserInteraction{Liked: liked}, nil
}
