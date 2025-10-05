package system

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type SystemChatDao struct {
	db *xsql.DB
}

func NewSystemChatDao(db *xsql.DB) *SystemChatDao {
	return &SystemChatDao{
		db: db,
	}
}

func (d *SystemChatDao) DB() *xsql.DB {
	return d.db
}

// 创建系统会话
func (d *SystemChatDao) Create(ctx context.Context, chat *SystemChatPO) error {
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
		chat.LastReadTime,
		chat.UnreadCount)
	return xsql.ConvertError(err)
}

// 获取用户的系统会话
func (d *SystemChatDao) GetByUidAndType(ctx context.Context,
	uid int64, chatType model.SystemChatType) (*SystemChatPO, error) {

	sql := fmt.Sprintf("SELECT %s FROM system_chat WHERE uid=? AND type=? LIMIT 1", systemChatFields)
	var chat SystemChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, uid, chatType)
	return &chat, xsql.ConvertError(err)
}

// 获取用户的所有系统会话
func (d *SystemChatDao) ListByUid(ctx context.Context, uid int64) ([]*SystemChatPO, error) {
	sql := fmt.Sprintf("SELECT %s FROM system_chat WHERE uid=? ORDER BY `type` DESC", systemChatFields)
	var chats []*SystemChatPO
	err := d.db.QueryRowsCtx(ctx, &chats, sql, uid)
	return chats, xsql.ConvertError(err)
}

// 更新系统会话的最后消息
func (d *SystemChatDao) UpdateLastMsg(ctx context.Context, chatId, lastMsgId uuid.UUID, incrUnread bool) error {
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
func (d *SystemChatDao) UpdateUnreadCount(ctx context.Context, chatId uuid.UUID, unreadCount int64) error {
	sql := "UPDATE system_chat SET unread_count=?, mtime=? WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, unreadCount, time.Now().UnixMicro(), chatId)
	return xsql.ConvertError(err)
}

// 清空未读消息数
func (d *SystemChatDao) ClearUnread(ctx context.Context, chatId uuid.UUID, lastReadMsgId uuid.UUID) error {
	sql := "UPDATE system_chat SET unread_count=0, last_read_msg_id=?, last_read_time=?, mtime=? WHERE id=?"
	now := time.Now().UnixMicro()
	_, err := d.db.ExecCtx(ctx, sql, lastReadMsgId, now, now, chatId)
	return xsql.ConvertError(err)
}

// 删除系统会话
func (d *SystemChatDao) Delete(ctx context.Context, chatId uuid.UUID) error {
	sql := "DELETE FROM system_chat WHERE id=?"
	_, err := d.db.ExecCtx(ctx, sql, chatId)
	return xsql.ConvertError(err)
}
