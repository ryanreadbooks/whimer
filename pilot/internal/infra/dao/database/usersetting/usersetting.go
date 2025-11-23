package usersetting

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	userSettingPOTableName = "user_setting"
)

var (
	userSettingPOFields = xsql.GetFieldSlice(&UserSettingPO{})
)

type UserSettingPO struct {
	Uid   int64  `db:"uid"   json:"uid"`
	Flags int64  `db:"flags" json:"flags"`
	Ext   []byte `db:"ext"   json:"ext"`
	Ctime int64  `db:"ctime" json:"ctime"`
	Utime int64  `db:"utime" json:"utime"`
}

func (UserSettingPO) TableName() string {
	return userSettingPOTableName
}

func (s *UserSettingPO) Values() []any {
	ext := s.Ext
	if s.Ext == nil {
		ext = []byte{}
	}
	return []any{
		s.Uid,
		s.Flags,
		ext,
		s.Ctime,
		s.Utime,
	}
}
