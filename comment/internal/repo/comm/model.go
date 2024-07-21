package comm

import "github.com/zeromicro/go-zero/core/stores/sqlx"

type Repo struct {
	db sqlx.SqlConn
}

func New(db sqlx.SqlConn) *Repo {
	return &Repo{
		db: db,
	}
}

// comment表
type Model struct {
	Id       uint64 `json:"id" db:"id"`
	Oid      uint64 `json:"oid" db:"oid"`
	CType    int8   `json:"ctype" db:"ctype"`
	Content  string `json:"content" db:"content"`
	Uid      uint64 `json:"uid" db:"uid"`
	RootId   uint64 `json:"root" db:"root"`
	ParentId uint64 `json:"parent" db:"parent"`
	ReplyUid uint64 `json:"ruid" db:"ruid"`
	State    int8   `json:"state" db:"state"`
	Like     int    `json:"like" db:"like"`
	Dislike  int    `json:"dislike" db:"dislike"`
	Report   int    `json:"-" db:"report"`
	IsPin    int8   `json:"pin" db:"pin"`
	Ip       int64  `json:"-" db:"ip"`
	Ctime    int64  `json:"ctime" db:"ctime"`
	Mtime    int64  `json:"-" db:"mtime"`
}
