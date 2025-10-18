package dao

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xcache/functions"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	maxCachedFanListCount = global.MaxFanListCountForDisplay

	// count cache在此数字下时 每次follow和unfollow都会删除count cache
	// 否则就incr和decr 并定期重建来保证数据一致性
	delFollowingCountThreshold = 200
	delFanCountThreshold       = 200
)

const (
	linkCacheKeyTmpl          = "relation:link:status:%d:%d" // 关注状态
	fansListCacheKeyTmpl      = "relation:fans:zset:"        // 粉丝列表
	followingListCacheKeyTmpl = "relation:followings:zset:"  // 关注列表
	fanCountKeyTmpl           = "relation:count:fans:"       // 粉丝数量
	followingCountKeyTmpl     = "relation:count:followings:" // 关注数量
)

//go:embed lua/relation_cache_fn.lua
var luaFunctionCodes string

// 用户关注关系的缓存 缓存结构定义如下
//
// 仅缓存最新的n个关注的粉丝从而限制数量
// Sorted set as fans list: [{uid, time}, {uid, time}, ..., {uid, time}]
//
// 外层限制最大关注人数从而限制数量
// Sorted set as followings list: [{uid, time}, {uid, time}, ..., {uid, time}]
//
// String as link cache, key=alpha:beta, value=(0,2,-2,-4)
type RelationCache struct {
	r *redis.Redis
}

func NewRelationCache(r *redis.Redis) *RelationCache {
	return &RelationCache{
		r: r,
	}
}

func (r *RelationCache) InitFunctions(ctx context.Context) error {
	err := functions.FunctionLoadReplace(ctx, r.r, luaFunctionCodes)
	if err != nil {
		return err
	}

	return nil
}

func getFanListCacheKey(uid int64) string {
	// relation:fans:zset:uid
	return fansListCacheKeyTmpl + xconv.FormatInt(uid)
}

func getFollowingListCacheKey(uid int64) string {
	// relation:followings:zset:uid
	return followingListCacheKeyTmpl + xconv.FormatInt(uid)
}

func getLinkCacheKey(a, b int64) string {
	// relation:link:status:a:b
	return fmt.Sprintf(linkCacheKeyTmpl, a, b)
}

func getFanCountCacheKey(uid int64) string {
	return fanCountKeyTmpl + xconv.FormatInt(uid)
}

func getFollowingCountCacheKey(uid int64) string {
	return followingCountKeyTmpl + xconv.FormatInt(uid)
}

// initiator执行关注动作后设置缓存
func (c *RelationCache) Follow(ctx context.Context, initiator int64, r *Relation) error {
	// 1. set link cache
	// 2. zadd following list
	// 3. zadd fan list
	// 4. incr or delete fan count/following count
	r = enforceRelationRule(r)

	var (
		followee   int64 // 被关注的uid
		followTime int64 // 被关注的时间

		fanListKey        string
		fanCountKey       string
		linkKey           = getLinkCacheKey(r.UserAlpha, r.UserBeta)
		followingListKey  = getFollowingListCacheKey(initiator)
		followingCountKey = getFollowingCountCacheKey(initiator)
	)

	if initiator == r.UserAlpha {
		fanListKey = getFanListCacheKey(r.UserBeta)
		fanCountKey = getFanCountCacheKey(r.UserBeta)
		followee = r.UserBeta
		followTime = r.Amtime
	} else {
		fanListKey = getFanListCacheKey(r.UserAlpha)
		fanCountKey = getFanCountCacheKey(r.UserAlpha)
		followee = r.UserAlpha
		followTime = r.Bmtime
	}

	_, err := functions.FunctionCall(ctx, c.r, "relation_do_follow",
		[]string{linkKey, followingListKey, fanListKey, followingCountKey, fanCountKey},
		int8(r.Link), xtime.NDayJitter(5, time.Hour*2).Seconds(),
		initiator, followee, followTime,
		maxCachedFanListCount,
		delFollowingCountThreshold, delFanCountThreshold,
	)
	if err != nil {
		return xerror.Wrapf(err, "function relation_do_follow failed")
	}

	return nil
}

