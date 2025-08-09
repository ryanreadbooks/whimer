package srv

import "github.com/ryanreadbooks/whimer/search/internal/config"

type Service struct {
	DocumentSrv *DocumentService
}

func NewService(c *config.Config) *Service {
	return &Service{
		DocumentSrv: &DocumentService{},
	}
}
