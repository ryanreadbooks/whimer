package procedure

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// 通用的tx逻辑
type txHelper struct {
	bizz *biz.Biz

	noteProcedureBiz *biz.NoteProcedureBiz
}

func newTxHelper(bizz *biz.Biz) *txHelper {
	return &txHelper{
		bizz:             bizz,
		noteProcedureBiz: bizz.Procedure,
	}
}

func (h *txHelper) txHandleSuccess(
	ctx context.Context,
	noteId int64,
	taskId string,
	protype model.ProcedureType,
	onSuccess onCompleteFunc,
) error {
	err := h.bizz.Tx(ctx, func(ctx context.Context) error {
		needUpdate, err := onSuccess(ctx, noteId, taskId)
		if err != nil {
			return xerror.Wrapf(err, "tx helper on success failed").
				WithExtras("note_id", noteId, "task_id", taskId).
				WithCtx(ctx)
		}

		if needUpdate {
			if err := h.noteProcedureBiz.MarkSuccess(ctx, noteId, protype); err != nil {
				return xerror.Wrapf(err, "tx helper mark record success failed").
					WithExtras("note_id", noteId, "task_id", taskId).
					WithCtx(ctx)
			}
		}

		return nil
	})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (h *txHelper) txHandleFailure(
	ctx context.Context,
	noteId int64,
	taskId string,
	protype model.ProcedureType,
	onFailure onCompleteFunc,
) error {
	err := h.bizz.Tx(ctx, func(ctx context.Context) error {
		needUpdate, err := onFailure(ctx, noteId, taskId)
		if err != nil {
			return xerror.Wrapf(err, "tx helper on failure failed").
				WithExtras("note_id", noteId, "task_id", taskId).
				WithCtx(ctx)
		}

		if needUpdate {
			if err := h.noteProcedureBiz.MarkFailed(ctx, noteId, protype); err != nil {
				return xerror.Wrapf(err, "tx helper mark record failed failed").
					WithExtras("note_id", noteId, "task_id", taskId).
					WithCtx(ctx)
			}
		}

		return nil
	})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
