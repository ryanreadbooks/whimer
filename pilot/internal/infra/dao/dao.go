package dao

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/database"
	kafkadao "github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	databaseDao *database.Dao
)

func Init(c *config.Config, r *redis.Redis) {
	kafkadao.Init(c)
	databaseDao = database.MustNew(c, r)
}

func Close() {
	kafkadao.Close()
	databaseDao.DB().Close()
}

func Database() *database.Dao {
	return databaseDao
}
