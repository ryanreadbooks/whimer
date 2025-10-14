package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
)

var (
	initOnce sync.Once
)

func Init(c *config.Config) {
	initOnce.Do(func() {
		initMisc(c)
		initCache(c)

		dep.Init(c)
	})
}
