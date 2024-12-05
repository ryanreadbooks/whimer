package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CommentDao struct {
	db          *xsql.DB
	cache       *CommentCache
	pinnedCache *xcache.Cache[*Comment] // 置顶评论缓存
	countCache  *xcache.Cache[uint64]   // 评论数量的缓存
}

func NewCommentDao(db *xsql.DB, cache *redis.Redis) *CommentDao {
	return &CommentDao{
		db:          db,
		cache:       NewCommentCache(cache),
		pinnedCache: xcache.New[*Comment](cache),
		countCache:  xcache.New[uint64](cache),
	}
}

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

	sqlPin   = "UPDATE comment SET pin=1, mtime=? WHERE id=? AND oid=? AND root=0"
	sqlUnpin = "UPDATE comment SET pin=0, mtime=? WHERE id=? AND oid=? AND root=0"
	// 一次性将已有的pin改为0，将目标id pin改为1
	sqlDoPin = `UPDATE comment SET pin=1-pin, mtime=? WHERE id=(
								SELECT id FROM (SELECT id FROM comment WHERE id>0 AND oid=? AND root=0 AND pin=1) tmp
							) OR id=?`

	sqlSetLike    = "UPDATE comment SET `like`=?, mtime=? WHERE id=?"
	sqlSetDisLike = "UPDATE comment SET dislike=?, mtime=? WHERE id=?"
	sqlSetReport  = "UPDATE comment SET report=?, mtime=? WHERE id=?"

	sqlDelById   = "DELETE FROM comment WHERE id=?"
	sqlDelByRoot = "DELETE FROM comment WHERE root=?"

	sqlFindUOIn = "SELECT DISTINCT uid, oid FROM comment WHERE uid IN (%s) AND oid IN (%s)"

	forUpdate = "FOR UPDATE"
)

var (
	sqlSelRootParentById = "SELECT id,root,parent,oid,pin FROM comment WHERE id=?"
	sqlCountByO          = "SELECT COUNT(*) FROM comment WHERE oid=?"
	sqlBatchCountByO     = "SELECT oid, COUNT(*) cnt FROM comment WHERE oid IN (%s) GROUP BY oid"
	sqlCountByOU         = "SELECT COUNT(*) FROM comment WHERE oid=? AND uid=?"
	sqlCountGbO          = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid"
	sqlCountGbOLimit     = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid LIMIT ?,?"
	sqlSelPinned         = fmt.Sprintf("SELECT %s FROM comment WHERE oid=? AND pin=1 LIMIT 1", fields)
	sqlSel               = fmt.Sprintf("SELECT %s FROM comment WHERE id=?", fields)
	sqlSel4Ud            = fmt.Sprintf("SELECT %s FROM comment WHERE id=? FOR UPDATE", fields)
	sqlSelByO            = fmt.Sprintf("SELECT %s FROM comment WHERE oid=? %%s", fields)
	sqlSelByRoot         = fmt.Sprintf("SELECT %s FROM comment WHERE root=? %%s", fields)
	sqlInsert            = fmt.Sprintf("INSERT INTO comment(%s) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", fields)

	sqlSelRoots = fmt.Sprintf("SELECT %s FROM comment WHERE %%s oid=? AND root=0 AND pin=0 ORDER BY ctime DESC LIMIT ?", fields)
	sqlSelSubs  = fmt.Sprintf("SELECT %s FROM comment WHERE id>? AND oid=? AND root=? ORDER BY ctime ASC LIMIT ?", fields)
)

