package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type MessageDao struct {
	db *xsql.DB
}

func NewMessageDao(db *xsql.DB) *MessageDao {
	return &MessageDao{
		db: db,
	}
}

func (d *MessageDao) DB() *xsql.DB {
	return d.db
}

func (d *MessageDao) Create(ctx context.Context, msg *Message) error {
	if msg.Utime == 0 {
		msg.Utime = time.Now().UnixNano()
	}
	sql := fmt.Sprintf("INSERT INTO p2p_message(%s) VALUES (%s)", insMsgFields, insMsgQst)
	_, err := d.db.ExecCtx(ctx, sql,
		msg.MsgId,
		msg.SenderId,
		msg.ChatId,
		msg.MsgType,
		msg.Content,
		msg.Status,
		msg.Seq,
		msg.Utime)
	return xsql.ConvertError(err)
}
