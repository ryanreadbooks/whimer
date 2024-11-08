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
	fields     = "id,oid,ctype,content,uid,root,parent,ruid,state,`like`,dislike,report,pin,ip,ctime,mtime"
	fieldsNoId = "oid,ctype,content,uid,root,parent,ruid,state,like,dislike,report,pin,ip,ctime,mtime"

	sqlUdState    = "UPDATE comment SET state=?, mtime=? WHERE id=?"
	sqlIncLike    = "UPDATE comment SET `like`=`like`+1, mtime=? WHERE id=?"
	sqlDecLike    = "UPDATE comment SET `like`=`like`-1, mtime=? WHERE id=?"
	sqlIncDislike = "UPDATE comment SET dislike=dislike+1, mtime=? WHERE id=?"
	sqlDecDislike = "UPDATE comment SET dislike=dislike-1, mtime=? WHERE id=?"
	sqlIncReport  = "UPDATE comment SET report=report+1, mtime=? WHERE id=?"
	sqlDecReport  = "UPDATE comment SET report=report-1, mtime=? WHERE id=?"

	sqlPin   = "UPDATE comment SET pin=1, mtime=? WHERE id=? AND root=0"
	sqlUnpin = "UPDATE comment SET pin=0, mtime=? WHERE id=? AND root=0"
	// 一次性将已有的pin改为0，将目标idpin改为1
	sqlDoPin = `UPDATE comment SET pin=1-pin, mtime=? WHERE id=(
								SELECT id FROM (SELECT id FROM comment WHERE id>0 AND oid=? AND root=0 AND pin=1) tmp
							) OR id=?`

	sqlSetLike    = "UPDATE comment SET `like`=?, mtime=? WHERE id=?"
	sqlSetDisLike = "UPDATE comment SET dislike=?, mtime=? WHERE id=?"
	sqlSetReport  = "UPDATE comment SET report=?, mtime=? WHERE id=?"

	sqlDelById   = "DELETE FROM comment WHERE id=?"
	sqlDelByRoot = "DELETE FROM comment WHERE root=?"

	forUpdate = "FOR UPDATE"
)

var (
	sqlSelRooParent  = "SELECT id,root,parent,oid,pin FROM comment WHERE id=?"
	sqlCountByO      = "SELECT COUNT(*) FROM comment WHERE oid=?"
	sqlCountByOU     = "SELECT COUNT(*) FROM comment WHERE oid=? AND uid=?"
	sqlCountGbO      = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid"
	sqlCountGbOLimit = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid LIMIT ?,?"
	sqlSelPinned     = fmt.Sprintf("SELECT %s FROM comment WHERE oid=? AND pin=1 LIMIT 1", fields)
	sqlSel           = fmt.Sprintf("SELECT %s FROM comment WHERE id=?", fields)
	sqlSel4Ud        = fmt.Sprintf("SELECT %s FROM comment WHERE id=? FOR UPDATE", fields)
	sqlSelByO        = fmt.Sprintf("SELECT %s FROM comment WHERE oid=? %%s", fields)
	sqlSelByRoot     = fmt.Sprintf("SELECT %s FROM comment WHERE root=? %%s", fields)
	sqlInsert        = fmt.Sprintf("INSERT INTO comment(%s) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", fields)

	sqlSelRoots = fmt.Sprintf("SELECT %s FROM comment WHERE %%s oid=? AND root=0 AND pin=0 ORDER BY ctime DESC LIMIT ?", fields)
	sqlSelSubs  = fmt.Sprintf("SELECT %s FROM comment WHERE id>? AND oid=? AND root=? ORDER BY ctime ASC LIMIT ?", fields)
)

func (r *Repo) FindByIdForUpdate(ctx context.Context, tx sqlx.Session, id uint64) (*Model, error) {
	var res Model
	err := tx.QueryRowCtx(ctx, &res, sqlSel4Ud, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *Repo) FindRootParent(ctx context.Context, id uint64) (*RootParent, error) {
	var res RootParent
	err := r.db.QueryRowCtx(ctx, &res, sqlSelRooParent, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *Repo) FindById(ctx context.Context, id uint64) (*Model, error) {
	var res Model
	err := r.db.QueryRowCtx(ctx, &res, sqlSel, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *Repo) insert(ctx context.Context, sess sqlx.Session, model *Model) (uint64, error) {
	if model.Ctime <= 0 {
		model.Ctime = time.Now().Unix()
	}

	if model.Mtime <= 0 {
		model.Mtime = model.Ctime
	}

	res, err := sess.ExecCtx(ctx, sqlInsert,
		model.Id,
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

func (r *Repo) delete(ctx context.Context, sess sqlx.Session, id uint64) error {
	_, err := sess.ExecCtx(ctx, sqlDelById, id)
	return xsql.ConvertError(err)
}

func (r *Repo) DeleteById(ctx context.Context, id uint64) error {
	return r.delete(ctx, r.db, id)
}

func (r *Repo) DeleteByIdTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.delete(ctx, tx, id)
}

func (r *Repo) DeleteByRootTx(ctx context.Context, tx sqlx.Session, rootId uint64) error {
	_, err := tx.ExecCtx(ctx, sqlDelByRoot, rootId)
	return xsql.ConvertError(err)
}

func (r *Repo) findByOId(ctx context.Context, sess sqlx.Session, oid uint64, lock bool) ([]*Model, error) {
	var rows = make([]*Model, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByO, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByO, "")
	}

	err := sess.QueryRowsCtx(ctx, &rows, sql, oid)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *Repo) FindByOid(ctx context.Context, oid uint64) ([]*Model, error) {
	return r.findByOId(ctx, r.db, oid, false)
}

func (r *Repo) FindByOidTx(ctx context.Context, tx sqlx.Session, oid uint64, lock bool) ([]*Model, error) {
	return r.findByOId(ctx, tx, oid, lock)
}

func (r *Repo) findByRootId(ctx context.Context, sess sqlx.Session, rootId uint64, lock bool) ([]*Model, error) {
	var rows = make([]*Model, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByRoot, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByRoot, "")
	}

	err := sess.QueryRowsCtx(ctx, &rows, sql, rootId)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *Repo) FindByRootId(ctx context.Context, rootId uint64) ([]*Model, error) {
	return r.findByRootId(ctx, r.db, rootId, false)
}

func (r *Repo) FindByParentIdTx(ctx context.Context, tx sqlx.Session, rootId uint64, lock bool) ([]*Model, error) {
	return r.findByRootId(ctx, tx, rootId, lock)
}

func (r *Repo) updateCount(ctx context.Context, sess sqlx.Session, query string, id uint64) error {
	_, err := sess.ExecCtx(ctx, query, time.Now().Unix(), id)
	return xsql.ConvertError(err)
}

func (r *Repo) AddLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlIncLike, id)
}

