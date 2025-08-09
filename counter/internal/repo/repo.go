package repo

import (
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/repo/record"
	"github.com/ryanreadbooks/whimer/counter/internal/repo/summary"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Repo struct {
	db sqlx.SqlConn

	RecordRepo  *record.Repo
	SummaryRepo *summary.Repo
}

func MustNew(c *config.Config) *Repo {
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

	r := &Repo{
		db:          db,
		RecordRepo:  record.New(db),
		SummaryRepo: summary.New(db),
	}

	return r
}

func (d *Repo) DB() sqlx.SqlConn {
	return d.db
}
