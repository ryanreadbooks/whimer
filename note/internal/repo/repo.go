package repo

import (
	"context"
	"fmt"

	msqlx "github.com/ryanreadbooks/whimer/misc/sqlx"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/repo/note"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db sqlx.SqlConn

	NoteRepo note.NoteModel
}

func New(c *config.Config) *Dao {
	db := sqlx.NewMysql(getDbDsn(
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
		db:       db,
		NoteRepo: note.NewNoteModel(db),
	}
}

func getDbDsn(user, pass, addr, dbName string) string {
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, pass, addr, dbName)
}

func (d *Dao) Transact(ctx context.Context, fns ...msqlx.TransactFunc) error {
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
