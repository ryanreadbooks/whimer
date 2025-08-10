package infra

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	infradao "github.com/ryanreadbooks/whimer/note/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	dao   *infradao.Dao
	cache *redis.Redis
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	dao = infradao.MustNew(c, cache)
	dep.Init(c)
}

func Dao() *infradao.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}

func Close() {
	logx.Info("closing infra")
	dao.Close()
}
