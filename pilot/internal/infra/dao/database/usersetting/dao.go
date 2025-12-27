package usersetting

import (
	"context"
	"strconv"
	"time"

	"github.com/huandu/go-sqlbuilder"
	xcachev2 "github.com/ryanreadbooks/whimer/misc/xcache/v2"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Dao struct {
	db    *xsql.DB
	cache *xcachev2.Cache[*UserSettingPO]
}

func fmtUserSettingCacheKey(uid int64) string {
	return "pilot:usersetting:uid:" + strconv.FormatInt(uid, 10)
}

func NewDao(db *xsql.DB, r *redis.Redis) *Dao {
	return &Dao{
		db:    db,
		cache: xcachev2.New[*UserSettingPO](r),
	}
}

func (d *Dao) Upsert(ctx context.Context, po *UserSettingPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(userSettingPOTableName)
	ib.Cols(userSettingPOFields...)
	ib.Values(po.Values()...)
	extra := "ON DUPLICATE KEY UPDATE utime=VALUES(utime), flags=VALUES(flags)"
	ib.SQL(extra)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	d.cache.Del(ctx, fmtUserSettingCacheKey(po.Uid))

	return nil
}

func (d *Dao) GetByUid(ctx context.Context, uid int64, forUpdate bool) (*UserSettingPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(userSettingPOFields...)
	sb.From(userSettingPOTableName)
	sb.Where(sb.EQ("uid", uid))
	if forUpdate {
		sb.ForUpdate()
		d.cache.Del(ctx, fmtUserSettingCacheKey(uid))
	}

	var (
		po  UserSettingPO
		err error
	)
	sql, args := sb.Build()
	if forUpdate {
		err = d.db.QueryRowCtx(ctx, &po, sql, args...)
	} else {
		// normal get can go cache
		var tmp *UserSettingPO
		tmp, err = d.getByUidWithCache(ctx, uid, sql, args...)
		if err == nil {
			po = *tmp
		}
	}
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &po, nil
}

func (d *Dao) getByUidWithCache(ctx context.Context, uid int64, sql string, args ...any) (*UserSettingPO, error) {
	res, err := d.cache.GetOrFetch(ctx, fmtUserSettingCacheKey(uid),
		func(ctx context.Context) (*UserSettingPO, time.Duration, error) {
			var tmp UserSettingPO
			err := d.db.QueryRowCtx(ctx, &tmp, sql, args...)
			if err != nil {
				return nil, 0, err
			}

			return &tmp, xtime.WeekJitter(time.Minute * 30), nil
		},
	)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}
