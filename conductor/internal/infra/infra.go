package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	infradao "github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra/etcd"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	dao     *infradao.Dao
	cache   *redis.Redis
	etcdCli *etcd.Client
	once    sync.Once
)

func Init(c *config.Config) {
	once.Do(func() {
		etcdCli = etcd.MustNew(c)
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

func Close() {
	dao.DB().Close()
}

func Etcd() *etcd.Client {
	return etcdCli
}