// initiator执行取关动作后设置缓存
func (c *RelationCache) UnFollow(ctx context.Context, initiator int64, r *Relation) error {
	// 1. del link cache
	// 2. zrem following list
	// 3. zrem fan list
	// 4. decr or delete fan count/following count
	r = enforceRelationRule(r)
	linkKey := getLinkCacheKey(r.UserAlpha, r.UserBeta)
	followingListKey := getFollowingListCacheKey(initiator)
	var (
		followee int64 // 被关注的uid

		fanListKey        string
		fanCountKey       string
		followingCountKey = getFollowingCountCacheKey(initiator)
	)

	if initiator == r.UserAlpha {
		fanListKey = getFanListCacheKey(r.UserBeta)
		fanCountKey = getFanCountCacheKey(r.UserBeta)
		followee = r.UserBeta
	} else {
		fanListKey = getFanListCacheKey(r.UserAlpha)
		fanCountKey = getFanCountCacheKey(r.UserAlpha)
		followee = r.UserAlpha
	}

	_, err := functions.FunctionCall(ctx, c.r, "relation_do_unfollow",
		[]string{linkKey, followingListKey, fanListKey, followingCountKey, fanCountKey},
		initiator, followee,
		delFollowingCountThreshold, delFanCountThreshold,
	)
	if err != nil {
		return xerror.Wrapf(err, "function relation_do_unfollow failed")
	}

	return nil
}

// 设置a和b的关系是link
func (c *RelationCache) SetLink(ctx context.Context, alpha, beta int64, link LinkStatus) error {
	alpha, beta, link = enforceUidRuleWithLink(alpha, beta, link)
	key := getLinkCacheKey(alpha, beta)
	err := c.r.SetexCtx(ctx, key, xconv.FormatInt(link), int(xtime.NDayJitter(5, time.Minute*20).Seconds()))
	if err != nil {
		return xerror.Wrapf(err, "setex failed")
	}

	return nil
}

func (c *RelationCache) BatchSetLinks(ctx context.Context, datas []CacheLink) error {
	if len(datas) == 0 {
		return nil
	}

	type pipeDataType struct {
		key    string
		link   int8
		expire time.Duration
	}

	pipeDatas := []pipeDataType{}
	args := make([]any, 0, len(pipeDatas)*2)
	for _, data := range datas {
		alpha, beta, link := enforceUidRuleWithLink(data.Alpha, data.Beta, data.Link)
		key := getLinkCacheKey(alpha, beta)
		pipeDatas = append(pipeDatas, pipeDataType{
			key:    key,
			link:   int8(link),
			expire: xtime.NDayJitter(5, time.Hour),
		})
		args = append(args, key, int8(link))
	}

	pipe, err := c.r.TxPipeline()
	if err != nil {
		return xerror.Wrapf(err, "begin pipeline failed")
	}

	pipe.MSet(ctx, args...)
	for _, pd := range pipeDatas {
		pipe.Expire(ctx, pd.key, pd.expire)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return xerror.Wrapf(err, "pipeline exec failed")
	}

	return nil
}

type CacheLink struct {
	Alpha int64
	Beta  int64
	Link  LinkStatus
}

// 获取a和b的关系
func (c *RelationCache) GetLink(ctx context.Context, alpha, beta int64) (*Relation, error) {
	alpha, beta = enforceUidRule(alpha, beta)
	key := getLinkCacheKey(alpha, beta)
	cl := &Relation{
		UserAlpha: alpha,
		UserBeta:  beta,
	}
	res, err := c.r.GetCtx(ctx, key)
	if err != nil {
		return cl, xerror.Wrapf(err, "get failed")
	}

	link, err := strconv.ParseInt(res, 10, 8)
	if err != nil {
		// dirty data
		c.r.DelCtx(ctx, key)
		return cl, xerror.Wrapf(err, "key %s can not be parsed into int8", key)
	}

	cl.Link = LinkStatus(link)

	return cl, nil
}

type UserPair struct {
	UserA, UserB int64
}

