package dao

import (
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/infra/dao/record"
	"github.com/ryanreadbooks/whimer/counter/internal/infra/dao/summary"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db sqlx.SqlConn

	RecordRepo  *record.Repo
	SummaryRepo *summary.Repo
}

func MustNew(c *config.Config) *Dao {
	db := sqlx.NewMysql(xsql.GetDsn(
		c.MySql.User,
		c.MySql.Pass,
		c.MySql.Addr,
		c.MySql.DbName,
	))

	// 启动时必须确保数据库有效
	rdb, err := db.RawDB()
	if err != nil {
		panic(err)
	}
	if err = rdb.Ping(); err != nil {
		panic(err)
	}

	r := &Dao{
		db:          db,
		RecordRepo:  record.New(db),
		SummaryRepo: summary.New(db),
	}

	return r
}

var ()

func (d *Dao) DB() sqlx.SqlConn {
	return d.db
}
