package record

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Repo struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) *Repo {
	return &Repo{
		db: db,
	}
}

const (
	ActUnspecified = 0
	ActDo          = 1
	ActUndo        = 2
)

type Model struct {
	Id      uint64 `db:"id"`
	BizCode int    `db:"biz_code"`
	Uid     uint64 `db:"uid"`
	Oid     uint64 `db:"oid"`
	Act     int8   `db:"act"`
	Ctime   int64  `db:"ctime"`
	Mtime   int64  `db:"mtime"`
}

type Summary struct {
	BizCode int    `db:"biz_code"`
	Oid     uint64 `db:"oid"`
	Cnt     uint64 `db:"cnt"`
}
