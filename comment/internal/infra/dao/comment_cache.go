package dao

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/misc/xlog"
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

type CommentCache struct {
	rd       *redis.Redis
	incrOnce sync.Once
	incrSha  string
	decrOnce sync.Once
	decrSha  string
}

func NewCommentCache(rd *redis.Redis) *CommentCache {
	return &CommentCache{
		rd: rd,
	}
}

// 被评论对象的评论数量
func (c *CommentCache) GetReplyCount(ctx context.Context, oid uint64) (uint64, error) {
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

func (c *CommentCache) BatchGetReplyCount(ctx context.Context, oids []uint64) (map[uint64]uint64, error) {
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

func (c *CommentCache) SetReplyCount(ctx context.Context, oid, count uint64) error {
	return c.rd.SetCtx(ctx, getCountCmtKey(oid), strconv.FormatUint(count, 10))
}

func (c *CommentCache) BatchSetReplyCount(ctx context.Context, batch map[uint64]uint64) error {
	err := c.rd.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		for oid, cnt := range batch {
			p.Set(ctx, getCountCmtKey(oid), strconv.FormatUint(cnt, 10), 0)
		}
		return nil
	})
	return err
}

func (c *CommentCache) IncrReplyCount(ctx context.Context, oid uint64, increment int64) error {
	_, err := c.rd.IncrbyCtx(ctx, getCountCmtKey(oid), increment)
	return err
}

func (c *CommentCache) DecrReplyCount(ctx context.Context, oid uint64, decrement int64) error {
	_, err := c.rd.DecrbyCtx(ctx, getCountCmtKey(oid), decrement)
	return err
}

func (c *CommentCache) DelReplyCount(ctx context.Context, oid uint64) error {
	_, err := c.rd.DelCtx(ctx, getCountCmtKey(oid))
	return err
}

func (c *CommentCache) IncrReplyCountWhenExist(ctx context.Context, oid uint64, increment int64) error {
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

func (c *CommentCache) DecrReplyCountWhenExist(ctx context.Context, oid uint64, decrement int64) error {
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
