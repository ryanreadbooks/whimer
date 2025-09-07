package srv

import (
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
)

type Service struct {
	Config *config.Config

	// domain service
	NoteCreatorSrv  *NoteCreatorSrv
	NoteFeedSrv     *NoteFeedSrv
	NoteInteractSrv *NoteInteractSrv
}

// 初始化一个service
func NewService(c *config.Config, bizz biz.Biz) *Service {
	s := &Service{
		Config: c,
	}

	// 各个子service初始化
	s.NoteCreatorSrv = NewNoteCreatorSrv(s, bizz)
	s.NoteFeedSrv = NewNoteFeedSrv(s, bizz)
	s.NoteInteractSrv = NewNoteInteractSrv(s, bizz)

	return s
}

func Close() {
}

type AsService struct{}

func (AsService) Start() {}
func (AsService) Stop()  { Close() }
