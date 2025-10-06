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
func (d *SystemMsgDao) Create(ctx context.Context, msg *MsgPO) error {
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
func (d *SystemMsgDao) BatchCreate(ctx context.Context, msgs []*MsgPO) error {
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
func (d *SystemMsgDao) GetById(ctx context.Context, msgId uuid.UUID) (*MsgPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_message WHERE id=?", systemMsgFields)
	var msg MsgPO
	err := d.db.QueryRowCtx(ctx, &msg, sql, msgId)
	return &msg, xsql.ConvertError(err)
}

// 批量获取消息
func (d *SystemMsgDao) BatchGetByIds(ctx context.Context, msgIds []uuid.UUID) ([]*MsgPO, error) {
	if len(msgIds) == 0 {
		return []*MsgPO{}, nil
	}

	var msgs []*MsgPO
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
func (d *SystemMsgDao) ListByChatId(ctx context.Context, chatId uuid.UUID, beforeTime int64, limit int32) ([]*MsgPO, error) {
	sql := fmt.Sprintf(
		"SELECT %s FROM system_message WHERE system_chat_id=? AND mtime<? ORDER BY mtime DESC LIMIT ?",
		systemMsgFields)
	var msgs []*MsgPO
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, chatId, beforeTime, limit)
	return msgs, xsql.ConvertError(err)
}

func (d *SystemMsgDao) ListUnreadMsgId(ctx context.Context, chatId uuid.UUID, recvUid int64) ([]uuid.UUID, error) {
	sql := "SELECT id FROM system_message WHERE system_chat_id=? AND recv_uid=? AND status=? ORDER BY mtime DESC"
	var msgIds []uuid.UUID
	err := d.db.QueryRowsCtx(ctx, &msgIds, sql, chatId, recvUid, model.SystemMsgStatusNormal)
	return msgIds, xsql.ConvertError(err)
}

func (d *SystemMsgDao) ListUnreadMsgIdWithLimit(ctx context.Context,
	chatId uuid.UUID, recvUid int64, offsetId uuid.UUID, cnt int32) ([]uuid.UUID, error) {
	sql := "SELECT id FROM system_message WHERE id>? AND system_chat_id=? AND recv_uid=? AND status=? " +
		" ORDER BY mtime DESC LIMIT ?"
	var msgIds []uuid.UUID
	err := d.db.QueryRowsCtx(ctx, &msgIds, sql,
		offsetId, chatId, recvUid, model.SystemMsgStatusNormal, cnt)
	return msgIds, xsql.ConvertError(err)
}

// 获取最新一条消息
func (d *SystemMsgDao) GetLastMsg(ctx context.Context, chatId uuid.UUID, recvUid int64) (*MsgPO, error) {
	sql := fmt.Sprintf(
		"SELECT %s FROM system_message WHERE system_chat_id=? AND recv_uid=? ORDER BY mtime DESC LIMIT 1",
		systemMsgFields)
	var msg MsgPO
	err := d.db.QueryRowCtx(ctx, &msg, sql, chatId, recvUid)
	return &msg, xsql.ConvertError(err)
}

// 更新消息状态
func (d *SystemMsgDao) UpdateStatus(ctx context.Context, msgId uuid.UUID, status model.SystemMsgStatus) error {
	sql := "UPDATE system_message SET status=?, mtime=? WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, status, time.Now().UnixMicro(), msgId)
	return xsql.ConvertError(err)
}

// 将消息状态从srcStatus更新为targetStatus
func (d *SystemMsgDao) UpdateStatusToTarget(ctx context.Context, chatId uuid.UUID, recvUid int64,
	targetStatus, srcStatus model.SystemMsgStatus) error {
	sql := "UPDATE system_message SET status=?, mtime=? WHERE system_chat_id=? AND recv_uid=? AND status=?"
	_, err := d.db.ExecCtx(ctx, sql,
		targetStatus,
		time.Now().UnixMicro(), chatId, recvUid, srcStatus)
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
func (d *SystemMsgDao) DeleteById(ctx context.Context, msgId uuid.UUID) error {
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

func (d *SystemMsgDao) DeleteByChatIdMsgIdRecvUid(ctx context.Context, chatId, msgId uuid.UUID, recvUid int64) error {
	sql := "DELETE FROM system_message WHERE id=? AND system_chat_id=? AND recv_uid=?"
	_, err := d.db.ExecCtx(ctx, sql, chatId, msgId, recvUid)
	return xsql.ConvertError(err)
}
