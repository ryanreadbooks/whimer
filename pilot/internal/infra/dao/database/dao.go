package database

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/database/usersetting"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db             *xsql.DB
	UserSettingDao *usersetting.Dao
}

func (d *Dao) DB() *xsql.DB {
	return d.db
}

func (d *Dao) Transact(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.Transact(ctx, fn)
}

func MustNew(c *config.Config) *Dao {
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
		UserSettingDao: usersetting.NewDao(db),
	}
}
