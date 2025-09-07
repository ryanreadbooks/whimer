package infra

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	cache *redis.Redis
)

func initCache(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
}

func Cache() *redis.Redis {
	return cache
}
