package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/p2p"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db *xsql.DB

	P2PChatDao *p2p.ChatDao
}

func New(c *config.Config, cache *redis.Redis) *Dao {
	conn := sqlx.NewMysql(xsql.GetDsn(
		c.MySql.User,
		c.MySql.Pass,
		c.MySql.Addr,
		c.MySql.DbName,
	))

	// 启动时必须确保数据库有效
	rdb, err := conn.RawDB()
	if err != nil {
		panic(err)
	}
	if err = rdb.Ping(); err != nil {
		panic(err)
	}

	db := xsql.New(conn)
	return &Dao{
		db:         db,
		P2PChatDao: p2p.NewChatDao(db),
	}
}

func (d *Dao) DB() *xsql.DB {
	return d.db
}
