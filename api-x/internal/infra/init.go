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
		initCache(c)
		
		InitPassport(c)
		InitNote(c)
		InitCommenter(c)
		InitRelation(c)
		InitMsger(c)
		InitSearch(c)
	})
}
