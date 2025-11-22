package chat

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type ChatMsgDao struct {
	db *xsql.DB
}

func NewChatMsgDao(db *xsql.DB) *ChatMsgDao {
	return &ChatMsgDao{
		db: db,
	}
}

func (d *ChatMsgDao) Create(ctx context.Context, cm *ChatMsgPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertIgnoreInto(chatMsgPOTableName)
	ib.Cols(chatMsgPOFields...)
	ib.Values(cm.Values()...)

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *ChatMsgDao) ListByPos(ctx context.Context,
	chatId uuid.UUID, cursor int64, count int32, desc bool) ([]*ChatMsgPO, error) {

	ib := sqlbuilder.NewSelectBuilder()
	ib.Select(chatMsgPOFields...)
	ib.From(chatMsgPOTableName)
	ib.Where(ib.EQ("chat_id", chatId))
	if desc {
		ib.OrderByDesc("pos")
		ib.Where(ib.LT("pos", cursor))
	} else {
		ib.OrderByAsc("pos")
		ib.Where(ib.GT("pos", cursor))
	}
	ib.Limit(int(count))

	sql, args := ib.Build()

	var msgs []*ChatMsgPO
	err := d.db.QueryRowsCtx(ctx, &msgs, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return msgs, nil
}

func (d *ChatMsgDao) BatchGetPos(ctx context.Context,
	chatId uuid.UUID, msgIds []uuid.UUID) (map[uuid.UUID]int64, error) {
	if len(msgIds) == 0 {
		return map[uuid.UUID]int64{}, nil
	}

	var pos []*ChatMsgPO_MsgIdPos = make([]*ChatMsgPO_MsgIdPos, 0, len(msgIds))
	err := xslice.BatchExec(msgIds, 100, func(start, end int) error {
		targetMsgIds := msgIds[start:end]
		sb := sqlbuilder.NewSelectBuilder()
		sb.Select("msg_id", "pos").
			From(chatMsgPOTableName).
			Where(sb.EQ("chat_id", chatId), sb.In("msg_id", xslice.Any(targetMsgIds)...))

		sql, args := sb.Build()

		var tmpPos []*ChatMsgPO_MsgIdPos
		err := d.db.QueryRowsCtx(ctx, &tmpPos, sql, args...)
		if err != nil {
			return xsql.ConvertError(err)
		}

		pos = append(pos, tmpPos...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]int64, len(pos))
	for _, p := range pos {
		result[p.MsgId] = p.Pos
	}

	return result, nil
}
