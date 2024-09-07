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

type Model struct {
	BizCode int    `db:"biz_code"`
	Uid     uint64 `db:"uid"`
	Oid     uint64 `db:"oid"`
	Act     int8   `db:"act"`
	Ctime   int64  `db:"ctime"`
	Mtime   int64  `db:"mtime"`
}
