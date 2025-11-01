package chat

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type MsgDao struct {
	db *xsql.DB
}

func NewMsgDao(db *xsql.DB) *MsgDao {
	return &MsgDao{
		db: db,
	}
}

func (d *MsgDao) Create(ctx context.Context, m *MsgPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(msgPOTableName)
	ib.Cols(msgPOFields...)
	ib.Values(m.Values()...)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)

	return xsql.ConvertError(err)
}

func (d *MsgDao) GetByCid(ctx context.Context, cid string) (*MsgPO, error) {
	rb := sqlbuilder.NewSelectBuilder()
	rb.Select(msgPOFields...)
	rb.From(msgPOTableName)
	rb.Where(rb.EQ("cid", cid))

	sql, args := rb.Build()
	var m MsgPO
	err := d.db.QueryRowCtx(ctx, &m, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &m, nil
}

func (d *MsgDao) GetById(ctx context.Context, id uuid.UUID) (*MsgPO, error) {
	rb := sqlbuilder.NewSelectBuilder()
	rb.Select(msgPOFields...)
	rb.From(msgPOTableName)
	rb.Where(rb.EQ("id", id))

	sql, args := rb.Build()
	var m MsgPO
	err := d.db.QueryRowCtx(ctx, &m, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &m, nil
}

func (d *MsgDao) BatchGetByIds(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*MsgPO, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*MsgPO{}, nil
	}

	var msgs []MsgPO

	err := xslice.BatchExec(ids, 200, func(start, end int) error {
		targets := ids[start:end]
		idArgs := make([]any, 0, len(targets))
		for _, id := range targets {
			idArgs = append(idArgs, id)
		}

		rb := sqlbuilder.NewSelectBuilder()
		rb.Select(msgPOFields...)
		rb.From(msgPOTableName)
		rb.Where(rb.In("id", idArgs...))

		sql, args := rb.Build()
		var tmpMsgs []MsgPO
		err := d.db.QueryRowsCtx(ctx, &tmpMsgs, sql, args...)
		if err != nil {
			return xsql.ConvertError(err)
		}

		msgs = append(msgs, tmpMsgs...)

		return nil
	})
	if err != nil {
		return nil, err
	}

	msgMap := make(map[uuid.UUID]*MsgPO, len(ids))
	for _, m := range msgs {
		msgMap[m.Id] = &m
	}

	return msgMap, nil
}

func (d *MsgDao) UpdateStatus(ctx context.Context, id uuid.UUID, status model.MsgStatus, mtime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(msgPOTableName)
	ub.Set(ub.EQ("status", status), ub.EQ("mtime", mtime))
	ub.Where(ub.EQ("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)

	return xsql.ConvertError(err)
}
