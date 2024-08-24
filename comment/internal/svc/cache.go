package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	hour = time.Hour
	day  = hour * 24
	week = 7 * day
)

// cache key templates
const (
	// 置顶评论
	pinnedCmtKey = "cmt:pinned:%d"
)

func getPinnedCmtKey(oid uint64) string {
	return fmt.Sprintf(pinnedCmtKey, oid)
}

type Cache struct {
	rd *redis.Redis
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
	err = c.rd.SetexCtx(ctx, getPinnedCmtKey(model.Oid), utils.Bytes2String(s), int(jitter(hour)))

	return err
}

// 返回随机过期时间
func jitter(d time.Duration) int64 {
	sec := time.Duration(rand.Int63n(int64(d.Seconds())))
	return int64(sec)
}
