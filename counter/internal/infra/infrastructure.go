package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/counter/internal/config"
	infradao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施
var (
	dao      *infradao.Dao
	cache    *redis.Redis
	initOnce sync.Once
)

func Init(c *config.Config) {
	initOnce.Do(func() {
		cache = redis.MustNewRedis(c.Redis)
		dao = infradao.MustNew(c, cache)
	})
}


func Dao() *infradao.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}
