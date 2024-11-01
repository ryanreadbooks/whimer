package infra

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	infrarepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	repo  *infrarepo.Repo
	cache *redis.Redis
)

func Init(c *config.Config) {
	repo = infrarepo.New(c)
	cache = redis.MustNewRedis(c.Redis)
	dep.Init(c)
}

func Repo() *infrarepo.Repo {
	return repo
}

func Cache() *redis.Redis {
	return cache
}
