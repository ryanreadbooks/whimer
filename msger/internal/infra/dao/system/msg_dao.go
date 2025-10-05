package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type SystemMsgDao struct {
	db *xsql.DB
}

func NewSystemMsgDao(db *xsql.DB) *SystemMsgDao {
	return &SystemMsgDao{
		db: db,
	}
}

func (d *SystemMsgDao) DB() *xsql.DB {
	return d.db
}

// 创建系统消息
func (d *SystemMsgDao) Create(ctx context.Context, msg *SystemMsgPO) error {
	if msg.Mtime == 0 {
		msg.Mtime = time.Now().UnixMicro()
	}
	sql := fmt.Sprintf("INSERT INTO system_message(%s) VALUES (%s)", insSystemMsgFields, insSystemMsgQst)
	_, err := d.db.ExecCtx(ctx, sql,
		msg.Id,
		msg.SystemChatId,
		msg.Uid,
		msg.RecvUid,
		msg.Status,
		msg.MsgType,
		msg.Content,
		msg.Mtime)
	return xsql.ConvertError(err)
}

// 批量创建系统消息
func (d *SystemMsgDao) BatchCreate(ctx context.Context, msgs []*SystemMsgPO) error {
	if len(msgs) == 0 {
		return nil
	}

	now := time.Now().UnixMicro()

	return xslice.BatchExec(msgs, 100, func(start, end int) error {
		datas := msgs[start:end]
		qm := "(" + insSystemMsgQst + ")"
		qsts := xslice.Repeat(qm, len(datas))
		// 批量插入
		sql := fmt.Sprintf("INSERT INTO system_message(%s) VALUES %s",
			insSystemMsgFields, strings.Join(qsts, ","))
		args := make([]any, 0, len(datas)*8)
		for _, data := range datas {
			if data.Mtime == 0 {
				data.Mtime = now
			}

			args = append(args,
				data.Id,
				data.SystemChatId,
				data.Uid,
				data.RecvUid,
				data.Status,
				data.MsgType,
				data.Content,
				data.Mtime)
		}
		_, err := d.db.ExecCtx(ctx, sql, args...)
		return xsql.ConvertError(err)
	})
}

// 根据消息ID获取消息
func (d *SystemMsgDao) GetById(ctx context.Context, msgId uuid.UUID) (*SystemMsgPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_message WHERE id=?", systemMsgFields)
	var msg SystemMsgPO
	err := d.db.QueryRowCtx(ctx, &msg, sql, msgId)
	return &msg, xsql.ConvertError(err)
}

// 批量获取消息
func (d *SystemMsgDao) BatchGetByIds(ctx context.Context, msgIds []uuid.UUID) ([]*SystemMsgPO, error) {
	if len(msgIds) == 0 {
		return []*SystemMsgPO{}, nil
	}

	var msgs []*SystemMsgPO
	params := make([]any, len(msgIds))
	for i, id := range msgIds {
		params[i] = id
	}
	sql := fmt.Sprintf(
		"SELECT %s FROM system_message WHERE id IN (%s) ORDER BY mtime DESC",
		systemMsgFields, xslice.JoinStrings(xslice.Repeat("?", len(msgIds))))
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, params...)
	return msgs, xsql.ConvertError(err)
}

// 获取系统会话中的消息列表
func (d *SystemMsgDao) ListByChatId(ctx context.Context, chatId uuid.UUID, beforeTime int64, limit int32) ([]*SystemMsgPO, error) {
	sql := fmt.Sprintf(
		"SELECT %s FROM system_message WHERE system_chat_id=? AND mtime<? ORDER BY mtime DESC LIMIT ?",
		systemMsgFields)
	var msgs []*SystemMsgPO
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, chatId, beforeTime, limit)
	return msgs, xsql.ConvertError(err)
}

// 更新消息状态
func (d *SystemMsgDao) UpdateStatus(ctx context.Context, msgId uuid.UUID, status model.SystemMsgStatus) error {
	sql := "UPDATE system_message SET status=?, mtime=? WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, status, time.Now().UnixMicro(), msgId)
	return xsql.ConvertError(err)
}

// 批量更新消息状态
func (d *SystemMsgDao) BatchUpdateStatus(ctx context.Context, msgIds []uuid.UUID, status model.SystemMsgStatus) error {
	if len(msgIds) == 0 {
		return nil
	}

	params := make([]any, 0, len(msgIds)+1)
	params = append(params, status)
	for _, id := range msgIds {
		params = append(params, id)
	}
	sql := fmt.Sprintf(
		"UPDATE system_message SET status=? WHERE id IN (%s)",
		xslice.JoinStrings(xslice.Repeat("?", len(msgIds))))
	_, err := d.db.ExecCtx(ctx, sql, params...)
	return xsql.ConvertError(err)
}

// 删除消息
func (d *SystemMsgDao) Delete(ctx context.Context, msgId uuid.UUID) error {
	sql := "DELETE FROM system_message WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, msgId)
	return xsql.ConvertError(err)
}

// 删除系统会话中的所有消息
func (d *SystemMsgDao) DeleteByChatId(ctx context.Context, chatId uuid.UUID) error {
	sql := "DELETE FROM system_message WHERE system_chat_id=?"
	_, err := d.db.ExecCtx(ctx, sql, chatId)
	return xsql.ConvertError(err)
}
