package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xslice"
	slices "github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type MsgDao struct {
	db *xsql.DB
}

func NewMsgDao(db *xsql.DB) *MsgDao {
	return &MsgDao{
		db: db,
	}
}

func (d *MsgDao) DB() *xsql.DB {
	return d.db
}

func (d *MsgDao) Create(ctx context.Context, msg *MsgPO) error {
	if msg.Utime == 0 {
		msg.Utime = time.Now().UnixMicro()
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

func (d *MsgDao) GetByChatIdMsgId(ctx context.Context, chatId, msgId int64) (*MsgPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM p2p_message WHERE chat_id=? AND msg_id=?", msgFields)
	var msg MsgPO
	err := d.db.QueryRowCtx(ctx, &msg, sql, chatId, msgId)
	return &msg, xsql.ConvertError(err)
}

func (d *MsgDao) GetByChatIdMsgIds(ctx context.Context, chatId int64, msgIds []int64) ([]*MsgPO, error) {
	if len(msgIds) == 0 {
		return []*MsgPO{}, nil
	}

	var msgs []*MsgPO
	params := slices.JoinInts(msgIds)
	sql := fmt.Sprintf(
		"SELECT %s FROM p2p_message WHERE chat_id=? AND msg_id IN (%s) ORDER BY seq DESC",
		msgFields, params)
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, chatId)
	return msgs, xsql.ConvertError(err)
}

func (d *MsgDao) BatchGetByChatIdMsgId(ctx context.Context, chatIds, msgIds []int64) ([]*MsgPO, error) {
	if len(msgIds) == 0 || len(chatIds) == 0 {
		return []*MsgPO{}, nil
	}

	var msgs []*MsgPO
	sql := fmt.Sprintf(
		"SELECT %s FROM p2p_message WHERE chat_id IN (%s) AND msg_id IN (%s) ORDER BY seq DESC",
		msgFields, xslice.JoinInts(chatIds), xslice.JoinInts(msgIds))
	err := d.db.QueryRowsCtx(ctx, &msgs, sql)
	return msgs, xsql.ConvertError(err)
}

// 更新消息状态
func (d *MsgDao) RevokeMsg(ctx context.Context, chatId, msgId int64) error {
	sql := "UPDATE p2p_message SET status=?, utime=? WHERE chat_id=? AND msg_id=?"
	_, err := d.db.ExecCtx(ctx, sql,
		gm.MsgStatusRevoked, time.Now().UnixMicro(), chatId, msgId)
	return xsql.ConvertError(err)
}
