package repo

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Repo struct {
	db sqlx.SqlConn

	CommentRepo *comm.Repo
	Bus         *queue.Bus
}

func New(c *config.Config) *Repo {
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
		CommentRepo: comm.New(db),
		Bus:         queue.New(c),
	}

	return r
}

func (d *Repo) DB() sqlx.SqlConn {
	return d.db
}
