package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/relation/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db *xsql.DB

	RelationDao *RelationDao
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
		db:          db,
		RelationDao: NewRelationDao(db, cache),
	}
}
