package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
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
func NewService(c *config.Config, bizz biz.Biz) *Service {
	rootCtx, rootCancel := context.WithCancel(context.Background())
	s := &Service{
		rootCtx:    rootCtx,
		rootCancel: rootCancel,
		c:          c,
	}

	// 各个子service初始化
	s.NoteCreatorSrv = NewNoteCreatorSrv(s, bizz)
	s.NoteFeedSrv = NewNoteFeedSrv(s, bizz)
	s.NoteInteractSrv = NewNoteInteractSrv(s, bizz)
	s.NoteProcedureSrv = NewNoteProcedureSrv(c, bizz)
	return s
}

func (s *Service) Start() {
	s.NoteProcedureSrv.goStartBackgroundHandle(s.rootCtx)
}

func (s *Service) Stop() {
	s.NoteProcedureSrv.StopBackgroundHandle()
}
