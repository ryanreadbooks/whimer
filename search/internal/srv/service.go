package srv

import "github.com/ryanreadbooks/whimer/search/internal/config"

type Service struct {
	DocumentSrv *DocumentService
	SearchSrv   *SearchService
}

func NewService(c *config.Config) *Service {
	return &Service{
		DocumentSrv: NewDocumentService(),
		SearchSrv:   &SearchService{},
	}
}

func (s *Service) Stop() {
	s.DocumentSrv.Close()
}
