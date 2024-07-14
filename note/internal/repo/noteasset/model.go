package noteasset

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
	Id        uint64 `db:"id"`
	AssetKey  string `db:"asset_key"`  // 资源key 不包含bucket name
	AssetType int8   `db:"asset_type"` // 资源类型
	NoteId    uint64 `db:"note_id"`    // 所属笔记id
	CreateAt  int64  `db:"create_at"`  // 创建时间
}
