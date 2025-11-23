package dao

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/database"
	kafkadao "github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka"
)

var (
	databaseDao *database.Dao
)

func Init(c *config.Config) {
	kafkadao.Init(c)
	databaseDao = database.MustNew(c)
}

func Close() {
	kafkadao.Close()
	databaseDao.DB().Close()
}

func Database() *database.Dao {
	return databaseDao
}
