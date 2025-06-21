package p2p

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xslice"
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

var (
	_chatInst  = &ChatPO{}
	chatFields = xsql.GetFields(_chatInst)

	initChatFields, initChatQuest = xsql.SelectFields2(_chatInst, "chat_id", "ctime", "user_id", "peer_id")
	insChatFields, insChatQst     = xsql.GetFields2(_chatInst, "id") // for insert

	sqlInitChat    = fmt.Sprintf("INSERT INTO p2p_chat(%s) VALUES (%s), (%s)", initChatFields, initChatQuest, initChatQuest)
	sqlCreateChat  = fmt.Sprintf("INSERT INTO p2p_chat(%s) VALUES (%s)", insChatFields, insChatQst)
	sqlGetByChatId = fmt.Sprintf("SELECT %s FROM p2p_chat WHERE chat_id=?", chatFields)
	sqlGetByUsers  = fmt.Sprintf(
		"SELECT %s FROM p2p_chat WHERE (user_id=? AND peer_id=?) OR (user_id=? AND peer_id=?) LIMIT 1", chatFields)
	sqlGetByChatIdUserId       = fmt.Sprintf("SELECT %s FROM p2p_chat WHERE chat_id=? AND user_id=?", chatFields)
	sqlBatchGetByChatIdsUserId = fmt.Sprintf("SELECT %s FROM p2p_chat WHERE user_id=? AND chat_id IN (%%s)", chatFields)
)

const (
	sqlCountByUserId    = "SELECT COUNT(*) AS cnt FROM p2p_chat WHERE user_id=?"
	sqlResetUnreadCount = "UPDATE p2p_chat SET unread_count=0, last_read_message_id=last_message_id, last_read_time=? " +
		"WHERE chat_id=? AND user_id=?"
)

// 初始化用户A和用户B的会话
//
// 数据表中新建两条记录
func (d *ChatDao) InitChat(ctx context.Context, chatId, userA, userB int64) error {
	now := time.Now().UnixMicro()
	ins1 := ChatPO{
		ChatId: chatId,
		Ctime:  now,
		UserId: userA,
		PeerId: userB,
	}
	ins2 := ChatPO{
		ChatId: chatId,
		Ctime:  now,
		UserId: userB,
		PeerId: userA,
	}
	_, err := d.db.ExecCtx(ctx, sqlInitChat,
		ins1.ChatId, ins1.Ctime, ins1.UserId, ins1.PeerId,
		ins2.ChatId, ins2.Ctime, ins2.UserId, ins2.PeerId,
	)

	return xsql.ConvertError(err)
}

// 插入一条记录
func (d *ChatDao) Create(ctx context.Context, chat *ChatPO) (int64, error) {
	if chat.Ctime == 0 {
		chat.Ctime = time.Now().UnixMicro()
	}

	res, err := d.db.ExecCtx(ctx, sqlCreateChat,
		chat.ChatId,
		chat.UserId,
		chat.PeerId,
		chat.UnreadCount,
		chat.Ctime,
		chat.LastMessageId,
		chat.LastMessageSeq,
		chat.LastReadMessageId,
		chat.LastReadTime,
	)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	newId, _ := res.LastInsertId()
	return newId, nil
}

// 按照会话id获取会话记录
func (d *ChatDao) GetByChatId(ctx context.Context, chatId int64) ([]*ChatPO, error) {
	var chat []*ChatPO
	err := d.db.QueryRowsCtx(ctx, &chat, sqlGetByChatId, chatId)
	return chat, xsql.ConvertError(err)
}

// 按照两个用户获取会话记录
func (d *ChatDao) GetByUsers(ctx context.Context, userA, userB int64) (*ChatPO, error) {
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sqlGetByUsers, userA, userB, userB, userA)
	return &chat, xsql.ConvertError(err)
}

// 获取用户的会话数量
func (d *ChatDao) CountByUserId(ctx context.Context, userId int64) (int64, error) {
	var cnt int64
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountByUserId, userId)
	return cnt, xsql.ConvertError(err)
}

func (d *ChatDao) GetByChatIdUserId(ctx context.Context, chatId, userId int64) (*ChatPO, error) {
	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sqlGetByChatIdUserId, chatId, userId)
	return &chat, xsql.ConvertError(err)
}

func (d *ChatDao) BatchGetByChatIdsUserId(ctx context.Context, chatIds []int64, userId int64) ([]*ChatPO, error) {
	var chats []*ChatPO
	sql := fmt.Sprintf(sqlBatchGetByChatIdsUserId, xslice.JoinInts(chatIds))
	err := d.db.QueryRowsCtx(ctx, &chats, sql, userId)

	return chats, xsql.ConvertError(err)
}

// 分页获取
func (d *ChatDao) PageListByUserId(ctx context.Context, userId, lastSeq int64, count int, unread bool) ([]*ChatPO, error) {
	var chats []*ChatPO
	sql := "SELECT %s FROM p2p_chat WHERE user_id=? AND last_message_seq<? "
	if unread {
		sql += "AND unread_count>0 "
	}
	sql += "ORDER BY last_message_seq DESC LIMIT ?"
	sql = fmt.Sprintf(sql, chatFields)
	err := d.db.QueryRowsCtx(ctx, &chats, sql, userId, lastSeq, count)

	return chats, xsql.ConvertError(err)
}

func (d *ChatDao) UpdateLastMsg(ctx context.Context,
	lastMsgId, lastMsgSeq int64,
	chatId, userId int64,
	incrUnread bool,
) error {

	sql := "UPDATE p2p_chat SET last_message_id=?, last_message_seq=?"
	if incrUnread {
		sql += ", unread_count=unread_count+1"
	}
	sql += " WHERE chat_id=? AND user_id=?"
	_, err := d.db.ExecCtx(ctx, sql, lastMsgId, lastMsgSeq, chatId, userId)

	return xsql.ConvertError(err)
}

// 更新最新消息和最新已读消息
func (d *ChatDao) UpdateMsg(ctx context.Context,
	lastMsgId, lastMsgSeq, lastReadMsgId int64,
	chatId, userId int64,
	incrUnread bool,
) error {

	sql := "UPDATE p2p_chat SET last_message_id=?, last_message_seq=?, " +
		"last_read_message_id=?, last_read_time=?"
	if incrUnread {
		sql += ", unread_count=unread_count+1"
	}
	sql += " WHERE chat_id=? AND user_id=?"
	_, err := d.db.ExecCtx(ctx,
		sql,
		lastMsgId, lastMsgSeq, lastReadMsgId, time.Now().UnixMicro(),
		chatId, userId,
	)
	return xsql.ConvertError(err)
}

// 清除未读数
func (d *ChatDao) ResetUnreadCount(ctx context.Context, chatId, userId int64) error {
	_, err := d.db.ExecCtx(ctx, sqlResetUnreadCount, time.Now().UnixMicro(), chatId, userId)
	return xsql.ConvertError(err)
}
