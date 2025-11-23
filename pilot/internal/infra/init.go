package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
)

var (
	initOnce sync.Once
)

func Init(c *config.Config) {
	initOnce.Do(func() {
		initCache(c)
		dao.Init(c, Cache())
		dep.Init(c)
		initMisc(c)
	})
}

func Close() {
	dao.Close()
}
