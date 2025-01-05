package dao

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db sqlx.SqlConn

	NoteDao       *NoteDao
	NoteAssetRepo *NoteAssetDao
}

func New(c *config.Config, cache *redis.Redis) *Dao {
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

	return &Dao{
		db:            db,
		NoteDao:       NewNoteDao(db, cache),
		NoteAssetRepo: NewNoteAssetDao(db),
	}
}

// 事务中执行
func (d *Dao) TransactCtx(ctx context.Context, fns ...xsql.TransactFunc) error {
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

func (d *Dao) DB() sqlx.SqlConn {
	return d.db
}

func (d *Dao) Close() {
	d.Close()
}
