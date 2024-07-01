package userbase

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Model struct {
	Uid       uint64 `db:"uid" json:"uid"`
	Nickname  string `db:"nickname" json:"nickname"`
	Avatar    string `db:"avatar" json:"avatar"`
	StyleSign string `db:"style_sign" json:"style_sign"`
	Gender    int8   `db:"gender" json:"gender"`
	Tel       string `db:"tel" json:"tel"`
	Email     string `db:"email" json:"email"`
	Pass      string `db:"pass" json:"-"`
	Salt      string `db:"salt" json:"-"`
	CreateAt  int64  `db:"create_at" json:"create_at,omitempty"`
	UpdateAt  int64  `db:"update_at" json:"update_at,omitempty"`
}

type Repo struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) *Repo {
	return &Repo{
		db: db,
	}
}
