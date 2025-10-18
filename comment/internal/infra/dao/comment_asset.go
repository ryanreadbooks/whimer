package dao

import (
	"encoding/json"
)

type CommentAsset struct {
	Id        int64           `db:"id"         redis:"id"`
	CommentId int64           `db:"comment_id" redis:"comment_id"`
	Type      int8            `db:"type"       redis:"type"` // 见 [internal/model/comment.go]
	StoreKey  string          `db:"store_key"  redis:"store_key"`
	Metadata  json.RawMessage `db:"metadata"   redis:"metadata"`
	Ctime     int64           `db:"ctime"      redis:"ctime"`
}
