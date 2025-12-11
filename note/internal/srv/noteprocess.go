package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
)

type NoteProcessSrv struct {
	bizz biz.Biz
}

func NewNoteProcessSrv(biz biz.Biz) *NoteProcessSrv {
	return &NoteProcessSrv{bizz: biz}
}

func (s *NoteProcessSrv) Process(ctx context.Context, taskId string) error {
	record, err := s.bizz.Process.GetRecordByTaskId(ctx, taskId)
	if err != nil {
		// 只打日志
		xlog.Msg("process biz get task failed").
			Err(err).
			Extras("taskId", taskId).
			Errorx(ctx)
		return nil
	}

	// TODO
	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		err := s.bizz.Creator.SetNoteStateProcessed(ctx, record.NoteId)
		if err != nil {
			return xerror.Wrapf(err, "creator set note state processed failed").
				WithExtra("noteId", record.NoteId).
				WithCtx(ctx)
		}

		// 设置成功
		err = s.bizz.Process.MarkRecordSuccess(ctx, taskId)
		if err != nil {
			return xerror.Wrapf(err, "process biz mark record success failed").
				WithExtra("taskId", taskId).
				WithCtx(ctx)
		}

		return nil
	})

	if err != nil {
		// log only
		xlog.Msg("process biz tx failed").
			Err(err).
			Extras("taskId", taskId).
			Errorx(ctx)

		return nil
	}

	xlog.Msgf("process biz tx success").
		Extras("taskId", taskId, "noteId", record.NoteId).
		Infox(ctx)

	return nil
}
