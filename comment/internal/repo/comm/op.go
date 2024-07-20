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
	fields     = "id,note_id,ctype,content,uid,parent_id,reply_id,reply_uid,state,like_count,dislike_count,report_count,is_top,client_info,ctime"
	fieldsNoId = "note_id,ctype,content,uid,parent_id,reply_id,reply_uid,state,like_count,dislike_count,report_count,is_top,client_info,ctime"

	sqlUdState    = "update comment set state=? where id=?"
	sqlAddLike    = "update comment set like_count=like_count+1 where id=?"
	sqlSubLike    = "update comment set like_count=like_count-1 where id=?"
	sqlAddDislike = "update comment set dislike_count=dislike_count+1 where id=?"
	sqlSubDislike = "update comment set dislike_count=dislike_count-1 where id=?"
	sqlAddReport  = "update comment set report_count=report_count+1 where id=?"
	sqlSubReport  = "update comment set report_count=report_count-1 where id=?"
	sqlSetIsTop   = "update comment set is_top=1 where id=?"

	forUpdate = "for update"
)

var (
	sqlSelByNote   = fmt.Sprintf("select %s from comment where note_id=? %%s", fields)
	sqlSelByParent = fmt.Sprintf("select %s from comment where parent_id=? %%s", fields)
	sqlInsert      = fmt.Sprintf("insert into comment(%s) values(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", fieldsNoId)
)

func (r *Repo) insert(ctx context.Context, sess sqlx.Session, model *Model) (uint64, error) {
	if model.Ctime <= 0 {
		model.Ctime = time.Now().Unix()
	}

	res, err := sess.ExecCtx(ctx, sqlInsert,
		model.NoteId,
		model.CType,
		model.Content,
		model.Uid,
		model.ParentId,
		model.ReplyId,
		model.ReplyUid,
		model.State,
		model.Like,
		model.Dislike,
		model.Report,
		model.IsTop,
		model.ClientInfo,
		model.Ctime)
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
	return r.udCount(ctx, r.db, sqlAddLike, id)
}

func (r *Repo) AddLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlAddLike, id)
}

func (r *Repo) AddReport(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlAddReport, id)
}

func (r *Repo) AddReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlAddReport, id)
}

func (r *Repo) AddDisLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlAddDislike, id)
}

func (r *Repo) AddDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlAddDislike, id)
}

func (r *Repo) SubLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlSubLike, id)
}

func (r *Repo) SubLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlSubLike, id)
}

func (r *Repo) SubReport(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlSubReport, id)
}

func (r *Repo) SubReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlSubReport, id)
}

func (r *Repo) SubDisLike(ctx context.Context, id uint64) error {
	return r.udCount(ctx, r.db, sqlSubDislike, id)
}

func (r *Repo) SubDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.udCount(ctx, tx, sqlSubDislike, id)
}

func (r *Repo) setTop(ctx context.Context, sess sqlx.Session, id uint64) error {
	_, err := sess.ExecCtx(ctx, sqlSetIsTop, id)
	return xsql.ConvertError(err)
}

func (r *Repo) SetTop(ctx context.Context, id uint64) error {
	return r.setTop(ctx, r.db, id)
}

func (r *Repo) SetTopTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.setTop(ctx, tx, id)
}
