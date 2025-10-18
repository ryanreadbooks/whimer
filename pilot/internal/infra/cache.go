package infra

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	infracache "github.com/ryanreadbooks/whimer/pilot/internal/infra/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	cache *redis.Redis
)

func initCache(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	infracache.Init(c, cache)
}

func Cache() *redis.Redis {
	return cache
}
