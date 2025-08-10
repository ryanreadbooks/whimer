package infra

import (
	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao"
)

var (
	esDao *esdao.EsDao
)

func Init(c *config.Config) {
	esDao = esdao.MustNew(c)
	esDao.MustInit(c)
}

func EsDao() *esdao.EsDao {
	return esDao
}