func (r *CommentDao) FindByIdForUpdate(ctx context.Context, id uint64) (*Comment, error) {
	var res Comment
	err := r.db.QueryRowCtx(ctx, &res, sqlSel4Ud, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *CommentDao) FindRootParent(ctx context.Context, id uint64) (*RootParent, error) {
	var res RootParent
	err := r.db.QueryRowCtx(ctx, &res, sqlSelRootParentById, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *CommentDao) FindById(ctx context.Context, id uint64) (*Comment, error) {
	var res Comment
	err := r.db.QueryRowCtx(ctx, &res, sqlSel, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *CommentDao) Insert(ctx context.Context, model *Comment) (uint64, error) {
	if model.Ctime <= 0 {
		model.Ctime = time.Now().Unix()
	}

	if model.Mtime <= 0 {
		model.Mtime = model.Ctime
	}

	res, err := r.db.ExecCtx(ctx, sqlInsert,
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

func (r *CommentDao) DeleteById(ctx context.Context, id uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlDelById, id)
	return xsql.ConvertError(err)
}

func (r *CommentDao) DeleteByRoot(ctx context.Context, rootId uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlDelByRoot, rootId)
	return xsql.ConvertError(err)
}

func (r *CommentDao) findByOId(ctx context.Context, oid uint64, lock bool) ([]*Comment, error) {
	var rows = make([]*Comment, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByO, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByO, "")
	}

	err := r.db.QueryRowsCtx(ctx, &rows, sql, oid)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *CommentDao) FindByOid(ctx context.Context, oid uint64, lock bool) ([]*Comment, error) {
	return r.findByOId(ctx, oid, lock)
}

func (r *CommentDao) findByRootId(ctx context.Context, rootId uint64, lock bool) ([]*Comment, error) {
	var rows = make([]*Comment, 0)
	var sql string
	if lock {
		sql = fmt.Sprintf(sqlSelByRoot, forUpdate)
	} else {
		sql = fmt.Sprintf(sqlSelByRoot, "")
	}

	err := r.db.QueryRowsCtx(ctx, &rows, sql, rootId)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return rows, nil
}

func (r *CommentDao) FindByRootId(ctx context.Context, rootId uint64, lock bool) ([]*Comment, error) {
	return r.findByRootId(ctx, rootId, lock)
}

func (r *CommentDao) FindByParentId(ctx context.Context, rootId uint64, lock bool) ([]*Comment, error) {
	return r.findByRootId(ctx, rootId, lock)
}

func (r *CommentDao) updateCount(ctx context.Context, query string, id uint64) error {
	_, err := r.db.ExecCtx(ctx, query, time.Now().Unix(), id)
	return xsql.ConvertError(err)
}

func (r *CommentDao) AddLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlIncLike, id)
}

func (r *CommentDao) AddReport(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlIncReport, id)
}

func (r *CommentDao) AddDisLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlIncDislike, id)
}

func (r *CommentDao) SubLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlDecLike, id)
}

func (r *CommentDao) SubReport(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlDecReport, id)
}

func (r *CommentDao) SubDisLike(ctx context.Context, id uint64) error {
	return r.updateCount(ctx, sqlDecDislike, id)
}

func (r *CommentDao) setPin(ctx context.Context, oid, id uint64, pin bool) error {
	// 移除缓存
	defer func() {
		if _, err := r.pinnedCache.Del(ctx, getPinnedCmtKey(oid)); err != nil {
			xlog.Msg("pinned cache del pinned failed").Extra("oid", oid).Errorx(ctx)
		}
	}()

	var sql string
	if pin {
		sql = sqlPin
	} else {
		sql = sqlUnpin
	}
	_, err := r.db.ExecCtx(ctx, sql, time.Now().Unix(), id, oid)
	return xsql.ConvertError(err)
}

// Deprecated
func (r *CommentDao) SetPin(ctx context.Context, oid, id uint64) error {
	return r.setPin(ctx, oid, id, true)
}

// 取消置顶
func (r *CommentDao) SetUnPin(ctx context.Context, oid, id uint64) error {
	return r.setPin(ctx, oid, id, false)
}

