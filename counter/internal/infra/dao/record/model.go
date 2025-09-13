package record

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Repo struct {
	db sqlx.SqlConn

	cache *Cache
}

func New(db sqlx.SqlConn, cache *Cache) *Repo {
	return &Repo{
		db:    db,
		cache: cache,
	}
}

const (
	ActUnspecified = 0
	ActDo          = 1
	ActUndo        = 2
)

type Record struct {
	Id      int64 `db:"id"`
	BizCode int32 `db:"biz_code"`
	Uid     int64 `db:"uid"`
	Oid     int64 `db:"oid"`
	Act     int8  `db:"act"`
	Ctime   int64 `db:"ctime"`
	Mtime   int64 `db:"mtime"`
}

type Summary struct {
	BizCode int32 `db:"biz_code"`
	Oid     int64 `db:"oid"`
	Cnt     int64 `db:"cnt"`
}
