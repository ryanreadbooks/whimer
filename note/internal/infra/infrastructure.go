package infra

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	infrarepo "github.com/ryanreadbooks/whimer/note/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	dao  *infrarepo.Dao
	cache *redis.Redis
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	dao = infrarepo.New(c, cache)
	dep.Init(c)
}

func Dao() *infrarepo.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}
