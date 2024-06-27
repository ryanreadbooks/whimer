package repo

import (
	"fmt"

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
