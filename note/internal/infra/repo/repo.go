package repo

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"
	"github.com/ryanreadbooks/whimer/note/internal/infra/repo/noteasset"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Repo struct {
	db sqlx.SqlConn

	NoteRepo      *note.Repo
	NoteAssetRepo *noteasset.Repo
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

	return &Repo{
		db:            db,
		NoteRepo:      note.New(db),
		NoteAssetRepo: noteasset.New(db),
	}
}

// 事务中执行
func (d *Repo) Transact(ctx context.Context, fns ...xsql.TransactFunc) error {
	return d.db.TransactCtx(ctx, func(ctx context.Context, s sqlx.Session) error {
		for _, fn := range fns {
			err := fn(ctx, s)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (d *Repo) DB() sqlx.SqlConn {
	return d.db
}
