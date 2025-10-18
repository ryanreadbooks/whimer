package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xcache"
	xcachev2 "github.com/ryanreadbooks/whimer/misc/xcache/v2"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	slices "github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CommentDao struct {
	db          *xsql.DB
	cache       *CommentCache
	cacheV2     *xcachev2.Cache[*Comment]
	pinnedCache *xcache.Cache[*Comment] // 置顶评论缓存
	countCache  *xcache.Cache[int64]    // 评论数量的缓存
}

func NewCommentDao(db *xsql.DB, rd *redis.Redis) *CommentDao {
	return &CommentDao{
		db:          db,
		cache:       NewCommentCache(rd),
		cacheV2:     xcachev2.New[*Comment](rd),
		pinnedCache: xcache.New[*Comment](rd),
		countCache:  xcache.New[int64](rd),
	}
}

// all sqls here
const (
	fields          = "id,oid,type,content,uid,root,parent,ruid,state,`like`,dislike,report,pin,ip,ctime,mtime"
	fieldsWithoutId = "oid,type,content,uid,root,parent,ruid,state,`like`,dislike,report,pin,ip,ctime,mtime"

	forUpdate = "FOR UPDATE"

	sqlUdState    = "UPDATE comment SET state=?, mtime=? WHERE id=?"
	sqlIncLike    = "UPDATE comment SET `like`=`like`+1, mtime=? WHERE id=?"
	sqlDecLike    = "UPDATE comment SET `like`=`like`-1, mtime=? WHERE id=?"
	sqlIncDislike = "UPDATE comment SET dislike=dislike+1, mtime=? WHERE id=?"
	sqlDecDislike = "UPDATE comment SET dislike=dislike-1, mtime=? WHERE id=?"
	sqlIncReport  = "UPDATE comment SET report=report+1, mtime=? WHERE id=?"
	sqlDecReport  = "UPDATE comment SET report=report-1, mtime=? WHERE id=?"
	sqlPin        = "UPDATE comment SET pin=1, mtime=? WHERE id=? AND oid=? AND root=0"
	sqlUnpin      = "UPDATE comment SET pin=0, mtime=? WHERE id=? AND oid=? AND root=0"
	sqlSetLike    = "UPDATE comment SET `like`=?, mtime=? WHERE id=?"
	sqlSetDisLike = "UPDATE comment SET dislike=?, mtime=? WHERE id=?"
	sqlSetReport  = "UPDATE comment SET report=?, mtime=? WHERE id=?"

	sqlDelById   = "DELETE FROM comment WHERE id=?"
	sqlDelByRoot = "DELETE FROM comment WHERE root=?"

	sqlFindUOIn = "SELECT DISTINCT uid, oid FROM comment WHERE uid IN (%s) AND oid IN (%s)"

	// 一次性将已有的pin改为0，将目标id pin改为1
	sqlDoPin = `UPDATE comment SET pin=1-pin, mtime=? WHERE id=(
								SELECT id FROM (SELECT id FROM comment WHERE id>0 AND oid=? AND root=0 AND pin=1) tmp
							) OR id=?`
)

const (
	sqlSelRootParentById = "SELECT id,root,parent,oid,pin FROM comment WHERE id=?"
	sqlCountByO          = "SELECT COUNT(*) FROM comment WHERE oid=?"
	sqlBatchCountByO     = "SELECT oid, COUNT(*) AS cnt FROM comment WHERE oid IN (%s) GROUP BY oid"
	sqlCountByOU         = "SELECT COUNT(*) FROM comment WHERE oid=? AND uid=?"
	sqlCountGbO          = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid"
	sqlCountGbOLimit     = "SELECT oid, COUNT(*) AS cnt FROM comment GROUP BY oid LIMIT ?,?"
	sqlSelPinned         = "SELECT " + fields + " FROM comment WHERE oid=? AND pin=1 LIMIT 1"
	sqlSel               = "SELECT " + fields + " FROM comment WHERE id=?"
	sqlSel4Ud            = "SELECT " + fields + " FROM comment WHERE id=? FOR UPDATE"
	sqlSelByO            = "SELECT " + fields + " FROM comment WHERE oid=? %s"
	sqlSelByRoot         = "SELECT " + fields + " FROM comment WHERE root=? %s"
	sqlSelRoots          = "SELECT " + fields + " FROM comment WHERE %s oid=? AND root=0 AND pin=0 ORDER BY ctime DESC LIMIT ?"
	sqlSelSubs           = "SELECT " + fields + " FROM comment WHERE id>? AND oid=? AND root=? ORDER BY ctime ASC LIMIT ?"
	sqlPageSelSubs       = "SELECT " + fields + " FROM comment WHERE oid=? AND root=? ORDER BY ctime ASC, id ASC LIMIT ?,?"
	sqlBatchCountSubs    = "SELECT root, COUNT(id) cnt FROM comment WHERE root!=0 AND root IN (%s) GROUP BY root"
	sqlCountSubs         = "SELECT COUNT(*) FROM comment WHERE oid=? AND root=?"
	sqlBatchSelAll       = "SELECT " + fields + " FROM comment WHERE id IN (%s)"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO comment(%s) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", fieldsWithoutId)
)

