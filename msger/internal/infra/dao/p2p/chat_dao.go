package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
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

// 初始化用户A和用户B的会话
//
// 数据表中新建两条记录
func (d *ChatDao) InitChat(ctx context.Context, chatId, userA, userB int64) error {
	now := time.Now().UnixNano()
	ins1 := Chat{
		ChatId: chatId,
		Ctime:  now,
		UserId: userA,
		PeerId: userB,
	}
	ins2 := Chat{
		ChatId: chatId,
		Ctime:  now,
		UserId: userB,
		PeerId: userA,
	}

	fields, quest := xsql.SelectFields2(_chatInst, "chat_id", "ctime", "user_id", "peer_id")
	sql := fmt.Sprintf("INSERT INTO p2p_chat(%s) VALUES (%s), (%s)", fields, quest, quest)
	_, err := d.db.ExecCtx(ctx, sql,
		ins1.ChatId, ins1.Ctime, ins1.UserId, ins1.PeerId,
		ins2.ChatId, ins2.Ctime, ins2.UserId, ins2.PeerId,
	)

	return xsql.ConvertError(err)
}

// 插入一条记录
func (d *ChatDao) Create(ctx context.Context, chat *Chat) (int64, error) {
	if chat.Ctime == 0 {
		chat.Ctime = time.Now().UnixNano()
	}

	sql := fmt.Sprintf("INSERT INTO p2p_chat(%s) VALUES (%s)", insChatFields, insChatQst)
	res, err := d.db.ExecCtx(ctx, sql,
		chat.ChatId,
		chat.UserId,
		chat.PeerId,
		chat.UnreadCount,
		chat.Ctime,
		chat.LastMessageId,
		chat.LastMessageSeq,
		chat.LastReadTime,
	)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	newId, _ := res.LastInsertId()
	return newId, nil
}

// 按照会话id获取会话记录
func (d *ChatDao) GetByChatId(ctx context.Context, chatId int64) ([]*Chat, error) {
	var chat []*Chat
	sql := fmt.Sprintf("SELECT %s FROM p2p_chat WHERE chat_id=?", chatFields)
	err := d.db.QueryRowsCtx(ctx, &chat, sql, chatId)
	return chat, xsql.ConvertError(err)
}

// 按照两个用户获取会话记录
func (d *ChatDao) GetByUsers(ctx context.Context, userA, userB int64) (*Chat, error) {
	var chat Chat
	sql := fmt.Sprintf(
		"SELECT %s FROM p2p_chat WHERE (user_id=? AND peer_id=?) OR (user_id=? AND peer_id=?) LIMIT 1", chatFields)
	err := d.db.QueryRowsCtx(ctx, &chat, sql, userA, userB, userB, userA)
	return &chat, xsql.ConvertError(err)
}

// 获取用户的会话数量
func (d *ChatDao) CountByUserId(ctx context.Context, userId int64) (int64, error) {
	var cnt int64
	sql := fmt.Sprintf("SELECT COUNT(*) AS cnt FROM p2p_chat WHERE user_id=?")
	err := d.db.QueryRowCtx(ctx, &cnt, sql, userId)
	return cnt, xsql.ConvertError(err)
}

func (d *ChatDao) GetByChatIdUserId(ctx context.Context, chatId, userId int64) (*Chat, error) {
	var chat Chat
	sql := fmt.Sprintf("SELECT %s FROM p2p_chat WHERE chat_id=? AND user_id=?", chatFields)
	err := d.db.QueryRowCtx(ctx, &chat, sql, chatId, userId)
	return &chat, xsql.ConvertError(err)
}

// 分页获取
func (d *ChatDao) PageGetByUserId(ctx context.Context, userId int64, lastSeq int64, count int) ([]*Chat, error) {
	var chats []*Chat
	sql := fmt.Sprintf(
		"SELECT %s FROM p2p_chat WHERE user_id=? AND last_message_time<? ORDER BY last_message_seq DESC LIMIT ?", chatFields)
	err := d.db.QueryRowsCtx(ctx, &chats, sql, userId, lastSeq, count)
	return chats, xsql.ConvertError(err)
}

func (d *ChatDao) UpdateLastMsg(ctx context.Context,
	lastMsgId, lastMsgSeq int64,
	chatId, userId int64,
	addUnread bool,
) error {

	sql := "UPDATE p2p_chat SET last_message_id=?, last_message_seq=?"
	if addUnread {
		sql += ", unread_count=unread_count+1"
	}
	sql += " WHERE chat_id=? AND user_id=?"
	_, err := d.db.ExecCtx(ctx, sql, lastMsgId, lastMsgSeq, chatId, userId)

	return xsql.ConvertError(err)
}

// 清除未读数
func (d *ChatDao) ResetUnreadCount(ctx context.Context, chatId, userId int64) error {
	sql := fmt.Sprintf("UPDATE p2p_chat SET unread_count=0, last_read_time=? WHERE chat_id=? AND user_id=?")
	_, err := d.db.ExecCtx(ctx, sql, time.Now().UnixNano(), chatId, userId)
	return xsql.ConvertError(err)
}
