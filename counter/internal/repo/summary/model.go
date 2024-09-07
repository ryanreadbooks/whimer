package summary

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Repo struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) *Repo {
	return &Repo{
		db: db,
	}
}

type Model struct {
	BizCode int    `db:"biz_code"`
	Oid     uint64 `db:"oid"`
	Cnt     uint64 `db:"cnt"`
	Ctime   int64  `db:"ctime"`
	Mtime   int64  `db:"mtime"`
}
