package dao

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	xcachev2 "github.com/ryanreadbooks/whimer/misc/xcache/v2"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type NamespaceDao struct {
	db    *xsql.DB
	cache *xcachev2.Cache[*NamespacePO]
}

func fmtNamespaceCacheKey(id string) string {
	return "conductor:namespace:id:" + id
}

func NewNamespaceDao(db *xsql.DB, r *redis.Redis) *NamespaceDao {
	return &NamespaceDao{
		db:    db,
		cache: xcachev2.New[*NamespacePO](r),
	}
}

func (d *NamespaceDao) Insert(ctx context.Context, po *NamespacePO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(namespacePOTableName)
	ib.Cols(namespacePOFields...)
	ib.Values(po.Values()...)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	// 清除缓存
	idStr := po.Id.String()
	d.cache.Del(ctx, fmtNamespaceCacheKey(idStr))

	return nil
}

func (d *NamespaceDao) GetById(ctx context.Context, id []byte) (*NamespacePO, error) {
	idStr := string(id)
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(namespacePOFields...)
	sb.From(namespacePOTableName)
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()

	// 先从缓存获取
	res, err := d.cache.GetOrFetch(ctx, fmtNamespaceCacheKey(idStr),
		func(ctx context.Context) (*NamespacePO, time.Duration, error) {
			var tmp NamespacePO
			err := d.db.QueryRowCtx(ctx, &tmp, sql, args...)
			if err != nil {
				return nil, 0, err
			}

			return &tmp, time.Duration(xtime.HourJitter(time.Hour * 1)), nil
		},
	)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

func (d *NamespaceDao) GetByName(ctx context.Context, name string) (*NamespacePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(namespacePOFields...)
	sb.From(namespacePOTableName)
	sb.Where(sb.Equal("name", name))

	sql, args := sb.Build()
	var po NamespacePO
	err := d.db.QueryRowCtx(ctx, &po, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	// 写入缓存
	idStr := po.Id.String()
	d.cache.Set(ctx, fmtNamespaceCacheKey(idStr), &po, xcachev2.WithTTL(time.Duration(xtime.HourJitter(time.Hour*1))))

	return &po, nil
}

func (d *NamespaceDao) UpdateById(ctx context.Context, id []byte, po *NamespacePO) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(namespacePOTableName)
	ub.Set(
		ub.Assign("name", po.Name),
	)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	// 清除缓存
	idStr := string(id)
	d.cache.Del(ctx, fmtNamespaceCacheKey(idStr))

	return nil
}

func (d *NamespaceDao) DeleteById(ctx context.Context, id []byte) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(namespacePOTableName)
	db.Where(db.Equal("id", id))

	sql, args := db.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	// 清除缓存
	idStr := string(id)
	d.cache.Del(ctx, fmtNamespaceCacheKey(idStr))

	return nil
}

// List 分页查询 namespace
func (d *NamespaceDao) List(ctx context.Context, offset, limit int) ([]*NamespacePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(namespacePOFields...)
	sb.From(namespacePOTableName)
	sb.OrderByAsc("id")
	sb.Offset(offset)
	sb.Limit(limit)

	sql, args := sb.Build()
	var pos []*NamespacePO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return pos, nil
}

// Count 统计 namespace 总数
func (d *NamespaceDao) Count(ctx context.Context) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("COUNT(*)")
	sb.From(namespacePOTableName)

	sql, args := sb.Build()
	var count int64
	err := d.db.QueryRowCtx(ctx, &count, sql, args...)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return count, nil
}
