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
	Id         uint64 `json:"id" db:"id"`
	NoteId     uint64 `json:"note_id" db:"note_id"`
	CType      int8   `json:"ctype" db:"ctype"`
	Content    string `json:"content" db:"content"`
	Uid        uint64 `json:"uid" db:"uid"`
	ParentId   uint64 `json:"parent_id" db:"parent_id"`
	ReplyId    uint64 `json:"reply_id" db:"reply_id"`
	ReplyUid   uint64 `json:"reply_uid" db:"reply_uid"`
	State      int8   `json:"state" db:"state"`
	Like       int    `json:"like" db:"like_count"`
	Dislike    int    `json:"dislike" db:"dislike_count"`
	Report     int    `json:"-" db:"report_count"`
	IsTop      int8   `json:"is_top" db:"is_top"`
	ClientInfo string `json:"-" db:"client_info"`
	Ctime      int64  `json:"ctime" db:"ctime"`
}
