package srv

import (
	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/asset-job/internal/infra"
)

type Service struct {
	Config *config.Config

	NoteImageService *NoteImageService
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{
		Config: c,
	}

	// 基础设施初始化
	infra.Init(c)

	s.NoteImageService = NewNoteImageService()

	return s
}

func Close() {
	infra.Close()
}

type AsService struct{}

func (s AsService) Start() {}
func (s AsService) Stop()  { Close() }