func (r *CommentDao) FindByIdForUpdate(ctx context.Context, id int64) (*Comment, error) {
	var res Comment
	err := r.db.QueryRowCtx(ctx, &res, sqlSel4Ud, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *CommentDao) FindRootParent(ctx context.Context, id int64) (*RootParent, error) {
	var res RootParent
	err := r.db.QueryRowCtx(ctx, &res, sqlSelRootParentById, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func (r *CommentDao) FindById(ctx context.Context, id int64) (*Comment, error) {
	// res, err := r.cacheV2.GetOrFetch(ctx, fmtCommentCacheKey(id),
	// 	func(ctx context.Context) (*Comment, time.Duration, error) {
	// 		var res Comment
	// 		err := r.db.QueryRowCtx(ctx, &res, sqlSel, id)
	// 		if err != nil {
	// 			return nil, 0, xsql.ConvertError(err)
	// 		}

	// 		return &res, xtime.DayJitter(time.Minute * 30), nil
	// 	},
	// 	xcachev2.WithSerializer(xcachev2.MsgpackSer),
	// )
	// if err != nil {
	// 	return nil, err
	// }

	// return res, nil
	var res Comment
	err := r.db.QueryRowCtx(ctx, &res, sqlSel, id)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &res, nil
}

func fmtCommentCacheKey(id int64) string {
	return commentIdCacheKeyTmpl + strconv.FormatInt(id, 10)
}

func (r *CommentDao) BatchFindById(ctx context.Context, ids []int64) ([]*Comment, error) {
	if len(ids) == 0 {
		return []*Comment{}, nil
	}

	// 	keys, keysMapping := xcachev2.KeysAndMap(ids, fmtCommentCacheKey)
	// 	result, err := r.cacheV2.MGetOrFetch(ctx,
	// 		keys,
	// 		func(ctx context.Context, keys []string) (map[string]*Comment, error) {
	// 			dbIds := xcachev2.RangeKeys(keys, keysMapping)
	// 			dbIds = xslice.Uniq(dbIds)

	// 			dbResult := make([]*Comment, 0)
	// 			sql := fmt.Sprintf(sqlBatchSelAll, xslice.JoinInts(dbIds))
	// 			err := r.db.QueryRowsCtx(ctx, &dbResult, sql)
	// 			if err != nil {
	// 				return nil, xerror.Wrapf(xsql.ConvertError(err), "comment dao query by ids failed")
	// 			}

	// 			ret := xslice.MakeMap(dbResult, func(v *Comment) string {
	// 				return fmtCommentCacheKey(v.Id)
	// 			})

	// 			return ret, nil
	// 		},
	// 		xcachev2.WithTTL(xtime.DayJitter(time.Minute*30)),
	// 		xcachev2.WithSerializer(xcachev2.MsgpackSer),
	// 	)

	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return xmap.Values(result), nil

	dbResult := make([]*Comment, 0)
	sql := fmt.Sprintf(sqlBatchSelAll, xslice.JoinInts(ids))
	err := r.db.QueryRowsCtx(ctx, &dbResult, sql)
	if err != nil {
		return nil, xerror.Wrapf(xsql.ConvertError(err), "comment dao query by ids failed")
	}

	return dbResult, nil
}

func (r *CommentDao) Insert(ctx context.Context, model *Comment) (int64, error) {
	if model.Ctime <= 0 {
		model.Ctime = time.Now().Unix()
	}

	if model.Mtime <= 0 {
		model.Mtime = model.Ctime
	}

	res, err := r.db.ExecCtx(ctx, sqlInsert,
		model.Oid,
		model.Type,
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
	return int64(newId), nil
}

func (r *CommentDao) DeleteById(ctx context.Context, id int64) error {
	_, err := r.db.ExecCtx(ctx, sqlDelById, id)
	return xsql.ConvertError(err)
}

func (r *CommentDao) DeleteByRoot(ctx context.Context, rootId int64) error {
	_, err := r.db.ExecCtx(ctx, sqlDelByRoot, rootId)
	return xsql.ConvertError(err)
}

func (r *CommentDao) findByOId(ctx context.Context, oid int64, lock bool) ([]*Comment, error) {
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

func (r *CommentDao) FindByOid(ctx context.Context, oid int64, lock bool) ([]*Comment, error) {
	return r.findByOId(ctx, oid, lock)
}

func (r *CommentDao) findByRootId(ctx context.Context, rootId int64, lock bool) ([]*Comment, error) {
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

func (r *CommentDao) FindByRootId(ctx context.Context, rootId int64, lock bool) ([]*Comment, error) {
	return r.findByRootId(ctx, rootId, lock)
}

func (r *CommentDao) FindByParentId(ctx context.Context, rootId int64, lock bool) ([]*Comment, error) {
	return r.findByRootId(ctx, rootId, lock)
}

func (r *CommentDao) updateCount(ctx context.Context, query string, id int64) error {
	_, err := r.db.ExecCtx(ctx, query, time.Now().Unix(), id)
	return xsql.ConvertError(err)
}

func (r *CommentDao) AddLike(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlIncLike, id)
}

func (r *CommentDao) AddReport(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlIncReport, id)
}

func (r *CommentDao) AddDisLike(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlIncDislike, id)
}

func (r *CommentDao) SubLike(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlDecLike, id)
}

func (r *CommentDao) SubReport(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlDecReport, id)
}

func (r *CommentDao) SubDisLike(ctx context.Context, id int64) error {
	return r.updateCount(ctx, sqlDecDislike, id)
}

func (r *CommentDao) setPin(ctx context.Context, oid, id int64, pin bool) error {
	// 移除缓存
	defer func() {
		if _, err := r.pinnedCache.Del(ctx, getPinnedCommentCacheKey(oid)); err != nil {
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
func (r *CommentDao) SetPin(ctx context.Context, oid, id int64) error {
	return r.setPin(ctx, oid, id, true)
}

// 取消置顶
func (r *CommentDao) SetUnPin(ctx context.Context, oid, id int64) error {
	return r.setPin(ctx, oid, id, false)
}

// 获取主评论
func (r *CommentDao) GetRoots(ctx context.Context, oid, cursor int64, want int) ([]*Comment, error) {
	var res = make([]*Comment, 0, want)
	hasCursor := ""
	var args []any
	if cursor > 0 {
		hasCursor = "id<? AND"
		args = []any{cursor, oid, want}
	} else {
		args = []any{oid, want}
	}
	// 按照创建时间从大到小排序 较新的评论在上面
	err := r.db.QueryRowsCtx(ctx, &res, fmt.Sprintf(sqlSelRoots, hasCursor), args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

// 获取主评论下的子评论的数量
// rootId -> cnt
func (r *CommentDao) BatchCountSubs(ctx context.Context, rootIds []int64) (map[int64]int64, error) {
	var res = make(map[int64]int64, 0)
	if len(rootIds) == 0 {
		return res, nil
	}

	batchRes := make([]RootCnt, 0)

	err := r.db.QueryRowsCtx(ctx, &batchRes, fmt.Sprintf(sqlBatchCountSubs, slices.JoinInts(rootIds)))
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	for _, r := range batchRes {
		res[r.Root] = r.Cnt
	}

	return res, nil
}

// 获取子评论
func (r *CommentDao) GetSubReplies(ctx context.Context, oid, root, cursor int64, want int) ([]*Comment, error) {
	var res = make([]*Comment, 0, want)
	err := r.db.QueryRowsCtx(ctx, &res, sqlSelSubs, cursor, oid, root, want)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

// 获取子评论数量
func (r *CommentDao) CountSubs(ctx context.Context, oid, root int64) (int64, error) {
	if root == 0 {
		return 0, nil
	}

	var res int64
	err := r.db.QueryRowCtx(ctx, &res, sqlCountSubs, oid, root)

	return res, xsql.ConvertError(err)
}

// 分页获取子评论
// page从1开始
func (r *CommentDao) PageGetSubs(ctx context.Context, oid, root int64, page, cnt int) ([]*Comment, error) {
	if page <= 0 || cnt <= 0 {
		return []*Comment{}, nil
	}

	var res = make([]*Comment, 0, cnt)
	err := r.db.QueryRowsCtx(ctx, &res, sqlPageSelSubs, oid, root, (page-1)*cnt, cnt)

	return res, xsql.ConvertError(err)
}

// 置顶
func (r *CommentDao) DoPin(ctx context.Context, oid, rid int64) error {
	_, err := r.db.ExecCtx(ctx, sqlDoPin, time.Now().Unix(), oid, rid)
	defer func() {
		if _, err := r.pinnedCache.Del(ctx, getPinnedCommentCacheKey(oid)); err != nil {
			xlog.Msg("pinned cache del pinned failed").Extra("oid", oid).Errorx(ctx)
		}
	}()

	return xsql.ConvertError(err)
}

// 拿出置顶评论
func (r *CommentDao) GetPinned(ctx context.Context, oid int64) (*Comment, error) {
	model, err := r.pinnedCache.Get(ctx, getPinnedCommentCacheKey(oid), xcache.WithGetFallback(
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
func (r *CommentDao) CountByOid(ctx context.Context, oid int64) (int64, error) {
	cnt, err := r.countCache.Get(ctx, getCommentCountCacheKey(oid),
		xcache.WithGetFallback(
			func(ctx context.Context) (int64, int, error) {
				var cnt int64
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

func (r *CommentDao) IncrCommentCount(ctx context.Context, oid int64) error {
	return r.cache.IncrCommentCountWhenExist(ctx, oid, 1)
}

func (r *CommentDao) DecrCommentCount(ctx context.Context, oid int64) error {
	return r.cache.DecrCommentCountWhenExist(ctx, oid, 1)
}

func (r *CommentDao) BatchCountByOid(ctx context.Context, oids []int64) (map[int64]int64, error) {
	keys := make([]string, 0, len(oids))
	oidKeysMap := make(map[string]int64, len(oids))
	for _, oid := range oids {
		key := getCommentCountCacheKey(oid)
		keys = append(keys, key)
		oidKeysMap[key] = oid
	}

	queriedResult, err := r.countCache.MGet(ctx, keys,
		r.countCache.WithMGetFallbackSec(xtime.DayJitterSec(time.Minute*30)),
		r.countCache.WithMGetBgSet(true),
		r.countCache.WithMGetFallback(
			func(ctx context.Context, missingKeys []string) (t map[string]int64, err error) {
				if len(missingKeys) == 0 {
					return
				}

				var (
					ret         []OidCnt
					missingOids []int64 = make([]int64, 0, len(missingKeys))
				)

				for _, k := range missingKeys {
					missingOids = append(missingOids, oidKeysMap[k])
				}

				query := fmt.Sprintf(sqlBatchCountByO, slices.JoinInts(missingOids))
				err = r.db.QueryRowsCtx(ctx, &ret, query)
				if err != nil {
					return nil, xerror.Wrap(xsql.ConvertError(err))
				}

				nr := make(map[string]int64, len(ret))
				for _, oidCnt := range ret {
					nr[getCommentCountCacheKey(oidCnt.Oid)] = oidCnt.Cnt
				}

				xlog.Msg(fmt.Sprintf("%v", nr)).Infox(ctx)

				return nr, nil
			},
		),
	)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]int64, len(queriedResult))
	for _, oid := range oids {
		key := getCommentCountCacheKey(oid)
		result[oid] = queriedResult[key]
	}

	return result, nil
}

func (r *CommentDao) CountByOidUid(ctx context.Context, oid int64, uid int64) (int64, error) {
	var cnt int64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByOU, oid, uid)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *CommentDao) CountGroupByOid(ctx context.Context) (map[int64]int64, error) {
	var res []struct {
		Oid int64 `db:"oid"`
		Cnt int64 `db:"cnt"`
	}
	err := r.db.QueryRowsCtx(ctx, &res, sqlCountGbO)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	ret := make(map[int64]int64, len(res))
	for _, item := range res {
		ret[item.Oid] = item.Cnt
	}

	if err := r.cache.BatchSetCommentCount(ctx, ret); err != nil {
		xlog.Msg("comment dao cache batch set comment count failed").Err(err).Errorx(ctx)
	}

	return ret, nil
}

func (r *CommentDao) CountGroupByOidLimit(ctx context.Context, offset, limit int64) (map[int64]int64, error) {
	var res []struct {
		Oid int64 `db:"oid"`
		Cnt int64 `db:"cnt"`
	}
	err := r.db.QueryRowsCtx(ctx, &res, sqlCountGbOLimit, offset, limit)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	ret := make(map[int64]int64, len(res))
	for _, item := range res {
		ret[item.Oid] = item.Cnt
	}

	if err := r.cache.BatchSetCommentCount(ctx, ret); err != nil {
		xlog.Msg("comment dao cache batch set comment count failed").Err(err).Errorx(ctx)
	}

	return ret, nil
}

// uid -> []oids
func (r *CommentDao) FindByUidsOids(ctx context.Context, uidOids map[int64][]int64) ([]UidOid, error) {
	var batchRes []UidOid
	// 分批操作
	err := maps.BatchExec(uidOids, 200, func(target map[int64][]int64) error {
		uids, oids := maps.All(target)
		var allOids []int64 = oids[0]
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
