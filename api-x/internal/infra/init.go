package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

var (
	initOnce sync.Once
)

func Init(c *config.Config) {
	initOnce.Do(func() {
		initMisc(c)
		initCache(c)

		InitPassport(c)
		InitNote(c)
		InitCommenter(c)
		InitRelation(c)
		InitMsger(c)
		InitSearch(c)
	})
}
