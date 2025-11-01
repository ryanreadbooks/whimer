package chat

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
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

func (d *ChatDao) Create(ctx context.Context, chat *ChatPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertIgnoreInto(chatPOTableName)
	ib.Cols(chatPOFields...)
	ib.Values(chat.Values()...)

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatDao) GetById(ctx context.Context, id uuid.UUID) (*ChatPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(chatPOFields...)
	sb.From(chatPOTableName)
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()

	var chat ChatPO
	err := d.db.QueryRowCtx(ctx, &chat, sql, args...)
	return &chat, xsql.ConvertError(err)
}

func (d *ChatDao) DeleteById(ctx context.Context, id uuid.UUID) error {
	bd := sqlbuilder.NewDeleteBuilder()
	bd.DeleteFrom(chatPOTableName)
	bd.Where(bd.Equal("id", id))

	sql, args := bd.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ChatDao) UpdateName(ctx context.Context, id uuid.UUID, name string, mtime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatPOTableName)
	ub.Set(ub.Assign("name", name), ub.Assign("mtime", mtime))
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ChatDao) UpdateLastMsgId(ctx context.Context, id uuid.UUID, lastMsgId uuid.UUID, mtime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatPOTableName)
	ub.Set(ub.Assign("last_msg_id", lastMsgId), ub.Assign("mtime", mtime))
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ChatDao) UpdateSettings(ctx context.Context, id uuid.UUID, settings int64, mtime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(chatPOTableName)
	ub.Set(ub.Assign("settings", settings), ub.Assign("mtime", mtime))
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}
