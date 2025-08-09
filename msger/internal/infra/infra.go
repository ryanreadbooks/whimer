package infra

import (
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	infradao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
var (
	dao   *infradao.Dao
	cache *redis.Redis
)

func Init(c *config.Config) {
	// cache := redis.MustNewRedis(c.Redis)
	dao = infradao.MustNew(c)
}

func Dao() *infradao.Dao {
	return dao
}
