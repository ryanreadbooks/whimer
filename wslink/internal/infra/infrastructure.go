package infra

import (
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	cache   *redis.Redis
	sessDao *dao.SessionDao
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	sessDao = dao.NewSessionDao(cache)

	dep.Init(c)
}

func SessDao() *dao.SessionDao {
	return sessDao
}
