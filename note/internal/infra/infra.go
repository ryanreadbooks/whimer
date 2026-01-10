package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/infra/etcd"
	"github.com/ryanreadbooks/whimer/note/internal/infra/kafka"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	once  sync.Once
	dt    *data.Data
	cache *redis.Redis

	etcdCli *etcd.Client
)

func Init(conf *config.Config) {
	once.Do(func() {
		etcdCli = etcd.MustNew(conf)
		cache = redis.MustNewRedis(conf.Redis)
		kafka.Init(conf)
		dt = data.MustNew(conf, cache, kafka.GetPublisher())
		dep.Init(conf)
	})
}

// Data 获取数据层实例
func Data() *data.Data {
	return dt
}

func Cache() *redis.Redis {
	return cache
}

func Close() {
	logx.Info("closing infra")
	kafka.Close()
	dt.Close()
}

func Etcd() *etcd.Client {
	return etcdCli
}
