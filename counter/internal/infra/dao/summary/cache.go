package summary

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	keyTmpl    = "counter:summary:%d:%d" // counter:summary:bizcode:oid
	defaultTTL = xtime.WeekSec
)

var (
	defaultTTLDuration = xtime.Week
)

var (
	ErrSummaryNotFound = fmt.Errorf("summary not found")
)

type Cache struct {
	c *redis.Redis
}

func NewCache(c *redis.Redis) *Cache {
	return &Cache{
		c: c,
	}
}

func getCacheKey(bizCode int32, oid int64) string {
	return fmt.Sprintf(keyTmpl, bizCode, oid)
}

func batchGetCacheKeys(keys []CacheKey) []string {
	cacheKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		cacheKeys = append(cacheKeys, getCacheKey(k.BizCode, k.Oid))
	}

	return cacheKeys
}

func (c *Cache) SetCount(ctx context.Context, bizCode int32, oid int64, count int64) error {
	err := c.c.SetexCtx(ctx, getCacheKey(bizCode, oid), xconv.FormatInt(count), defaultTTL)
	if err != nil {
		return xerror.Wrapf(err, "setex failed")
	}

	return nil
}

type CacheKey struct {
	BizCode int32
	Oid     int64
}

type CacheData struct {
	CacheKey
	Count int64
}

func (c *Cache) BatchSetCount(ctx context.Context, datas []CacheData) error {
	args := []any{}
	for _, data := range datas {
		key := getCacheKey(data.BizCode, data.Oid)
		args = append(args, key, data.Count)
	}

	_, err := c.c.MsetCtx(ctx, args...)
	if err != nil {
		return xerror.Wrapf(err, "mset failed")
	}

	err = c.c.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		for _, data := range datas {
			key := getCacheKey(data.BizCode, data.Oid)
			p.Expire(ctx, key, defaultTTLDuration)
		}
		return nil
	})
	if err != nil {
		xlog.Msg("batch set expire failed").Err(err).Errorx(ctx)
	}

	return nil
}

func (c *Cache) GetCount(ctx context.Context, bizCode int32, oid int64) (int64, error) {
	key := getCacheKey(bizCode, oid)
	res, err := c.c.GetCtx(ctx, key)
	if res == "" && err == nil {
		return 0, ErrSummaryNotFound
	}
	count, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return 0, ErrSummaryNotFound
	}

	return count, nil
}

func (c *Cache) BatchGetCount(ctx context.Context, keys []CacheKey) (map[CacheKey]int64, error) {
	cacheKeys := batchGetCacheKeys(keys)
	resp, err := c.c.MgetCtx(ctx, cacheKeys...)
	if err != nil {
		return nil, xerror.Wrapf(err, "mget failed")
	}

	result := make(map[CacheKey]int64, len(keys))
	for idx, r := range resp {
		count, err := strconv.ParseInt(r, 10, 64)
		if err == nil {
			key := keys[idx]
			result[key] = count
		}
	}

	return result, nil
}

func (c *Cache) DelCount(ctx context.Context, bizCode int32, oid int64) error {
	_, err := c.c.DelCtx(ctx, getCacheKey(bizCode, oid))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		return xerror.Wrapf(err, "del failed")
	}

	return nil
}

func (c *Cache) BatchDelCount(ctx context.Context, keys []CacheKey) error {
	cacheKeys := batchGetCacheKeys(keys)
	_, err := c.c.DelCtx(ctx, cacheKeys...)
	return xerror.Wrapf(err, "batch del failed")
}
