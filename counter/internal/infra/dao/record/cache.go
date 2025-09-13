package record

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xcache/functions"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	//go:embed lua/functions.lua
	luaFunctionCodes string
)

const (
	defaultMaxMemberPerKey     = 5000
	defultEvitNumberOnOverflow = 100
	keyTmpl                    = "counter:record:user:%d:%d" // counter:record:user:bizcode:uid
)

type Cache struct {
	c *redis.Redis

	maxMemberPerKey      int
	evitNumberOnOverflow int
}

// cache structure: sorted set
//
// key -> sorted set of {member:oid, score:unix_time}, only ActDo action is cached
func NewCache(c *redis.Redis) *Cache {
	cache := &Cache{
		c:                    c,
		maxMemberPerKey:      defaultMaxMemberPerKey,
		evitNumberOnOverflow: defultEvitNumberOnOverflow,
	}

	return cache
}

// init libcounter functions
func (c *Cache) InitFunction(ctx context.Context) error {
	err := functions.FunctionLoadReplace(ctx, c.c, luaFunctionCodes)
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) SetMaxMemberPerKey(limit int) {
	c.maxMemberPerKey = limit
}

func (c *Cache) SetEvitNumberOnOverflow(number int) {
	c.evitNumberOnOverflow = number
}

type CacheKey struct {
	BizCode int32
	Uid     int64
}

func getCacheKey(bizCode int32, uid int64) string {
	return fmt.Sprintf(keyTmpl, bizCode, uid)
}

func getCacheObjValue(oid int64) string {
	return xconv.FormatInt(oid)
}

type CacheRecord struct {
	Act   int8
	Oid   int64
	Mtime int64
}

func (c *Cache) Size(ctx context.Context, bizCode int32, uid int64) (int, error) {
	key := getCacheKey(bizCode, uid)
	size, err := c.c.ZcardCtx(ctx, key)
	return size, xerror.Wrapf(err, "zcard failed")
}

// add a record to sorted set
func (c *Cache) Add(ctx context.Context, bizCode int32, uid int64, record *CacheRecord) error {
	if record.Act != ActDo {
		return nil
	}

	key := getCacheKey(bizCode, uid)
	_, err := c.c.ZaddCtx(ctx, key, record.Mtime, getCacheObjValue(record.Oid))
	if err != nil {
		return xerror.Wrapf(err, "zadd failed")
	}

	return nil
}

func getBatchCacheRecordTargets(records []*CacheRecord) []*CacheRecord {
	targets := make([]*CacheRecord, 0, len(records))
	for _, r := range records {
		if r.Act == ActDo {
			targets = append(targets, r)
		}
	}

	return targets
}

// batch add, bizCode and uid will be taken as cache key
func (c *Cache) BatchAdd(ctx context.Context, bizCode int32, uid int64, records []*CacheRecord) error {
	targets := getBatchCacheRecordTargets(records)
	if len(targets) == 0 {
		return nil
	}

	key := getCacheKey(bizCode, uid)
	pairs := make([]redis.Pair, 0, len(targets))
	for _, t := range targets {
		pairs = append(pairs, redis.Pair{
			Key:   getCacheObjValue(t.Oid),
			Score: t.Mtime,
		})
	}

	_, err := c.c.ZaddsCtx(ctx, key, pairs...)
	if err != nil {
		return xerror.Wrapf(err, "batch zadd failed")
	}

	return nil
}

// Strictly restraining the number of member in a sorted set key.
//
// SizeLimitBatchAdd operation will be performed by a lua script.
func (c *Cache) SizeLimitBatchAdd(ctx context.Context, bizCode int32, uid int64, records []*CacheRecord) error {
	key := getCacheKey(bizCode, uid)
	args := make([]any, 0, len(records)*2)

	targets := getBatchCacheRecordTargets(records)
	if len(targets) == 0 {
		return nil
	}

	args = append(args, c.maxMemberPerKey, c.evitNumberOnOverflow)
	for _, t := range targets {
		// score comes first, then is member key
		score := t.Mtime
		member := getCacheObjValue(t.Oid)
		args = append(args, score, member)
	}

	err := functions.FunctionCall(ctx, c.c, "counter_sizelimit_batchadd", []string{key}, args...)
	if err != nil {
		return xerror.Wrapf(err, "script run sizelimit_batch_add failed")
	}

	return nil
}

// check if a record is in sorted set
func (c *Cache) ExistsOid(ctx context.Context, bizCode int32, uid int64, oid int64) (bool, error) {
	key := getCacheKey(bizCode, uid)
	score, err := c.c.ZscoreCtx(ctx, key, getCacheObjValue(oid))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, xerror.Wrapf(err, "zscore failed")
	}

	return score > 0, nil
}

func (c *Cache) BatchExistsOid(ctx context.Context, bizCode int32, uid int64, oids ...int64) (map[int64]bool, error) {
	key := getCacheKey(bizCode, uid)
	pipe, err := c.c.TxPipeline()
	if err != nil {
		return nil, xerror.Wrapf(err, "get pipe failed")
	}

	vals := make([]string, 0, len(oids))
	for _, o := range oids {
		vals = append(vals, getCacheObjValue(o))
	}

	resCmd := pipe.ZMScore(ctx, key, vals...)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "pipe exec failed")
	}

	scores, err := resCmd.Result()
	if err != nil {
		return nil, xerror.Wrapf(err, "pipe exec float slice cmd result err")
	}

	existence := make(map[int64]bool, len(scores))

	for idx, score := range scores {
		if score != 0 {
			existence[oids[idx]] = true
		}
	}

	return existence, nil
}

func (c *Cache) RemoveOid(ctx context.Context, bizCode int32, uid int64, oid int64) error {
	key := getCacheKey(bizCode, uid)
	_, err := c.c.ZremCtx(ctx, key, getCacheObjValue(oid))
	if err != nil {
		return xerror.Wrapf(err, "zrem failed")
	}

	return nil
}

func (c *Cache) BatchRemoveOids(ctx context.Context, bizCode int32, uid int64, oids ...int64) error {
	key := getCacheKey(bizCode, uid)
	vals := make([]any, 0, len(oids))
	for _, oid := range oids {
		vals = append(vals, getCacheObjValue(oid))
	}

	_, err := c.c.ZremCtx(ctx, key, vals...)
	if err != nil {
		return xerror.Wrapf(err, "batch zrem failed")
	}

	return nil
}
