package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
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

	note, err := s.noteBiz.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", in.NoteId).WithCtx(ctx)
	}

	// 非公开笔记 其它人不能点赞
	if note.Privacy != global.PrivacyPublic && opUid != note.Owner {
		return nil, global.ErrNoteNotPublic
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
	var (
		uid = metadata.Uid(ctx)
	)

	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return 0, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy != global.PrivacyPublic && uid != note.Owner {
		return 0, global.ErrNoteNotPublic
	}

	return s.noteInteractBiz.GetNoteLikes(ctx, noteId)
}

func (s *NoteInteractSrv) CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error) {
	return s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
}

// 获取笔记的交互信息
func (s *NoteInteractSrv) GetNoteInteraction(ctx context.Context, noteId uint64) (*model.UserInteraction, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	// 非公开笔记非作者不能获取互动状态
	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get public note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy != global.PrivacyPublic && uid != note.Owner {
		return nil, global.ErrNoteNotPublic
	}

	liked, err := s.noteInteractBiz.CheckUserLikeStatus(ctx, uid, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "common check user like status failed")
	}

	return &model.UserInteraction{Liked: liked}, nil
}
