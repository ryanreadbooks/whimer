package chat

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type ChatMemberP2PDao struct {
	db *xsql.DB
}

func NewChatMemberP2PDao(db *xsql.DB) *ChatMemberP2PDao {
	return &ChatMemberP2PDao{
		db: db,
	}
}

func (d *ChatMemberP2PDao) Create(ctx context.Context, member *ChatMemberP2PPO) error {
	member.Normalize()

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertIgnoreInto(chatMemberP2PPOTableName)
	ib.Cols(chatMemberP2PPoFieldsNoId...)
	ib.Values(member.ValuesNoId()...)

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ChatMemberP2PDao) GetByChatId(ctx context.Context, chatId uuid.UUID) (*ChatMemberP2PPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(chatMemberP2PPoFields...)
	sb.From(chatMemberP2PPOTableName)
	sb.Where(sb.Equal("chat_id", chatId))

	sql, args := sb.Build()

	var member ChatMemberP2PPO
	err := d.db.QueryRowCtx(ctx, &member, sql, args...)
	return &member, xsql.ConvertError(err)
}

func (d *ChatMemberP2PDao) GetByUids(ctx context.Context, uidA, uidB int64) (*ChatMemberP2PPO, error) {
	if uidA > uidB {
		uidA, uidB = uidB, uidA
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(chatMemberP2PPoFields...)
	sb.From(chatMemberP2PPOTableName)
	sb.Where(sb.Equal("uid_a", uidA), sb.Equal("uid_b", uidB))

	sql, args := sb.Build()

	var member ChatMemberP2PPO
	err := d.db.QueryRowCtx(ctx, &member, sql, args...)
	return &member, xsql.ConvertError(err)
}

func (d *ChatMemberP2PDao) GetByChatIdUid(ctx context.Context, chatId uuid.UUID, uid int64) (*ChatMemberP2PPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(chatMemberP2PPoFields...)
	sb.From(chatMemberP2PPOTableName)
	sb.Where(
		sb.Equal("chat_id", chatId),
		sb.Or(
			sb.Equal("uid_a", uid),
			sb.Equal("uid_b", uid),
		),
	)

	sql, args := sb.Build()
	var result ChatMemberP2PPO
	err := d.db.QueryRowCtx(ctx, &result, sql, args...)
	return &result, xsql.ConvertError(err)
}
