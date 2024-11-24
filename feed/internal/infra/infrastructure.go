package infra

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/feed/internal/infra/dep"
)

func Init(c *config.Config) {
	dep.Init(c)
}
