package dao

const (
	AlreadyPinned = 1
	NotPinned     = 0
)

// comment表
type Comment struct {
	Id       int64  `json:"id" db:"id"`
	Oid      int64  `json:"oid" db:"oid"`
	CType    int8   `json:"ctype" db:"ctype"`
	Content  string `json:"content" db:"content"`
	Uid      int64  `json:"uid" db:"uid"`
	RootId   int64  `json:"root" db:"root"`
	ParentId int64  `json:"parent" db:"parent"`
	ReplyUid int64  `json:"ruid" db:"ruid"`
	State    int8   `json:"state" db:"state"`
	Like     int    `json:"like" db:"like"`
	Dislike  int    `json:"dislike" db:"dislike"`
	Report   int    `json:"repot" db:"report"`
	IsPin    int8   `json:"pin" db:"pin"`
	Ip       int64  `json:"ip" db:"ip"`
	Ctime    int64  `json:"ctime" db:"ctime"`
	Mtime    int64  `json:"mtime" db:"mtime"`
}

type RootParent struct {
	Id       int64 `json:"id" db:"id"`
	RootId   int64 `json:"root" db:"root"`
	ParentId int64 `json:"parent" db:"parent"`
	Oid      int64 `json:"oid" db:"oid"`
	IsPin    int8  `json:"is_pin" db:"pin"`
}

type UidOid struct {
	Uid int64
	Oid int64
}

type RootCnt struct {
	Root int64 `db:"root"`
	Cnt  int64 `db:"cnt"`
}