// 批量获取关注关系
func (c *RelationCache) BatchGetLinks(ctx context.Context, users []UserPair) ([]*Relation, error) {
	keys := make([]string, 0, len(users))
	for _, pair := range users {
		alpha, beta := enforceUidRule(pair.UserA, pair.UserB)
		keys = append(keys, getLinkCacheKey(alpha, beta))
	}

	res, err := c.r.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, xerror.Wrapf(err, "mget failed")
	}

	dirtyKeys := make([]string, 0, len(keys))
	cachedData := make([]*Relation, 0, len(keys))
	for idx, r := range res {
		key := keys[idx]

		if len(r) == 0 {
			continue
		}

		// found
		link, err := strconv.ParseInt(r, 10, 8)
		if err != nil {
			// dirty data
			dirtyKeys = append(dirtyKeys, key)
			continue
		}
		linkSt := LinkStatus(link)

		// found and valid
		user := users[idx]
		user.UserA, user.UserB = enforceUidRule(user.UserA, user.UserB)
		cachedData = append(cachedData, &Relation{
			UserAlpha: user.UserA,
			UserBeta:  user.UserB,
			Link:      linkSt,
		})
	}

	if len(dirtyKeys) != 0 {
		c.r.DelCtx(ctx, dirtyKeys...)
	}

	return cachedData, nil
}

// 把other加入uid的粉丝列表
// func (c *RelationCache) AddFanList(ctx context.Context, uid int64, other int64, time int64) error {
// 	key := getFanListCacheKey(uid)
// 	_, err := c.r.ZaddCtx(ctx, key, time, xconv.FormatInt(other))
// 	if err != nil {
// 		return xerror.Wrapf(err, "zadd failed")
// 	}

// 	return nil
// }

// 获取uid粉丝列表长度（粉丝数量）
func (c *RelationCache) CountFanList(ctx context.Context, uid int64) (int64, error) {
	key := getFanListCacheKey(uid)
	cnt, err := c.r.ZcardCtx(ctx, key)
	if err != nil {
		return 0, xerror.Wrapf(err, "zcard failed")
	}

	return int64(cnt), nil
}

// 把other加入uid的关注列表
// func (c *RelationCache) AddFollowingList(ctx context.Context, uid int64, other int64, time int64) error {
// 	key := getFollowingListCacheKey(uid)
// 	_, err := c.r.ZaddCtx(ctx, key, time, xconv.FormatInt(other))
// 	if err != nil {
// 		return xerror.Wrapf(err, "zadd failed")
// 	}

// 	return nil
// }

// 获取uid关注列表长度（关注数量）
func (c *RelationCache) CountFollowingList(ctx context.Context, uid int64) (int64, error) {
	key := getFollowingListCacheKey(uid)
	cnt, err := c.r.ZcardCtx(ctx, key)
	if err != nil {
		return 0, xerror.Wrapf(err, "zcard failed")
	}

	return int64(cnt), nil
}

func (c *RelationCache) setCount(ctx context.Context, key string, count int64) error {
	// 一天过期时间
	err := c.r.SetexCtx(ctx, key, xconv.FormatInt(count), xtime.DayJitterSec(time.Minute*15))
	if err != nil {
		return xerror.Wrapf(err, "setex failed")
	}

	return nil
}

func (c *RelationCache) getCount(ctx context.Context, key string) (int64, error) {
	res, err := c.r.GetCtx(ctx, key)
	if err != nil {
		return 0, xerror.Wrapf(err, "get failed")
	}

	cnt, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		defer func() { concurrent.SafeGo(func() { c.r.DelCtx(ctx, key) }) }()
		return 0, xerror.Wrapf(err, "invalid count string")
	}

	if cnt < 0 {
		defer func() { concurrent.SafeGo(func() { c.r.DelCtx(ctx, key) }) }()
		cnt = 0
	}

	return cnt, nil
}

func (c *RelationCache) SetFanCount(ctx context.Context, uid, count int64) error {
	return c.setCount(ctx, getFanCountCacheKey(uid), count)
}

func (c *RelationCache) GetFanCount(ctx context.Context, uid int64) (int64, error) {
	return c.getCount(ctx, getFanCountCacheKey(uid))
}

func (c *RelationCache) SetFollowingCount(ctx context.Context, uid, count int64) error {
	return c.setCount(ctx, getFollowingCountCacheKey(uid), count)
}

func (c *RelationCache) GetFollowingCount(ctx context.Context, uid int64) (int64, error) {
	return c.getCount(ctx, getFollowingCountCacheKey(uid))
}
