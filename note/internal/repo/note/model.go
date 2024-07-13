package note

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
	Id       uint64 `db:"id"`
	Title    string `db:"title"`     // 标题
	Desc     string `db:"desc"`      // 描述
	Privacy  int8   `db:"privacy"`   // 公开类型
	Owner    uint64 `db:"owner"`     // 笔记作者
	CreateAt int64  `db:"create_at"` // 创建时间
	UpdateAt int64  `db:"update_at"` // 更新时间
}
