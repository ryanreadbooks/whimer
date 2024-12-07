package infra

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/bus"
	infradao "github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dep"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
var (
	dao   *infradao.Dao
	cache *redis.Redis
	queue *bus.Bus
)

func Init(c *config.Config) {
	cache := redis.MustNewRedis(c.Redis) // TODO make it less dependent
	dao = infradao.New(c, cache)
	queue = bus.New(c)
	dep.Init(c)
}

func Dao() *infradao.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}

func Bus() *bus.Bus {
	return queue
}
