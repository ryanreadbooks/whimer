package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/srv/procedure"
)

type Service struct {
	rootCtx    context.Context
	rootCancel context.CancelFunc
	c          *config.Config

	// domain service
	NoteCreatorSrv   *NoteCreatorSrv
	NoteFeedSrv      *NoteFeedSrv
	NoteInteractSrv  *NoteInteractSrv
	NoteProcedureSrv *NoteProcedureSrv
}

// 初始化一个service
func MustNewService(c *config.Config, bizz *biz.Biz, dt *data.Data) *Service {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	s := &Service{
		rootCtx:    rootCtx,
		rootCancel: rootCancel,
		c:          c,
	}

	// 笔记发布流程管理
	procedureMgr, err := procedure.NewManager(c, bizz)
	if err != nil {
		panic(err)
	}

	// 各个子service初始化
	s.NoteCreatorSrv = NewNoteCreatorSrv(s, bizz, procedureMgr)
	s.NoteFeedSrv = NewNoteFeedSrv(s, bizz, dt)
	s.NoteInteractSrv = NewNoteInteractSrv(s, bizz)
	s.NoteProcedureSrv = NewNoteProcedureSrv(c, bizz, procedureMgr)

	return s
}

func (s *Service) Start() {
	s.NoteProcedureSrv.goStartBackgroundHandle(s.rootCtx)
}

func (s *Service) Stop() {
	s.NoteProcedureSrv.StopBackgroundHandle()
}
