package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv/procedure"
)

// 负责笔记状态流转
type NoteProcedureSrv struct {
	c *config.Config

	// 各业务逻辑
	bizz             *biz.Biz
	noteBiz          *biz.NoteBiz
	noteProcedureBiz *biz.NoteProcedureBiz
	noteCreatorBiz   *biz.NoteCreatorBiz

	// 流程管理器
	procedureMgr *procedure.Manager
}

func NewNoteProcedureSrv(
	c *config.Config,
	biz *biz.Biz,
	procedureMgr *procedure.Manager,
) *NoteProcedureSrv {
	return &NoteProcedureSrv{
		c: c,

		bizz:             biz,
		noteBiz:          biz.Note,
		noteProcedureBiz: biz.Procedure,
		noteCreatorBiz:   biz.Creator,

		procedureMgr: procedureMgr,
	}
}

type HandleAssetProcessResultReq struct {
	NoteId      int64
	TaskId      string
	Success     bool
	Videos      []*model.VideoAsset
	ErrorOutput []byte
}

// HandleCallbackAssetProcedureResult 调度任务完成后回调处理逻辑
// 此处需要将状态标记为已成功并且进入下一流程
func (s *NoteProcedureSrv) HandleCallbackAssetProcedureResult(
	ctx context.Context,
	req *HandleAssetProcessResultReq,
) error {
	if req.Success {
		return s.procedureMgr.CompleteAssetSuccess(ctx, req.NoteId, req.TaskId, req.Videos)
	} else {
		return s.procedureMgr.CompleteAssetFailure(ctx, req.NoteId, req.TaskId, req.ErrorOutput)
	}
}

// goStartBackgroundHandle 启动后台处理
func (s *NoteProcedureSrv) goStartBackgroundHandle(ctx context.Context) {
	s.procedureMgr.StartRetryLoop(ctx)
}

// StopBackgroundHandle 停止后台处理
func (s *NoteProcedureSrv) StopBackgroundHandle() {
	s.procedureMgr.StopRetryLoop()
}
