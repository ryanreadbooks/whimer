package srv

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/feed/internal/infra"
)

type Service struct {
}

func Init(c *config.Config) *Service {
	infra.Init(c)

	return &Service{}
}
