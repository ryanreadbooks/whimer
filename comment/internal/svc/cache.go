package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	hour = time.Hour
	day  = hour * 24
	week = 7 * day
)

// cache key templates
const (
	pinnedCmtKey = "comment:pinned:%d" // 置顶评论
	countCmtKey  = "comment:count:%d"  // 评论数量
)

func getPinnedCmtKey(oid uint64) string {
	return fmt.Sprintf(pinnedCmtKey, oid)
}

func getCountCmtKey(oid uint64) string {
	return fmt.Sprintf(countCmtKey, oid)
}

type Cache struct {
	rd       *redis.Redis
	incrOnce sync.Once
	incrSha  string
	decrOnce sync.Once
	decrSha  string
}

func NewCache(rd *redis.Redis) *Cache {
	return &Cache{
		rd: rd,
	}
}

func (c *Cache) GetPinned(ctx context.Context, oid uint64) (*comm.Model, error) {
	key := getPinnedCmtKey(oid)
	res, err := c.rd.GetCtx(ctx, key)
	if err != nil {
		return nil, err
	}

	var ret comm.Model
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *Cache) DelPinned(ctx context.Context, oid uint64) error {
	_, err := c.rd.DelCtx(ctx, getPinnedCmtKey(oid))
	return err
}

func (c *Cache) SetPinned(ctx context.Context, model *comm.Model) error {
	s, err := json.Marshal(model)
	if err != nil {
		return err
	}

	// err = c.rd.SetCtx(ctx, getPinnedCmtKey(model.Oid), utils.Bytes2String(s))
	// 默认过期时间设为7天 随机前后一个小时过期
	ttl := week + xtime.JitterDuration(2*hour)
	err = c.rd.SetexCtx(ctx, getPinnedCmtKey(model.Oid), utils.Bytes2String(s), int(ttl))

	return err
}

// 被评论对象的评论数量
func (c *Cache) GetReplyCount(ctx context.Context, oid uint64) (uint64, error) {
	cnt, err := c.rd.GetCtx(ctx, getCountCmtKey(oid))
	if err != nil {
		return 0, err
	}
	num, err := strconv.ParseUint(cnt, 10, 64)
	if err != nil {
		return 0, err
	}

	return num, nil
}

func (c *Cache) BatchGetReplyCount(ctx context.Context, oids []uint64) (map[uint64]uint64, error) {
	keys := make([]string, 0, len(oids))
	for _, oid := range oids {
		keys = append(keys, getCountCmtKey(oid))
	}
	cnts, err := c.rd.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, err
	}

	if len(oids) != len(cnts) {
		return nil, global.ErrInternal.Msg("缓存出错")
	}

	result := make(map[uint64]uint64)
	// 返回结果是按照顺序的
	for i := 0; i < len(oids); i++ {
		oid := oids[i]
		cnt := cnts[i]
		num, err := strconv.ParseUint(cnt, 10, 64)
		if err != nil {
			num = 0
		}
		result[oid] = num
	}

	return result, nil
}

func (c *Cache) SetReplyCount(ctx context.Context, oid, count uint64) error {
	return c.rd.SetCtx(ctx, getCountCmtKey(oid), strconv.FormatUint(count, 10))
}

func (c *Cache) BatchSetReplyCount(ctx context.Context, batch map[uint64]uint64) error {
	err := c.rd.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		for oid, cnt := range batch {
			p.Set(ctx, getCountCmtKey(oid), strconv.FormatUint(cnt, 10), 0)
		}
		return nil
	})
	return err
}

func (c *Cache) IncrReplyCount(ctx context.Context, oid uint64, increment int64) error {
	_, err := c.rd.IncrbyCtx(ctx, getCountCmtKey(oid), increment)
	return err
}

func (c *Cache) DecrReplyCount(ctx context.Context, oid uint64, decrement int64) error {
	_, err := c.rd.DecrbyCtx(ctx, getCountCmtKey(oid), decrement)
	return err
}

func (c *Cache) DelReplyCount(ctx context.Context, oid uint64) error {
	_, err := c.rd.DelCtx(ctx, getCountCmtKey(oid))
	return err
}

func (c *Cache) IncrReplyCountWhenExist(ctx context.Context, oid uint64, increment int64) error {
	const script = `
		local key = KEYS[1]
		local value
		if redis.call('exists', key) == 1 then
    	value = redis.call('incr', key)
		end
		return value`
	var err error
	c.incrOnce.Do(func() {
		c.incrSha, err = c.rd.ScriptLoadCtx(ctx, script)
		xlog.Msg(fmt.Sprintf("incrSha = %s", c.incrSha)).Info()
	})

	if err != nil {
		return err
	}

	_, err = c.rd.EvalShaCtx(ctx, c.incrSha, []string{getCountCmtKey(oid)})
	return err
}

func (c *Cache) DecrReplyCountWhenExist(ctx context.Context, oid uint64, decrement int64) error {
	const script = `
		local key = KEYS[1]
		local value
		if redis.call('exists', key) == 1 then
    	value = redis.call('decr', key)
		end
		return value`
	var err error
	c.decrOnce.Do(func() {
		c.decrSha, err = c.rd.ScriptLoadCtx(ctx, script)
		xlog.Msg(fmt.Sprintf("decrSha = %s", c.decrSha)).Info()
	})

	if err != nil {
		return err
	}

	_, err = c.rd.EvalShaCtx(ctx, c.decrSha, []string{getCountCmtKey(oid)})
	return err
}
