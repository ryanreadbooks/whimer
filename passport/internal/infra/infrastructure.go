package infra

import (
	"github.com/ryanreadbooks/whimer/misc/encrypt"
	"github.com/ryanreadbooks/whimer/misc/encrypt/aes"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	infradao "github.com/ryanreadbooks/whimer/passport/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 基础设施集合
// 包含持久化、外部依赖等
var (
	dao   *infradao.Dao
	cache *redis.Redis
	enc   encrypt.Encryptor
)

func Init(c *config.Config) {
	cache = redis.MustNewRedis(c.Redis)
	dao = infradao.MustNew(c, cache)
	var err error
	enc, err = aes.NewAes256GCMEncryptor(c.Encrypt.Key, aes.WithMd5Nonce())
	if err != nil {
		panic(err)
	}

	dep.Init(c)
}

func Dao() *infradao.Dao {
	return dao
}

func Cache() *redis.Redis {
	return cache
}

func Encryptor() encrypt.Encryptor {
	return enc
}
