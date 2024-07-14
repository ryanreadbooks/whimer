package note

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	sqlFindAll     = "select id,title,`desc`,privacy,owner,create_at,update_at from note where id=?"
	sqlInsertAll   = "insert into note(title,`desc`,privacy,owner,create_at,update_at) values(?,?,?,?,?,?)"
	sqlUpdateAll   = "update note set title=?,`desc`=?,privacy=?,owner=?,update_at=? where id=?"
	sqlDeleteById  = "delete from note where id=?"
	sqlListByOwner = "select id,title,`desc`,privacy,owner,create_at,update_at from note where owner=?"
)

func (r *Repo) FindOne(ctx context.Context, id uint64) (*Model, error) {
	model := new(Model)
	err := r.db.QueryRowCtx(ctx, model, sqlFindAll, id)
	return model, xsql.ConvertError(err)
}

func (r *Repo) ListByOwner(ctx context.Context, uid uint64) ([]*Model, error) {
	res := make([]*Model, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListByOwner, uid)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}
	return res, nil
}

func (r *Repo) insert(ctx context.Context, sess sqlx.Session, note *Model) (uint64, error) {
	now := time.Now().Unix()
	res, err := sess.ExecCtx(ctx,
		sqlInsertAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		now,
		now)

	if err != nil {
		return 0, xsql.ConvertError(err)
	}
	newId, _ := res.LastInsertId()
	return uint64(newId), nil
}

func (r *Repo) Insert(ctx context.Context, note *Model) (uint64, error) {
	return r.insert(ctx, r.db, note)
}

func (r *Repo) InsertTx(ctx context.Context, tx sqlx.Session, note *Model) (uint64, error) {
	return r.insert(ctx, tx, note)
}

func (r *Repo) update(ctx context.Context, sess sqlx.Session, note *Model) error {
	_, err := sess.ExecCtx(ctx,
		sqlUpdateAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		time.Now().Unix(),
		note.Id,
	)

	return xsql.ConvertError(err)
}

func (r *Repo) Update(ctx context.Context, note *Model) error {
	return r.update(ctx, r.db, note)
}

func (r *Repo) UpdateTx(ctx context.Context, tx sqlx.Session, note *Model) error {
	return r.update(ctx, tx, note)
}

func (r *Repo) delete(ctx context.Context, sess sqlx.Session, id uint64) error {
	_, err := sess.ExecCtx(ctx,
		sqlDeleteById, id)
	return xsql.ConvertError(err)
}

func (r *Repo) Delete(ctx context.Context, id uint64) error {
	return r.delete(ctx, r.db, id)
}

func (r *Repo) DeleteTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.delete(ctx, tx, id)
}
