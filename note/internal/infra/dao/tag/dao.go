package tag

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TagDao struct {
	db       *xsql.DB
	cache    *redis.Redis
	tagCache *xcache.Cache[*Tag]
}

func NewTagDao(db *xsql.DB, cache *redis.Redis) *TagDao {
	return &TagDao{
		db:       db,
		cache:    cache,
		tagCache: xcache.New[*Tag](cache),
	}
}

func (d *TagDao) Create(ctx context.Context, tag *Tag) (int64, error) {
	const sqlInsert = "INSERT INTO tag(name,ctime) VALUES(?,?)"
	if tag.Ctime == 0 {
		tag.Ctime = time.Now().Unix()
	}

	res, err := d.db.ExecCtx(ctx, sqlInsert, tag.Name, tag.Ctime)
	if err != nil {
		err = xsql.ConvertError(err)
		return 0, xerror.Wrap(err)
	}

	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return newId, nil
}

func (d *TagDao) rawFind(ctx context.Context, name string) (*Tag, error) {
	const sqlFind = "SELECT id,name,ctime FROM tag WHERE name=?"
	var tag Tag
	err := d.db.QueryRowCtx(ctx, &tag, sqlFind, name)
	return &tag, err
}

const (
	tagCacheNameKey = "tag:name:%s"
	tagCacheIdKey   = "tag:id:%d"
)

func getTagCacheIdKey(id int64) string {
	return fmt.Sprintf(tagCacheIdKey, id)
}

func (d *TagDao) Find(ctx context.Context, name string) (*Tag, error) {
	res, err := d.tagCache.Get(ctx, fmt.Sprintf(tagCacheNameKey, name),
		xcache.WithGetFallback(func(ctx context.Context) (*Tag, int, error) {
			t, err := d.rawFind(ctx, name)
			return t, xtime.WeekJitterSec(time.Hour), xerror.Wrap(err)
		}))

	return res, err
}

func (d *TagDao) FindById(ctx context.Context, id int64) (*Tag, error) {
	res, err := d.tagCache.Get(ctx, getTagCacheIdKey(id),
		xcache.WithGetFallback(func(ctx context.Context) (*Tag, int, error) {
			const sqlFind = "SELECT id,name,ctime FROM tag WHERE id=?"
			var tag Tag
			err := d.db.QueryRowCtx(ctx, &tag, sqlFind, id)
			return &tag, xtime.WeekJitterSec(time.Hour), xerror.Wrap(xsql.ConvertError(err))
		}))

	return res, err
}

func (d *TagDao) BatchGetById(ctx context.Context, ids []int64) ([]*Tag, error) {
	const sql = "SELECT id,name,ctime FROM tag WHERE id IN (%s)"

	keys := []string{}
	keysMap := make(map[string]int64, len(ids))
	for _, id := range ids {
		key := getTagCacheIdKey(id)
		keys = append(keys, key)
		keysMap[key] = id
	}

	// get from cache first, then we get the missing keys from db again
	result, err := d.tagCache.MGet(ctx, keys,
		xcache.WithMGetFallbackSec[*Tag](xtime.WeekJitterSec(time.Hour)),
		xcache.WithMGetFallback(func(ctx context.Context, missingKeys []string) (t map[string]*Tag, err error) {
			if len(missingKeys) == 0 {
				return
			}
			var (
				tags     []*Tag
				missings []int64
			)

			for _, k := range missingKeys {
				missings = append(missings, keysMap[k])
			}

			err = d.db.QueryRowsCtx(ctx, &tags, fmt.Sprintf(sql, xslice.JoinInts(missings)))
			if err != nil {
				return nil, xerror.Wrap(xsql.ConvertError(err))
			}
			return xslice.MakeMap(tags, func(v *Tag) string { return getTagCacheIdKey(v.Id) }), nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return xmap.Values(result), nil
}
