package svc

import (
	"context"

	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteInteractSvc struct {
	Ctx *ServiceContext
}

func NewNoteInteractSvc(ctx *ServiceContext) *NoteInteractSvc {
	s := &NoteInteractSvc{
		Ctx: ctx,
	}

	return s
}

// 点赞笔记
func (s *NoteInteractSvc) LikeNote(ctx context.Context, in *notev1.LikeNoteRequest) (*notev1.LikeNoteResponse, error) {
	var (
		opUid = metadata.Uid(ctx)
		err   error
	)

	if opUid != in.Uid {
		return nil, xerror.ErrPermission
	}

	if ok, err := IsNoteExist(ctx, in.NoteId); err != nil || !ok {
		return nil, xerror.Wrapf(err, "check note exist failed").WithCtx(ctx)
	}

	if in.Operation == notev1.LikeNoteRequest_OPERATION_UNDO_LIKE {
		// 取消点赞
		_, err = dep.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     in.Uid,
			Oid:     in.NoteId,
		})
	} else {
		// 点赞
		_, err = dep.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     in.Uid,
			Oid:     in.NoteId,
		})
	}

	if err != nil {
		return nil, xerror.Wrapf(err, "counter add record failed").
			WithExtra("op", in.Operation).
			WithExtra("noteId", in.NoteId).WithCtx(ctx)
	}

	return &notev1.LikeNoteResponse{}, nil
}

// 获取笔记点赞数量
func (s *NoteInteractSvc) GetNoteLikes(ctx context.Context, noteId uint64) (uint64, error) {
	if ok, err := IsNoteExist(ctx, noteId); err != nil || !ok {
		return 0, xerror.Wrapf(err, "check note exist failed").WithCtx(ctx)
	}

	resp, err := dep.GetCounter().GetSummary(ctx, &counterv1.GetSummaryRequest{
		BizCode: global.NoteLikeBizcode,
		Oid:     noteId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "counter add record failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.Count, nil
}

func (s *NoteInteractSvc) CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error) {
	if ok, err := IsNoteExist(ctx, noteId); err != nil || !ok {
		// 笔记不存在当作没有点赞过
		return false, xerror.Wrapf(err, "check user like status check note exists failed")
	}

	if liked, err := CheckUserLikeStatus(ctx, uid, noteId); err != nil {
		return false, xerror.Wrapf(err, "common check user like status failed")
	} else {
		return liked, nil
	}
}

func (s *NoteInteractSvc) GetNoteInteraction(ctx context.Context, noteId uint64) (*notemodel.UserInteraction, error) {
	ok, err := IsNoteExist(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "GetNoteInteraction check note exists failed")
	}

	if !ok {
		return nil, global.ErrNoteNotFound
	}

	var (
		uid = metadata.Uid(ctx)
	)

	liked, err := CheckUserLikeStatus(ctx, uid, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "common check user like status failed")
	}

	return &notemodel.UserInteraction{Liked: liked}, nil
}
