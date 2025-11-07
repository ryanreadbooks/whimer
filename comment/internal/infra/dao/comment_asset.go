package dao

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/comment/internal/model"
)

type CommentAsset struct {
	Id        int64                  `db:"id"         redis:"id"         msgpack:"id"`
	CommentId int64                  `db:"comment_id" redis:"comment_id" msgpack:"comment_id"`
	Type      model.CommentAssetType `db:"type"       redis:"type"       msgpack:"type"` // 见 [internal/model/comment.go]
	StoreKey  string                 `db:"store_key"  redis:"store_key"  msgpack:"store_key"`
	Metadata  json.RawMessage        `db:"metadata"   redis:"metadata"   msgpack:"metadata"`
	Ctime     int64                  `db:"ctime"      redis:"ctime"      msgpack:"ctime"`
}