func (r *Repo) AddLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlIncLike, id)
}

func (r *Repo) AddReport(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlIncReport, id)
}

func (r *Repo) AddReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlIncReport, id)
}

func (r *Repo) AddDisLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlIncDislike, id)
}

func (r *Repo) AddDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlIncDislike, id)
}

func (r *Repo) SubLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlDecLike, id)
}

func (r *Repo) SubLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlDecLike, id)
}

func (r *Repo) SubReport(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlDecReport, id)
}

func (r *Repo) SubReportTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlDecReport, id)
}

func (r *Repo) SubDisLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, r.db, sqlDecDislike, id)
}

func (r *Repo) SubDisLikeTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.updateCount(ctx, tx, sqlDecDislike, id)
}

func (r *Repo) setPin(ctx context.Context, sess sqlx.Session, id uint64, pin bool) error {
	var sql string
	if pin {
		sql = sqlPin
	} else {
		sql = sqlUnpin
	}
	_, err := sess.ExecCtx(ctx, sql, time.Now().Unix(), id)
	return xsql.ConvertError(err)
}

func (r *Repo) SetPin(ctx context.Context, id uint64) error {
	return r.setPin(ctx, r.db, id, true)
}

func (r *Repo) SetPinTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.setPin(ctx, tx, id, true)
}

func (r *Repo) SetUnPin(ctx context.Context, id uint64) error {
	return r.setPin(ctx, r.db, id, false)
}

func (r *Repo) SetUnPinTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.setPin(ctx, tx, id, false)
}

// 获取主评论
func (r *Repo) GetRootReplies(ctx context.Context, oid, cursor uint64, want int) ([]*Model, error) {
	var res = make([]*Model, 0, want)
	hasCursor := ""
	var args []any
	if cursor > 0 {
		hasCursor = "id<? AND"
		args = []any{cursor, oid, want}
	} else {
		args = []any{oid, want}
	}
	err := r.db.QueryRowsCtx(ctx, &res, fmt.Sprintf(sqlSelRoots, hasCursor), args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

// 获取子评论
func (r *Repo) GetSubReply(ctx context.Context, oid, root, cursor uint64, want int) ([]*Model, error) {
	var res = make([]*Model, 0, want)
	err := r.db.QueryRowsCtx(ctx, &res, sqlSelSubs, cursor, oid, root, want)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

// 置顶
func (r *Repo) DoPin(ctx context.Context, oid, rid uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlDoPin, time.Now().Unix(), oid, rid)
	return xsql.ConvertError(err)
}

func (r *Repo) GetPinned(ctx context.Context, oid uint64) (*Model, error) {
	var model = new(Model)
	err := r.db.QueryRowCtx(ctx, model, sqlSelPinned, oid)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return model, nil
}

// 查出oid评论总量
func (r *Repo) CountByOid(ctx context.Context, oid uint64) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByO, oid)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) CountByOidUid(ctx context.Context, oid, uid uint64) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByOU, oid, uid)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) CountGroupByOid(ctx context.Context) (map[uint64]uint64, error) {
	var res []struct {
		Oid uint64 `db:"oid"`
		Cnt uint64 `db:"cnt"`
	}
	err := r.db.QueryRowsCtx(ctx, &res, sqlCountGbO)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	ret := make(map[uint64]uint64, len(res))
	for _, item := range res {
		ret[item.Oid] = item.Cnt
	}

	return ret, nil
}

func (r *Repo) CountGroupByOidLimit(ctx context.Context, offset, limit int64) (map[uint64]uint64, error) {
	var res []struct {
		Oid uint64 `db:"oid"`
		Cnt uint64 `db:"cnt"`
	}
	err := r.db.QueryRowsCtx(ctx, &res, sqlCountGbOLimit, offset, limit)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	ret := make(map[uint64]uint64, len(res))
	for _, item := range res {
		ret[item.Oid] = item.Cnt
	}

	return ret, nil
}
