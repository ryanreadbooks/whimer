package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	lockPOTableName = "conductor_lock"
)

var (
	lockPOFields = xsql.GetFieldSlice(&LockPO{})
)

type LockPO struct {
	Id      int64  `db:"id"       json:"id"`
	LockKey string `db:"lock_key" json:"lock_key"`
	LockVal string `db:"lock_val" json:"lock_val"`
	HeldBy  string `db:"held_by"  json:"held_by"`
	Expire  int64  `db:"expire"   json:"expire"`
	Ctime   int64  `db:"ctime"    json:"ctime"`
}

func (LockPO) TableName() string {
	return lockPOTableName
}

func (s *LockPO) Values() []any {
	return []any{
		s.Id,
		s.LockKey,
		s.LockVal,
		s.HeldBy,
		s.Expire,
		s.Ctime,
	}
}
