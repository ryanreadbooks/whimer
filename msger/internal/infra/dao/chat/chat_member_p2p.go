package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	chatMemberP2PPOTableName = "chat_member_p2p"
)

var (
	chatMemberP2PPoFields     = xsql.GetFieldSlice(&ChatMemberP2PPO{})
	chatMemberP2PPoFieldsNoId = xsql.GetFieldSlice(&ChatMemberP2PPO{}, "id")
)

// 单聊用户
type ChatMemberP2PPO struct {
	Id     int64     `db:"id"`
	ChatId uuid.UUID `db:"chat_id"`
	UidA   int64     `db:"uid_a"` // uid小
	UidB   int64     `db:"uid_b"` // uid大
	Ctime  int64     `db:"ctime"` // 创建时间
	Mtime  int64     `db:"mtime"` // 更新时间
}

func (p *ChatMemberP2PPO) Normalize() {
	if p.UidA > p.UidB {
		p.UidA, p.UidB = p.UidB, p.UidA
	}
}

func (p *ChatMemberP2PPO) ValuesNoId() []any {
	return []any{
		p.ChatId,
		p.UidA,
		p.UidB,
		p.Ctime,
		p.Mtime,
	}
}
