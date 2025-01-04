package infra

import (
	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/asset-job/internal/infra/oss"
)

func Init(c *config.Config) {
	oss.Init(c)
}
