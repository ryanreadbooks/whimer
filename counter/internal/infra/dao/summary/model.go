package summary

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

type Model struct {
	BizCode int32 `db:"biz_code"`
	Oid     int64 `db:"oid"`
	Cnt     int64 `db:"cnt"`
	Ctime   int64 `db:"ctime"`
	Mtime   int64 `db:"mtime"`
}

type PrimaryKey struct {
	BizCode int32 `db:"biz_code"`
	Oid     int64 `db:"oid"`
}

type PrimaryKeyList []PrimaryKey

func (l PrimaryKeyList) Oids() []int64 {
	r := make([]int64, 0, len(l))
	for _, item := range l {
		r = append(r, item.Oid)
	}
	return r
}

func (l PrimaryKeyList) BizCodes() []int32 {
	r := make([]int32, 0, len(l))
	for _, item := range l {
		r = append(r, item.BizCode)
	}
	return r
}

type GetsResult struct {
	BizCode int32 `db:"biz_code"`
	Oid     int64 `db:"oid"`
	Cnt     int64 `db:"cnt"`
}
