package record

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xcache/functions"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xtime"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	//go:embed lua/functions.lua
	luaFunctionCodes string
)

const (
	defaultCounterListMaxMember  = 5000
	defultCounterListEvictNumber = 100

	counterListKeyTmpl   = "counter:record:zset:b%d:u%d"    // counter:record:zset:b{bizcode}:u{uid}
	counterRecordKeyTmpl = "counter:record:all:b%d:u%d:o%d" // counter:record:all:b{bizcode}:u{uid}:o{oid}
)

type Cache struct {
	c *redis.Redis

	keyPrefix              string
	maxtCounterListMembers int
	counterListEvictNumber int
}

// cache structure: sorted set + string
//
// 1. counter list -> sorted set of {member:oid, score:unix_time}, only ActDo action is cached
// 2. counter record -> hash
func NewCache(c *redis.Redis, opts ...CacheOption) *Cache {
	opt := cacheOption{}
	for _, o := range opts {
		o(&opt)
	}
	cache := &Cache{
		c:                      c,
		maxtCounterListMembers: defaultCounterListMaxMember,
		counterListEvictNumber: defultCounterListEvictNumber,
		keyPrefix:              opt.KeyPrefix,
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

func (c *Cache) SetCounterListMaxMember(limit int) {
	c.maxtCounterListMembers = limit
}

// the number of members to be evicted when counter list is overflow
func (c *Cache) SetCounterListEvitNumber(number int) {
	c.counterListEvictNumber = number
}

type CacheKey struct {
	BizCode int32
	Uid     int64
	Oid     int64
}

type CacheRecord struct {
	Act   int8
	Oid   int64
	Mtime int64
}

func (c *Cache) getCounterListCacheKey(bizCode int32, uid int64) string {
	k := fmt.Sprintf(counterListKeyTmpl, bizCode, uid)
	if c.keyPrefix != "" {
		return c.keyPrefix + "_" + k
	}

	return k
}

func getCounterListCacheObjValue(oid int64) string {
	return xconv.FormatInt(oid)
}

func (c *Cache) getCounterRecordCacheKey(bizCode int32, uid, oid int64) string {
	k := fmt.Sprintf(counterRecordKeyTmpl, bizCode, uid, oid)
	if c.keyPrefix != "" {
		return c.keyPrefix + "_" + k
	}
	return k
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

func (c *Cache) batchGetCounterRecordCacheKeys(keys []CacheKey) []string {
	cacheKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		cacheKeys = append(cacheKeys, c.getCounterRecordCacheKey(k.BizCode, k.Uid, k.Oid))
	}
	return cacheKeys
}

func (c *Cache) CounterListSize(ctx context.Context, bizCode int32, uid int64) (int, error) {
	key := c.getCounterListCacheKey(bizCode, uid)
	size, err := c.c.ZcardCtx(ctx, key)
	return size, xerror.Wrapf(err, "zcard failed")
}

// add a record to sorted set
func (c *Cache) CounterListAdd(ctx context.Context, bizCode int32, uid int64, record *CacheRecord) error {
	if record.Act != ActDo {
		return nil
	}

	key := c.getCounterListCacheKey(bizCode, uid)
	_, err := c.c.ZaddCtx(ctx, key, record.Mtime, getCounterListCacheObjValue(record.Oid))
	if err != nil {
		return xerror.Wrapf(err, "zadd failed")
	}

	return nil
}

// batch add, bizCode and uid will be taken as cache key
func (c *Cache) CounterListBatchAdd(ctx context.Context, bizCode int32, uid int64, records []*CacheRecord) error {
	targets := getBatchCacheRecordTargets(records)
	if len(targets) == 0 {
		return nil
	}

	key := c.getCounterListCacheKey(bizCode, uid)
	pairs := make([]redis.Pair, 0, len(targets))
	for _, t := range targets {
		pairs = append(pairs, redis.Pair{
			Key:   getCounterListCacheObjValue(t.Oid),
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
// CounterListSizeLimitBatchAdd operation will be performed by a lua script.
func (c *Cache) CounterListSizeLimitBatchAdd(ctx context.Context, bizCode int32, uid int64, records []*CacheRecord) error {
	key := c.getCounterListCacheKey(bizCode, uid)
	args := make([]any, 0, len(records)*2)

	targets := getBatchCacheRecordTargets(records)
	if len(targets) == 0 {
		return nil
	}

	args = append(args, c.maxtCounterListMembers, c.counterListEvictNumber)
	for _, t := range targets {
		// score comes first, then is member key
		score := t.Mtime
		member := getCounterListCacheObjValue(t.Oid)
		args = append(args, score, member)
	}

	_, err := functions.FunctionCall(ctx, c.c, "counter_sizelimit_batchadd", []string{key}, args...)
	if err != nil {
		return xerror.Wrapf(err, "script run sizelimit_batch_add failed")
	}

	return nil
}

// check if a record is in sorted set
func (c *Cache) CounterListExistsOid(ctx context.Context, bizCode int32, uid int64, oid int64) (bool, error) {
	key := c.getCounterListCacheKey(bizCode, uid)
	score, err := c.c.ZscoreCtx(ctx, key, getCounterListCacheObjValue(oid))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, xerror.Wrapf(err, "zscore failed")
	}

	return score > 0, nil
}

func (c *Cache) CounterListBatchExistsOid(ctx context.Context, bizCode int32, uid int64, oids ...int64) (map[int64]bool, error) {
	key := c.getCounterListCacheKey(bizCode, uid)
	pipe, err := c.c.TxPipeline()
	if err != nil {
		return nil, xerror.Wrapf(err, "get pipe failed")
	}

	vals := make([]string, 0, len(oids))
	for _, o := range oids {
		vals = append(vals, getCounterListCacheObjValue(o))
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

func (c *Cache) CounterListRemoveOid(ctx context.Context, bizCode int32, uid int64, oid int64) error {
	key := c.getCounterListCacheKey(bizCode, uid)
	_, err := c.c.ZremCtx(ctx, key, getCounterListCacheObjValue(oid))
	if err != nil {
		return xerror.Wrapf(err, "zrem failed")
	}

	return nil
}

func (c *Cache) CounterListBatchRemoveOids(ctx context.Context, bizCode int32, uid int64, oids ...int64) error {
	key := c.getCounterListCacheKey(bizCode, uid)
	vals := make([]any, 0, len(oids))
	for _, oid := range oids {
		vals = append(vals, getCounterListCacheObjValue(oid))
	}

	_, err := c.c.ZremCtx(ctx, key, vals...)
	if err != nil {
		return xerror.Wrapf(err, "batch zrem failed")
	}

	return nil
}

func (c *Cache) AddRecord(ctx context.Context, record *Record, opts ...CacheOption) error {
	opt := cacheOption{
		Expire: xtime.NDayJitter(2, time.Minute*15),
	}

	for _, o := range opts {
		o(&opt)
	}

	key := c.getCounterRecordCacheKey(record.BizCode, record.Uid, record.Oid)
	var data = map[string]any{}
	err := mapstructure.Decode(record, &data)
	if err != nil {
		return xerror.Wrapf(err, "mapstructure record failed")
	}

	err = c.c.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		p.HMSet(ctx, key, xmap.KVs(data)...)
		p.Expire(ctx, key, opt.Expire)
		return nil
	})

	return xerror.Wrapf(err, "hmset pipeline failed")
}

func (c *Cache) BatchAddRecord(ctx context.Context, records []*Record, opts ...CacheOption) error {
	if len(records) == 0 {
		return nil
	}

	opt := cacheOption{
		Expire: xtime.NDayJitter(2, time.Minute*15),
	}

	for _, o := range opts {
		o(&opt)
	}

	err := c.c.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		for _, r := range records {
			key := c.getCounterRecordCacheKey(r.BizCode, r.Uid, r.Oid)
			var data map[string]any
			err := mapstructure.Decode(r, &data)
			if err == nil {
				p.HMSet(ctx, key, xmap.KVs(data)...)
				p.Expire(ctx, key, opt.Expire)
			}
		}

		return nil
	})

	if err != nil {
		xlog.Msg("batch set expire failed").Err(err).Errorx(ctx)
	}

	return nil
}

func (c *Cache) GetRecord(ctx context.Context, bizCode int32, uid, oid int64) (*Record, error) {
	var r Record
	err := xcache.HGetAllWithScan(ctx, c.c, c.getCounterRecordCacheKey(bizCode, uid, oid), &r)
	if err != nil {
		return nil, xerror.Wrapf(err, "get failed")
	}

	return &r, nil
}

func (c *Cache) BatchGetRecord(ctx context.Context, keys []CacheKey) (map[CacheKey]*Record, error) {
	pipe, err := c.c.TxPipeline()
	if err != nil {
		return nil, xerror.Wrapf(err, "tx pipeline failed")
	}

	cmds := make([]*goredis.MapStringStringCmd, 0)
	for _, key := range keys {
		cacheKey := c.getCounterRecordCacheKey(key.BizCode, key.Uid, key.Oid)
		cmd := pipe.HGetAll(ctx, cacheKey)
		cmds = append(cmds, cmd)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "pipe exec failed")
	}

	result := make(map[CacheKey]*Record, len(keys))
	for idx, cmd := range cmds {
		if val, err := cmd.Result(); err == nil && len(val) > 0 {
			var r Record
			if err = cmd.Scan(&r); err == nil {
				// check if scanned record is valid
				if r.BizCode == keys[idx].BizCode &&
					r.Uid == keys[idx].Uid &&
					r.Oid == keys[idx].Oid &&
					r.Mtime > 0 {
					result[keys[idx]] = &r
				}
			}
		}
	}

	return result, nil
}

func (c *Cache) DelRecord(ctx context.Context, bizCode int32, uid, oid int64) error {
	_, err := c.c.DelCtx(ctx, c.getCounterRecordCacheKey(bizCode, uid, oid))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		return xerror.Wrapf(err, "del failed")
	}

	return nil
}

func (c *Cache) BatchDelCount(ctx context.Context, keys []CacheKey) error {
	cacheKeys := c.batchGetCounterRecordCacheKeys(keys)
	_, err := c.c.DelCtx(ctx, cacheKeys...)
	return xerror.Wrapf(err, "batch del failed")
}

func (c *Cache) CheckHasCountedRecord(ctx context.Context, bizCode int32, uid, oid int64) (bool, error) {
	counterListKey := c.getCounterListCacheKey(bizCode, uid)
	counterRecordKey := c.getCounterRecordCacheKey(bizCode, uid, oid)
	counterListMember := getCounterListCacheObjValue(oid)

	cmd, err := functions.FunctionCall(ctx, c.c,
		"counter_check_actdo_record",
		[]string{counterListKey, counterRecordKey}, // KEYS
		counterListMember, ActDo, 0) // ARGS
	if err != nil {
		return false, xerror.Wrapf(err, "script run counter_check_actdo_record failed")
	}

	ok, err := cmd.Bool()
	if err != nil {
		return false, xerror.Wrapf(err, "cmd can convert to bool")
	}

	return ok, nil
}
