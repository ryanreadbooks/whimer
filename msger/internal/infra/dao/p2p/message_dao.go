package p2p

import (
	"context"
	"fmt"
	"time"

	slices "github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
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

func (d *MessageDao) Create(ctx context.Context, msg *MessagePO) error {
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

func (d *MessageDao) GetByChatIdMsgId(ctx context.Context, chatId, msgId int64) (*MessagePO, error) {
	sql := fmt.Sprintf("SELECT %s FROM p2p_message WHERE chat_id=? AND msg_id=?", msgFields)
	var msg MessagePO
	err := d.db.QueryRowCtx(ctx, &msg, sql, chatId, msgId)
	return &msg, xsql.ConvertError(err)
}

func (d *MessageDao) GetByMsgIds(ctx context.Context, chatId int64, msgIds []int64) ([]*MessagePO, error) {
	if len(msgIds) == 0 {
		return []*MessagePO{}, nil
	}

	var msgs []*MessagePO
	params := slices.JoinInts(msgIds)
	sql := fmt.Sprintf(
		"SELECT %s FROM p2p_message WHERE chat_id=? AND msg_id IN (%s) ORDER BY seq DESC",
		msgFields, params)
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, chatId)
	return msgs, xsql.ConvertError(err)
}

// 按照msgId获取消息
func (d *MessageDao) GetByMsgId(ctx context.Context, msgId int64) (*MessagePO, error) {
	sql := fmt.Sprintf("SELECT %s FROM p2p_message WHERE msg_id=?", msgFields)
	var po MessagePO
	err := d.db.QueryRowCtx(ctx, &po, sql, msgId)
	if err != nil {
		return nil, err
	}

	return &po, nil
}

// 更新消息状态
func (d *MessageDao) RevokeMsg(ctx context.Context, chatId, msgId int64) error {
	sql := "UPDATE p2p_message SET status=?, utime=? WHERE chat_id=? AND msg_id=?"
	_, err := d.db.ExecCtx(ctx, sql,
		gm.MsgStatusRevoked, time.Now().UnixNano(), chatId, msgId)
	return xsql.ConvertError(err)
}
