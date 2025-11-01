package system

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatDao struct {
	db *xsql.DB
}

func NewChatDao(db *xsql.DB) *ChatDao {
	return &ChatDao{
		db: db,
	}
}

func (d *ChatDao) DB() *xsql.DB {
	return d.db
}

// 创建系统会话
func (d *ChatDao) Create(ctx context.Context, chat *ChatPO) error {
	if chat.Mtime == 0 {
		chat.Mtime = time.Now().UnixMicro()
	}
	sql := fmt.Sprintf("INSERT INTO system_chat(%s) VALUES (%s) ON DUPLICATE KEY UPDATE mtime=VALUES(mtime)",
		insSystemChatFields, insSystemChatQst)
	_, err := d.db.ExecCtx(ctx, sql,
		chat.Id,
		chat.Type,
		chat.Uid,
		chat.Mtime,
		chat.LastMsgId,
		chat.LastReadMsgId,
		chat.UnreadCount)
	return xsql.ConvertError(err)
}

func (d *ChatDao) UpdateMsgs(ctx context.Context, chatId,
	lastMsgId, lastReadMsgId uuid.UUID, unread int64) error {

	now := time.Now().UnixNano()
	sql := "UPDATE system_chat SET last_msg_id=?, last_read_msg_id=?, unread_count=?, mtime=? WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, lastMsgId, lastReadMsgId, unread, now, chatId)
	return xsql.ConvertError(err)
}

// 获取系统会话
func (d *ChatDao) GetById(ctx context.Context, id uuid.UUID) (*ChatPO, error) {
	var sql = fmt.Sprintf("SELECT %s FROM system_chat WHERE id=?", systemChatFields)
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, id)
	return &chat, xsql.ConvertError(err)
}

func (d *ChatDao) GetByIdForUpdate(ctx context.Context, id uuid.UUID) (*ChatPO, error) {
	var sql = fmt.Sprintf("SELECT %s FROM system_chat WHERE id=? FOR UPDATE", systemChatFields)
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, id)
	return &chat, xsql.ConvertError(err)
}

// 批量获取
func (d *ChatDao) BatchGetById(ctx context.Context, chatIds []uuid.UUID) ([]*ChatPO, error) {
	if len(chatIds) == 0 {
		return nil, nil
	}

	var sql = fmt.Sprintf("SELECT %s FROM system_chat WHERE id IN (%s)",
		systemChatFields, strings.Repeat("?,", len(chatIds)-1)+"?")
	var chats []*ChatPO
	err := d.db.QueryRowsCtx(ctx, &chats, sql, xslice.Any(chatIds)...)
	return chats, xsql.ConvertError(err)
}

// 获取用户的系统会话
func (d *ChatDao) GetByUidAndType(ctx context.Context, uid int64, chatType model.SystemChatType) (*ChatPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_chat WHERE uid=? AND type=? LIMIT 1", systemChatFields)
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, uid, chatType)
	return &chat, xsql.ConvertError(err)
}

func (d *ChatDao) GetByUidAndTypeForUpdate(ctx context.Context, uid int64, chatType model.SystemChatType) (*ChatPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_chat WHERE uid=? AND type=? LIMIT 1 FOR UPDATE", systemChatFields)
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, uid, chatType)
	return &chat, xsql.ConvertError(err)
}

// 获取用户的所有系统会话
func (d *ChatDao) ListByUid(ctx context.Context, uid int64) ([]*ChatPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_chat WHERE uid=? ORDER BY `type` DESC", systemChatFields)
	var chats []*ChatPO
	err := d.db.QueryRowsCtx(ctx, &chats, sql, uid)
	return chats, xsql.ConvertError(err)
}

// 更新系统会话的最后消息 并设置是否增加未读消息数
func (d *ChatDao) UpdateLastMsg(ctx context.Context, chatId, lastMsgId uuid.UUID, incrUnread bool) error {
	mtime := time.Now().UnixMicro()
	var sql string
	var args []any

	if incrUnread {
		sql = "UPDATE system_chat SET last_msg_id=?, unread_count=unread_count+1, mtime=? WHERE id=?"
		args = []any{lastMsgId, mtime, chatId}
	} else {
		sql = "UPDATE system_chat SET last_msg_id=?, mtime=? WHERE id=?"
		args = []any{lastMsgId, mtime, chatId}
	}

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

// 更新未读消息数
func (d *ChatDao) UpdateUnreadCount(ctx context.Context, chatId uuid.UUID, unreadCount int64) error {
	sql := "UPDATE system_chat SET unread_count=?, mtime=? WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, unreadCount, time.Now().UnixMicro(), chatId)
	return xsql.ConvertError(err)
}

// 清空未读消息数 并设置最后读取消息id
func (d *ChatDao) ClearUnread(ctx context.Context, chatId uuid.UUID, lastReadMsgId uuid.UUID) error {
	sql := "UPDATE system_chat SET unread_count=0, last_read_msg_id=?, mtime=? WHERE id=?"
	now := time.Now().UnixMicro()
	_, err := d.db.ExecCtx(ctx, sql, lastReadMsgId, now, chatId)
	return xsql.ConvertError(err)
}

// 删除系统会话
func (d *ChatDao) Delete(ctx context.Context, chatId uuid.UUID) error {
	sql := "DELETE FROM system_chat WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, chatId)
	return xsql.ConvertError(err)
}

func (d *ChatDao) GetUnreadCountById(ctx context.Context, chatId uuid.UUID) (int64, error) {
	const sql = "SELECT unread_count FROM system_chat WHERE id=?"
	var unreadCount int64
	err := d.db.QueryRowCtx(ctx, &unreadCount, sql, chatId)
	return unreadCount, xsql.ConvertError(err)
}
