package note

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/global"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	sqlFind        = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE id=?"
	sqlInsertAll   = "INSERT INTO note(title,`desc`,privacy,owner,create_at,update_at) VALUES(?,?,?,?,?,?)"
	sqlUpdateAll   = "UPDATE note SET title=?,`desc`=?,privacy=?,owner=?,update_at=? WHERE id=?"
	sqlDeleteById  = "DELETE FROM note WHERE id=?"
	sqlListByOwner = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE owner=?"
	sqlGetByCursor = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE id>=? AND privacy=? LIMIT ?"
	sqlGetLastId   = "SELECT id FROM note WHERE privacy=? ORDER BY id DESC LIMIT 1"
	sqlGetAll      = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE privacy=?"
	sqlGetCount    = "SELECT COUNT(*) FROM note WHERE privacy=?"
)

func (r *Repo) FindOne(ctx context.Context, id uint64) (*Model, error) {
	model := new(Model)
	err := r.db.QueryRowCtx(ctx, model, sqlFind, id)
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

func (r *Repo) GetPublicByCursor(ctx context.Context, id uint64, count int) ([]*Model, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPublic)
}

func (r *Repo) GetPrivateByCursor(ctx context.Context, id uint64, count int) ([]*Model, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPrivate)
}

func (r *Repo) getByCursor(ctx context.Context, id uint64, count, privacy int) ([]*Model, error) {
	var res = make([]*Model, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetByCursor, id, privacy, count)
	return res, xsql.ConvertError(err)
}

func (r *Repo) GetPublicLastId(ctx context.Context) (uint64, error) {
	return r.getLastId(ctx, global.PrivacyPublic)
}

func (r *Repo) GetPrivateLastId(ctx context.Context) (uint64, error) {
	return r.getLastId(ctx, global.PrivacyPrivate)
}

func (r *Repo) getLastId(ctx context.Context, privacy int) (uint64, error) {
	var lastId uint64
	err := r.db.QueryRowCtx(ctx, &lastId, sqlGetLastId, privacy)
	return lastId, xsql.ConvertError(err)
}

func (r *Repo) getAll(ctx context.Context, privacy int) ([]*Model, error) {
	var res = make([]*Model, 0, 16)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetAll, privacy)
	return res, xsql.ConvertError(err)
}

func (r *Repo) GetPublicAll(ctx context.Context) ([]*Model, error) {
	return r.getAll(ctx, global.PrivacyPublic)
}

func (r *Repo) GetPrivateAll(ctx context.Context) ([]*Model, error) {
	return r.getAll(ctx, global.PrivacyPrivate)
}

func (r *Repo) GetPublicCount(ctx context.Context) (uint64, error) {
	return r.getCount(ctx, global.PrivacyPublic)
}

func (r *Repo) GetPrivateCount(ctx context.Context) (uint64, error) {
	return r.getCount(ctx, global.PrivacyPrivate)
}

func (r *Repo) getCount(ctx context.Context, privacy int) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlGetCount, privacy)
	return cnt, xsql.ConvertError(err)
}
