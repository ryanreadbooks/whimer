package profile

import (
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
)

type Service struct {
	c    *config.Config
	repo *repo.Repo
}

func New(c *config.Config, repo *repo.Repo) *Service {
	return &Service{
		c:    c,
		repo: repo,
	}
}
