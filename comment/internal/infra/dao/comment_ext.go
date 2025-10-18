package dao

import "encoding/json"

type CommentExt struct {
	CommentId int64           `db:"comment_id" json:"comment_id"`
	AtUsers   json.RawMessage `db:"at_users"   json:"at_users"`
	Ctime     int64           `db:"ctime"      json:"ctime"`
	Mtime     int64           `db:"mtime"      json:"mtime"`
}
