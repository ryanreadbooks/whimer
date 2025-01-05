package srv

import (
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
)

type Service struct {
	Config *config.Config

	// domain service
	NoteCreatorSrv  *NoteCreatorSrv
	NoteFeedSrv     *NoteFeedSrv
	NoteInteractSrv *NoteInteractSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{
		Config: c,
	}

	// 基础设施初始化
	infra.Init(c)
	// 业务初始化
	biz := biz.New()
	// 各个子service初始化
	s.NoteCreatorSrv = NewNoteCreatorSrv(s, biz)
	s.NoteFeedSrv = NewNoteFeedSrv(s, biz)
	s.NoteInteractSrv = NewNoteInteractSrv(s, biz)

	return s
}

func Close() {
	infra.Close()
}

type AsService struct{}

func (AsService) Start() {}
func (AsService) Stop()  { Close() }
