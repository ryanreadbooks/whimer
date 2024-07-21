package comm

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	fields     = "id,oid,ctype,content,uid,root,parent,ruid,state,`like`,dislike,report,pin,cli,ctime,mtime"
	fieldsNoId = "oid,ctype,content,uid,root,parent,ruid,state,like,dislike,report,pin,cli,ctime,mtime"

	sqlUdState    = "UPDATE comment SET state=? WHERE id=?"
	sqlIncLike    = "UPDATE comment SET `like`=`like`+1 WHERE id=?"
	sqlDecLike    = "UPDATE comment SET `like`=`like`-1 WHERE id=?"
	sqlIncDislike = "UPDATE comment SET dislike=dislike+1 WHERE id=?"
	sqlDecDislike = "UPDATE comment SET dislike=dislike-1 WHERE id=?"
	sqlIncReport  = "UPDATE comment SET report=report+1 WHERE id=?"
	sqlDecReport  = "UPDATE comment SET report=report-1 WHERE id=?"
	sqlSetPin     = "UPDATE comment SET pin=1 WHERE id=?"

	sqlSetLike    = "UPDATE comment SET `like`=? WHERE id=?"
	sqlSetDisLike = "UPDATE comment SET dislike=? WHERE id=?"
	sqlSetReport  = "UPDATE comment SET report=? WHERE id=?"

	forUpdate = "FOR UPDATE"
)

var (
	sqlSelByNote   = fmt.Sprintf("SELECT %s FROM comment WHERE oid=? %%s", fields)
	sqlSelByParent = fmt.Sprintf("SELECT %s FROM comment WHERE root=? %%s", fields)
	sqlInsert      = fmt.Sprintf("INSERT INTO comment(%s) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", fieldsNoId)
)

func (r *Repo) insert(ctx context.Context, sess sqlx.Session, model *Model) (uint64, error) {
	if model.Ctime <= 0 {
		model.Ctime = time.Now().Unix()
	}

	if model.Mtime <= 0 {
		model.Mtime = model.Ctime
	}

	res, err := sess.ExecCtx(ctx, sqlInsert,
		model.Oid,
		model.CType,
		model.Content,
		model.Uid,
		model.RootId,
		model.ParentId,
		model.ReplyUid,
		model.State,
		model.Like,
		model.Dislike,
		model.Report,
		model.IsPin,
		model.Ip,
		model.Ctime,
		model.Mtime)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	newId, _ := res.LastInsertId()
	return uint64(newId), nil
}

func (r *Repo) Insert(ctx context.Context, model *Model) (uint64, error) {
	return r.insert(ctx, r.db, model)
}

func (r *Repo) InsertTx(ctx context.Context, tx sqlx.Session, model *Model) (uint64, error) {
	return r.insert(ctx, tx, model)
}

func (r *Repo) findByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64, lock bool) ([]*Model, error) {
	var rows = make([]*Model, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByNote, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByNote, "")
	}

	err := sess.QueryRowsCtx(ctx, &rows, sql, noteId)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *Repo) FindByNoteId(ctx context.Context, noteId uint64) ([]*Model, error) {
	return r.findByNoteId(ctx, r.db, noteId, false)
}

func (r *Repo) FindByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64, lock bool) ([]*Model, error) {
	return r.findByNoteId(ctx, tx, noteId, lock)
}

func (r *Repo) findByParentId(ctx context.Context, sess sqlx.Session, parentId uint64, lock bool) ([]*Model, error) {
	var rows = make([]*Model, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByParent, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByParent, "")
	}

	err := sess.QueryRowsCtx(ctx, &rows, sql, parentId)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *Repo) FindByParentId(ctx context.Context, parentId uint64) ([]*Model, error) {
	return r.findByParentId(ctx, r.db, parentId, false)
}

func (r *Repo) FindByParentIdTx(ctx context.Context, tx sqlx.Session, parentId uint64, lock bool) ([]*Model, error) {
	return r.findByParentId(ctx, tx, parentId, lock)
}

func (r *Repo) udCount(ctx context.Context, sess sqlx.Session, query string, id uint64) error {
	_, err := sess.ExecCtx(ctx, query, id)
	return xsql.ConvertError(err)
}

func (r *Repo) AddLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlIncLike, id)
}

func (r *Repo) AddLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlIncLike, id)
}

func (r *Repo) AddReport(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlIncReport, id)
}

func (r *Repo) AddReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlIncReport, id)
}

func (r *Repo) AddDisLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlIncDislike, id)
}

func (r *Repo) AddDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlIncDislike, id)
}

func (r *Repo) SubLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlDecLike, id)
}

func (r *Repo) SubLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlDecLike, id)
}

func (r *Repo) SubReport(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlDecReport, id)
}

func (r *Repo) SubReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlDecReport, id)
}

func (r *Repo) SubDisLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlDecDislike, id)
}

func (r *Repo) SubDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlDecDislike, id)
}

func (r *Repo) setTop(ctx context.Context, sess sqlx.Session, id uint64) error {
	_, err := sess.ExecCtx(ctx, sqlSetPin, id)
	return xsql.ConvertError(err)
}

func (r *Repo) SetTop(ctx context.Context, id uint64) error {
	return r.setTop(ctx, r.db, id)
}

func (r *Repo) SetTopTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.setTop(ctx, tx, id)
}