// 获取主评论
func (r *CommentDao) GetRootReplies(ctx context.Context, oid, cursor uint64, want int) ([]*Comment, error) {
	var res = make([]*Comment, 0, want)
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
func (r *CommentDao) GetSubReplies(ctx context.Context, oid, root, cursor uint64, want int) ([]*Comment, error) {
	var res = make([]*Comment, 0, want)
	err := r.db.QueryRowsCtx(ctx, &res, sqlSelSubs, cursor, oid, root, want)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

// 置顶
func (r *CommentDao) DoPin(ctx context.Context, oid, rid uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlDoPin, time.Now().Unix(), oid, rid)
	defer func() {
		if _, err := r.pinnedCache.Del(ctx, getPinnedCmtKey(oid)); err != nil {
			xlog.Msg("pinned cache del pinned failed").Extra("oid", oid).Errorx(ctx)
		}
	}()

	return xsql.ConvertError(err)
}

// 拿出置顶评论
func (r *CommentDao) GetPinned(ctx context.Context, oid uint64) (*Comment, error) {
	model, err := r.pinnedCache.Get(ctx, getPinnedCmtKey(oid), xcache.WithGetFallback(
		func(ctx context.Context) (*Comment, int, error) {
			var ret Comment
			err := r.db.QueryRowCtx(ctx, &ret, sqlSelPinned, oid)
			if err != nil {
				return nil, 0, err
			}
			return &ret, xtime.WeekJitterSec(2 * time.Hour), nil
		},
	))

	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return model, nil
}

// 查出oid评论总量
func (r *CommentDao) CountByOid(ctx context.Context, oid uint64) (uint64, error) {
	cnt, err := r.countCache.Get(ctx, getCountCmtKey(oid), xcache.WithGetFallback(
		func(ctx context.Context) (uint64, int, error) {
			var cnt uint64
			err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByO, oid)
			if err != nil {
				return 0, 0, nil
			}
			return cnt, xtime.HourJitterSec(3 * time.Hour), nil
		},
	))

	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *CommentDao) IncrReplyCount(ctx context.Context, oid uint64) error {
	return r.cache.IncrReplyCount(ctx, oid, 1)
}

// TODO 注意小于0的情况发生
func (r *CommentDao) DecrReplyCount(ctx context.Context, oid uint64) error {
	return r.cache.DecrReplyCount(ctx, oid, 1)
}

func (r *CommentDao) BatchCountByOid(ctx context.Context, oids []uint64) (map[uint64]uint64, error) {
	var ret []struct {
		Oid uint64 `db:"oid"`
		Cnt uint64 `db:"cnt"`
	}

	query := fmt.Sprintf(sqlBatchCountByO, slices.JoinInts(oids))
	err := r.db.QueryRowsCtx(ctx, &ret, query)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	result := make(map[uint64]uint64, len(ret))
	for _, r := range ret {
		result[r.Oid] = r.Cnt
	}

	r.cache.BatchSetReplyCount(ctx, result)

	return result, nil
}

func (r *CommentDao) CountByOidUid(ctx context.Context, oid, uid uint64) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByOU, oid, uid)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *CommentDao) CountGroupByOid(ctx context.Context) (map[uint64]uint64, error) {
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

	r.cache.BatchSetReplyCount(ctx, ret)

	return ret, nil
}

func (r *CommentDao) CountGroupByOidLimit(ctx context.Context, offset, limit int64) (map[uint64]uint64, error) {
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

	r.cache.BatchSetReplyCount(ctx, ret)

	return ret, nil
}

type UidOid struct {
	Uid uint64
	Oid uint64
}

// uid -> []oids
func (r *CommentDao) FindByUidsOids(ctx context.Context, uidOids map[uint64][]uint64) ([]UidOid, error) {
	var batchRes []UidOid
	// 分批操作
	err := maps.BatchExec(uidOids, 200, func(target map[uint64][]uint64) error {
		uids, oids := maps.All(target)
		var allOids []uint64 = oids[0]
		for i := 1; i < len(oids); i++ {
			allOids = slices.Concat(allOids, oids[i])
		}

		var ret = make([]UidOid, 0, len(uids)*len(allOids)) // we should strictly limit the length of them
		query := fmt.Sprintf(sqlFindUOIn, slices.JoinInts(uids), slices.JoinInts(allOids))
		err := r.db.QueryRowsCtx(ctx, &ret, query)
		if err != nil {
			return err
		}

		batchRes = append(batchRes, ret...)
		return nil
	})

	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return batchRes, nil
}
