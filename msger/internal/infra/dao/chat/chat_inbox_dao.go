package chat

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatInboxDao struct {
	db *xsql.DB
}

func NewChatInboxDao(db *xsql.DB) *ChatInboxDao {
	return &ChatInboxDao{
		db: db,
	}
}

func (d *ChatInboxDao) Create(ctx context.Context, b *ChatInboxPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(chatInboxPOTableName)
	ib.Cols(chatInboxPOFields...)
	ib.Values(b.Values()...)
	ib.SQL("ON DUPLICATE KEY UPDATE mtime=mtime") // duplicate key时不处理

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatInboxDao) BatchCreate(ctx context.Context, bs []*ChatInboxPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(chatInboxPOTableName)
	ib.Cols(chatInboxPOFields...)
	for _, b := range bs {
		ib.Values(b.Values()...)
	}
	ib.SQL("ON DUPLICATE KEY UPDATE mtime=mtime") // duplicate key时不处理

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatInboxDao) GetByUidChatId(ctx context.Context, uid int64, chatId uuid.UUID) (*ChatInboxPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(chatInboxPOFields...)
	sb.From(chatInboxPOTableName)
	sb.Where(sb.Equal("uid", uid), sb.Equal("chat_id", chatId))

	sql, args := sb.Build()
	var result ChatInboxPO
	err := d.db.QueryRowCtx(ctx, &result, sql, args...)
	return &result, xsql.ConvertError(err)
}

func (d *ChatInboxDao) UpdataStatus(ctx context.Context,
	uid int64, chatId uuid.UUID,
	status model.ChatInboxStatus,
	mtime int64) error {

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatInboxPOTableName)
	ub.Set(ub.EQ("status", status), ub.EQ("mtime", mtime))
	ub.Where(ub.EQ("uid", uid), ub.EQ("chat_id", chatId))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatInboxDao) UpdateLastMsgId(ctx context.Context,
	uid int64, chatId uuid.UUID,
	lastMsgId uuid.UUID, mtime int64) error {

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatInboxPOTableName)
	ub.Set(ub.EQ("last_msg_id", lastMsgId), ub.EQ("mtime", mtime))
	ub.Where(ub.EQ("uid", uid), ub.EQ("chat_id", chatId))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatInboxDao) SetPinned(ctx context.Context,
	uid int64, chatId uuid.UUID, pinned bool, mtime int64) error {

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatInboxPOTableName)
	ub.Set(ub.EQ("pinned", pinned), ub.EQ("mtime", mtime))
	ub.Where(ub.EQ("uid", uid), ub.EQ("chat_id", chatId))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// 批量更新last_msg_id并且unread_count++
func (d *ChatInboxDao) BatchUpdateLastMsgId(ctx context.Context,
	chatId uuid.UUID, uids []int64, lastMsgId uuid.UUID, mtime int64) error {

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatInboxPOTableName).
		Set(ub.EQ("last_msg_id", lastMsgId), ub.EQ("mtime", mtime), ub.Incr("unread_count")).
		Where(ub.In("uid", uids), ub.EQ("chat_id", chatId))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)

	return xsql.ConvertError(err)
}
