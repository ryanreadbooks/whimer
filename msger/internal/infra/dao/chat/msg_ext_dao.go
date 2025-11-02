package chat

import (
	"context"
	"encoding/json"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type MsgExtDao struct {
	db *xsql.DB
}

func NewMsgExtDao(db *xsql.DB) *MsgExtDao {
	return &MsgExtDao{
		db: db,
	}
}

func (d *MsgExtDao) Create(ctx context.Context, ext *MsgExtPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertIgnoreInto(msgExtPOTableName)
	ib.Cols(msgExtPOFields...)
	ib.Values(ext.Values()...)

	sql, args := ib.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *MsgExtDao) GetById(ctx context.Context, msgId uuid.UUID) (*MsgExtPO, error) {
	ib := sqlbuilder.NewSelectBuilder()
	ib.Select(msgExtPOFields...)
	ib.From(msgExtPOTableName)
	ib.Where(ib.EQ("msg_id", msgId))

	sql, args := ib.Build()

	var ext MsgExtPO
	err := d.db.QueryRowCtx(ctx, &ext, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &ext, nil
}

func (d *MsgExtDao) BatchGetByIds(ctx context.Context, msgIds []uuid.UUID) (map[uuid.UUID]*MsgExtPO, error) {
	if len(msgIds) == 0 {
		return map[uuid.UUID]*MsgExtPO{}, nil
	}

	var exts []MsgExtPO

	err := xslice.BatchExec(msgIds, 200, func(start, end int) error {
		targets := msgIds[start:end]
		idArgs := make([]any, 0, len(targets))
		for _, id := range targets {
			idArgs = append(idArgs, id)
		}

		rb := sqlbuilder.NewSelectBuilder()
		rb.Select(msgExtPOFields...)
		rb.From(msgExtPOTableName)
		rb.Where(rb.In("msg_id", idArgs...))

		sql, args := rb.Build()
		var tmpExts []MsgExtPO
		err := d.db.QueryRowsCtx(ctx, &tmpExts, sql, args...)
		if err != nil {
			return xsql.ConvertError(err)
		}

		exts = append(exts, tmpExts...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	msgMap := make(map[uuid.UUID]*MsgExtPO, len(msgIds))
	for _, m := range exts {
		msgMap[m.MsgId] = &m
	}

	return msgMap, nil
}

func (d *MsgExtDao) SetRecall(ctx context.Context, id uuid.UUID, recall json.RawMessage) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(msgExtPOTableName)
	ub.Set(ub.EQ("recall", recall))
	ub.Where(ub.EQ("msg_id", id))

	sql, args := ub.Build()

	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}
