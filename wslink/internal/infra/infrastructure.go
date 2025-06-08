package infra

import (
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	infradao "github.com/ryanreadbooks/whimer/wslink/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	cache *redis.Redis
	dao   *infradao.Dao
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	dao = infradao.New(cache)

	dep.Init(c)
}

func Dao() *infradao.Dao {
	return dao
}
