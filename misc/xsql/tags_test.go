package xsql

import (
	"testing"
)

type TagTs struct {
	Id                int64 `db:"id"`
	ChatId            int64 `db:"chat_id"`
	UserId            int64 `db:"user_id"`
	PeerId            int64 `db:"peer_id"`
	UnreadCount       int64 `db:"unread_count"`
	Ctime             int64 `db:"ctime"`
	LastMessageId     int64 `db:"last_message_id"`
	LastMessageTime   int64 `db:"last_message_time"`
	LastReadMessageId int64 `db:"last_read_message_id"`
	LastReadTime      int64 `db:"last_read_time"`
	NoTag             int64
}

type NoTags struct {
	A int64
	B int64
	C string
}

type PTags struct {
	Id   *int64 `db:"id"`
	Name string `db:"name"`
}

type PTags2 struct {
	Name string `db:"name"`
}

func TestGetDbTags(t *testing.T) {
	s := GetFields(TagTs{})
	t.Log(s)
	t.Log(GetFields(TagTs{}, "id", "ctime"))

	s1 := GetFields(&NoTags{})
	t.Log(s1)

	t.Log(GetFields(PTags{}))
}

func TestGetFields2(t *testing.T) {
	t.Log(GetFields2(TagTs{}))
	t.Log(GetFields2(TagTs{}, "id", "ctime"))
	t.Log(GetFields2(&NoTags{}))
	t.Log(GetFields2(&PTags2{}))
	t.Log(GetFields2(PTags{}))
}
