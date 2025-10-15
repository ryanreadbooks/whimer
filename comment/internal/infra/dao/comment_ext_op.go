package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	xcachev2 "github.com/ryanreadbooks/whimer/misc/xcache/v2"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CommentExtDao struct {
	db    *xsql.DB
	cache *xcachev2.Cache[*CommentExt]
}

func NewCommentExtDao(db *xsql.DB, cache *redis.Redis) *CommentExtDao {
	return &CommentExtDao{
		db:    db,
		cache: xcachev2.New[*CommentExt](cache),
	}
}

// sqls here
const (
	extFields = "comment_id,at_users,ctime,mtime"

	sqlSelExtAll = "SELECT " + extFields + " FROM comment_ext WHERE comment_id=?"
	sqlUpsertExt = "INSERT INTO comment_ext(" + extFields + ") VALUES(?,?,?,?)" +
		" ON DUPLICATE KEY UPDATE at_users=VALUES(at_users), mtime=VALUES(mtime)"

	sqlDelExt         = "DELETE FROM comment_ext WHERE comment_id=? LIMIT 1"
	sqlBatchSelExtAll = "SELECT " + extFields + " FROM comment_ext WHERE comment_id IN (%s)"
)

const (
	commentExtCacheKeyTmpl = "comment:ext:" // comment:ext:commentId
)

func fmtCommentExtCacheKey(cmtId int64) string {
	return commentExtCacheKeyTmpl + strconv.FormatInt(cmtId, 10)
}

func (d *CommentExtDao) Get(ctx context.Context, cmtId int64) (*CommentExt, error) {
	ext, err := d.cache.GetOrFetch(ctx, fmtCommentExtCacheKey(cmtId),
		func(ctx context.Context) (*CommentExt, time.Duration, error) {
			var ext CommentExt
			err := d.db.QueryRowCtx(ctx, &ext, sqlSelExtAll, cmtId)
			if err != nil {
				return nil, 0, xsql.ConvertError(err)
			}

			return &ext, xtime.WeekJitter(time.Minute * 15), nil
		},
		xcachev2.WithSerializer(xcachev2.MsgPackSerializer{}))
	if err != nil {
		return nil, err
	}

	return ext, nil
}

// 返回的切片不按照入参的cmtIds顺序
func (d *CommentExtDao) BatchGet(ctx context.Context, cmtIds []int64) ([]*CommentExt, error) {
	if len(cmtIds) == 0 {
		return make([]*CommentExt, 0), nil
	}

	keys, keysMapping := xcachev2.KeysAndMap(cmtIds, fmtCommentExtCacheKey)
	result, err := d.cache.MGetOrFetch(ctx, keys,
		func(ctx context.Context, keys []string) (map[string]*CommentExt, error) {
			dbIds := xcachev2.RangeKeys(keys, keysMapping)
			dbIds = xslice.Uniq(dbIds)

			dbResult := make([]*CommentExt, 0)
			sql := fmt.Sprintf(sqlBatchSelExtAll, xslice.JoinInts(dbIds))
			err := d.db.QueryRowsCtx(ctx, &dbResult, sql)
			if err != nil {
				return nil, xerror.Wrapf(xsql.ConvertError(err), "comment ext dao query by ids failed")
			}

			ret := xslice.MakeMap(dbResult, func(v *CommentExt) string {
				return fmtCommentExtCacheKey(v.CommentId)
			})

			return ret, nil
		},
		xcachev2.WithTTL(xtime.WeekJitter(time.Minute*15)),
		xcachev2.WithSerializer(xcachev2.MsgPackSerializer{}),
	)

	if err != nil {
		return nil, err
	}

	return xmap.Values(result), nil
}

func (d *CommentExtDao) Upsert(ctx context.Context, ext *CommentExt) error {
	now := time.Now().Unix()
	if ext.Ctime == 0 {
		ext.Ctime = now
	}

	ext.Mtime = now

	_, err := d.db.ExecCtx(ctx, sqlUpsertExt, ext.CommentId, ext.AtUsers, ext.Ctime, ext.Mtime)
	defer d.cache.Del(ctx, fmtCommentExtCacheKey(ext.CommentId))
	return xsql.ConvertError(err)
}

func (d *CommentExtDao) Delete(ctx context.Context, cmtId int64) error {
	_, err := d.db.ExecCtx(ctx, sqlDelExt, cmtId)
	defer d.cache.Del(ctx, fmtCommentExtCacheKey(cmtId))
	return xsql.ConvertError(err)
}
