package biz

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

const (
	UnDoLike = 0 // 取消点赞
	DoLike   = 1 // 点赞
)

// 笔记互动
type NoteInteractBiz interface {
	// 点赞笔记
	LikeNote(ctx context.Context, uid, noteId uint64, operation int) error
	// 用户是否点赞笔记
	CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error)
	// 批量检查用户是否点赞笔记
	BatchCheckUserLikeStatus(ctx context.Context, uidNoteIds map[uint64][]uint64) (
		map[uint64][]*model.LikeStatus, error)
	// 获取笔记点赞信息并填充
	AssignNoteLikes(ctx context.Context, batch *model.Notes) (*model.Notes, error)
	// 获取笔记点赞数量
	GetNoteLikes(ctx context.Context, noteId uint64) (uint64, error)
	// 获取笔记评论数量
	GetNoteReplyCount(ctx context.Context, noteId uint64) (uint64, error)
	// 获取笔记的评论信息并填充
	AssignNoteReplies(ctx context.Context, batch *model.Notes) (*model.Notes, error)
}

type noteInteractBiz struct {
	noteBiz
}

func NewNoteInteractBiz() NoteInteractBiz {
	b := &noteInteractBiz{}

	return b
}

// 点赞笔记
func (b *noteInteractBiz) LikeNote(ctx context.Context, uid, noteId uint64, operation int) error {
	var (
		err error
	)

	ok, err := b.IsNoteExist(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "GetNoteInteraction check note exists failed")
	}

	if !ok {
		return global.ErrNoteNotFound
	}

	if operation == UnDoLike {
		// 取消点赞
		_, err = dep.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     uid,
			Oid:     noteId,
		})
	} else {
		// 点赞
		_, err = dep.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     uid,
			Oid:     noteId,
		})
	}

	if err != nil {
		return xerror.Wrapf(err, "counter add record failed").
			WithExtra("op", operation).
			WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return nil
}

// 获取用户是否点赞过笔记
func (b *noteInteractBiz) CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error) {
	ok, err := b.IsNoteExist(ctx, noteId)
	if err != nil {
		return false, xerror.Wrapf(err, "GetNoteInteraction check note exists failed")
	}

	if !ok {
		return false, global.ErrNoteNotFound
	}

	resp, err := dep.GetCounter().GetRecord(ctx, &counterv1.GetRecordRequest{
		BizCode: global.NoteLikeBizcode,
		Uid:     uid,
		Oid:     noteId,
	})
	if err != nil {
		return false, xerror.Wrapf(err, "CheckUserLikeStatus counter get record failed").
			WithExtra("noteId", noteId).
			WithExtra("user", uid).
			WithCtx(ctx)
	}

	return resp.GetRecord().GetAct() == counterv1.RecordAct_RECORD_ACT_ADD, nil
}

// 批量获取用户是否点赞过笔记
// 批量查找就不检查noteId是否存在
func (b *noteInteractBiz) BatchCheckUserLikeStatus(ctx context.Context, uidNoteIds map[uint64][]uint64) (
	map[uint64][]*model.LikeStatus, error) {

	var req = make(map[uint64]*counterv1.ObjectList)
	for uid, noteIds := range uidNoteIds {
		req[uid] = &counterv1.ObjectList{
			Oids: noteIds,
		}
	}
	resp, err := dep.GetCounter().BatchGetRecord(ctx, &counterv1.BatchGetRecordRequest{
		BizCode: global.NoteLikeBizcode,
		Params:  req,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "note interact biz failed to batch like status").WithCtx(ctx)
	}

	var result = make(map[uint64][]*model.LikeStatus, len(resp.GetResults()))
	for uid, noteIds := range uidNoteIds {
		likeRecords := resp.GetResults()[uid]
		for _, noteId := range noteIds {
			hasLiked := false
			for _, likeRecord := range likeRecords.GetList() {
				if likeRecord.Oid == noteId && likeRecord.Act == counterv1.RecordAct_RECORD_ACT_ADD {
					hasLiked = true
				}
			}

			result[uid] = append(result[uid], &model.LikeStatus{
				NoteId: noteId,
				Liked:  hasLiked,
			})
		}
	}

	return result, nil
}

func (b *noteInteractBiz) AssignNoteLikes(ctx context.Context, batch *model.Notes) (*model.Notes, error) {
	var (
		notes   = batch.Items
		noteIds = make([]uint64, 0, len(notes))
		reqs    = make([]*counterv1.GetSummaryRequest, 0, len(notes))
	)

	for _, note := range notes {
		noteIds = append(noteIds, note.NoteId)
		reqs = append(reqs, &counterv1.GetSummaryRequest{
			BizCode: global.NoteLikeBizcode,
			Oid:     note.NoteId,
		})
	}

	// 获取点赞数量
	resp, err := dep.GetCounter().BatchGetSummary(ctx,
		&counterv1.BatchGetSummaryRequest{
			Requests: reqs,
		})
	if err != nil {
		// 仅打印日志不返回error
		xlog.Msg("counter failed to batch get summary").
			Err(err).
			Extra("note_ids", noteIds).
			Infox(ctx)
	}

	if resp != nil {
		m := make(map[uint64]uint64, len(resp.Responses))
		for _, r := range resp.Responses {
			m[r.Oid] = r.Count
		}
		// 赋值
		for _, item := range batch.Items {
			if likeCnt, ok := m[item.NoteId]; ok {
				item.Likes = likeCnt
			}
		}
	}

	return batch, nil
}

// 获取笔记点赞数量
func (b *noteInteractBiz) GetNoteLikes(ctx context.Context, noteId uint64) (uint64, error) {
	ok, err := b.IsNoteExist(ctx, noteId)
	if err != nil {
		return 0, xerror.Wrapf(err, "GetNoteInteraction check note exists failed")
	}

	if !ok {
		return 0, global.ErrNoteNotFound
	}

	resp, err := dep.GetCounter().GetSummary(ctx, &counterv1.GetSummaryRequest{
		BizCode: global.NoteLikeBizcode,
		Oid:     noteId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "counter get summary failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.Count, nil
}

func (b *noteInteractBiz) GetNoteReplyCount(ctx context.Context, noteId uint64) (uint64, error) {
	ok, err := b.IsNoteExist(ctx, noteId)
	if err != nil {
		return 0, xerror.Wrapf(err, "GetNoteInteraction check note exists failed")
	}

	if !ok {
		return 0, global.ErrNoteNotFound
	}

	resp, err := dep.GetCommenter().CountReply(ctx, &commentv1.CountReplyRequest{
		Oid: noteId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "commenter count reply failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.NumReply, nil
}

func (b *noteInteractBiz) AssignNoteReplies(ctx context.Context, batch *model.Notes) (*model.Notes, error) {
	var (
		noteIds = batch.GetIds()
	)

	resp, err := dep.GetCommenter().BatchCountReply(ctx, &commentv1.BatchCountReplyRequest{
		Oids: noteIds,
	})
	if err != nil {
		xlog.Msg("counter failed to batch count reply").
			Err(err).
			Extra("note_ids", noteIds).
			Infox(ctx)
	}

	if resp != nil {
		m := make(map[uint64]uint64, len(resp.Numbers))
		for nid, rcnt := range resp.Numbers {
			m[nid] = rcnt
		}

		for _, note := range batch.Items {
			note.Replies = m[note.NoteId]
		}
	}

	return batch, nil
}
