package usersetting

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	userSettingPOTableName = "user_setting"
)

var (
	userSettingPOFields = xsql.GetFieldSlice(&UserSettingPO{})
)

type UserSettingPO struct {
	Uid   int64           `db:"uid"`
	Flags int64           `db:"flags"`
	Ext   json.RawMessage `db:"ext"`
	Ctime int64           `db:"ctime"`
	Utime int64           `db:"utime"`
}

func (UserSettingPO) TableName() string {
	return userSettingPOTableName
}

func (s *UserSettingPO) Values() []any {
	ext := s.Ext
	if s.Ext == nil {
		ext = json.RawMessage{}
	}
	return []any{
		s.Uid,
		s.Flags,
		ext,
		s.Ctime,
		s.Utime,
	}
}
