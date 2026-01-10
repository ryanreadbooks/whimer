package dao

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db *xsql.DB

	NamespaceDao   *NamespaceDao
	TaskDao        *TaskDao
	TaskHistoryDao *TaskHistoryDao
}

func MustNew(c *config.Config, cache *redis.Redis) *Dao {
	sqlx.DisableStmtLog()

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
		db:             db,
		NamespaceDao:   NewNamespaceDao(db, cache),
		TaskDao:        NewTaskDao(db),
		TaskHistoryDao: NewTaskHistoryDao(db),
	}
}

func (d *Dao) DB() *xsql.DB {
	return d.db
}

func (d *Dao) Transact(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.Transact(ctx, fn)
}
