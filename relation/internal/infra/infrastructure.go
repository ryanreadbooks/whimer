package infra

import (
	"github.com/ryanreadbooks/whimer/relation/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"

	infradao "github.com/ryanreadbooks/whimer/relation/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dep"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	dao   *infradao.Dao
	cache *redis.Redis
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	dao = infradao.New(c, cache)
	dep.Init(c)
}

func Dao() *infradao.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}

func Close() {
	dao.Close()
}
