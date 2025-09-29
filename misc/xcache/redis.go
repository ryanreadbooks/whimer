package xcache

import (
	"context"
	"encoding/json"
	"maps"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/generics"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Cache[T any] struct {
	r *redis.Redis
}

func New[T any](rd *redis.Redis) *Cache[T] {
	return &Cache[T]{
		r: rd,
	}
}

type unmarshaler func([]byte, any) error
type marshaler func(any) ([]byte, error)

// get操作从缓存处获取失败时执行函数
//
// 返回的三个参数为：
//
//	T 结果对象
//	int 结果对象存入缓存的过期时间
//	err error
type GetFallbackFn[T any] func(ctx context.Context) (t T, sec int, err error)

type getOpt[T any] struct {
	unmarshaler unmarshaler
	marshaler   marshaler
	fallback    GetFallbackFn[T]
	bgSet       bool
}

func (o getOpt[T]) Default() getOpt[T] {
	return getOpt[T]{
		unmarshaler: json.Unmarshal,
		marshaler:   json.Marshal,
		bgSet:       false,
		fallback:    nil,
	}
}

type GetOpt[T any] func(o *getOpt[T])

func WithGetUnmarshaler[T any](um func([]byte, any) error) GetOpt[T] {
	return func(o *getOpt[T]) {
		o.unmarshaler = um
	}
}

func WithGetFallback[T any](fn GetFallbackFn[T]) GetOpt[T] {
	return func(o *getOpt[T]) {
		o.fallback = fn
	}
}

func WithGetBgSet[T any](b bool) GetOpt[T] {
	return func(o *getOpt[T]) {
		o.bgSet = b
	}
}

func (c *Cache[T]) WithGetUnmarshaler(um func([]byte, any) error) GetOpt[T] {
	return WithGetUnmarshaler[T](um)
}

func (c *Cache[T]) WithGetFallback(fn GetFallbackFn[T]) GetOpt[T] {
	return WithGetFallback(fn)
}

func (c *Cache[T]) WithGetBgSet(b bool) GetOpt[T] {
	return WithGetBgSet[T](b)
}

// 从缓存中获取对象
// 如果T是一个对象，自动进行json序列化
func (c *Cache[T]) Get(ctx context.Context, key string, opts ...GetOpt[T]) (t T, err error) {
	opt := generics.MakeOpt(opts...)
	resp, err := c.r.GetCtx(ctx, key)
	if err != nil || resp == "" {
		if opt.fallback != nil {
			var sec int
			t, sec, err = opt.fallback(ctx)
			if err != nil {
				return
			}

			// we can put result back to cache here
			sf := func(ctx context.Context) error {
				return c.Setex(ctx, key, t, sec, WithSetMarshaler[T](opt.marshaler))
			}

			if opt.bgSet {
				concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
					Name: "misc.cache.get.bgset",
					Job: func(ctx context.Context) error {
						return sf(ctx)
					},
				})
				return
			}
			sf(ctx)
		}

		return
	}

	err = opt.unmarshaler(utils.StringToBytes(resp), &t)
	if err != nil {
		return
	}
	return
}

type MGetFallbackFn[T any] func(ctx context.Context, missingKeys []string) (t map[string]T, err error)

type mgetOpt[T any] struct {
	unmarshaler unmarshaler
	marshaler   marshaler
	fallback    MGetFallbackFn[T]
	fSec        int
	bgSet       bool
}

func (o mgetOpt[T]) Default() mgetOpt[T] {
	return mgetOpt[T]{
		unmarshaler: json.Unmarshal,
		marshaler:   json.Marshal,
		fallback:    nil,
		fSec:        -1,
		bgSet:       true,
	}
}

type MGetOpt[T any] func(o *mgetOpt[T])

func WithMGetUnmarshaler[T any](um func([]byte, any) error) MGetOpt[T] {
	return func(o *mgetOpt[T]) {
		o.unmarshaler = um
	}
}

func WithMGetFallback[T any](fn MGetFallbackFn[T]) MGetOpt[T] {
	return func(o *mgetOpt[T]) {
		o.fallback = fn
	}
}

func WithMGetBgSet[T any](b bool) MGetOpt[T] {
	return func(o *mgetOpt[T]) {
		o.bgSet = b
	}
}

func WithMGetFallbackSec[T any](sec int) MGetOpt[T] {
	return func(o *mgetOpt[T]) {
		o.fSec = sec
	}
}

func (c *Cache[T]) WithMGetFallbackSec(sec int) MGetOpt[T] {
	return WithMGetFallbackSec[T](sec)
}

func (c *Cache[T]) WithMGetBgSet(b bool) MGetOpt[T] {
	return WithMGetBgSet[T](b)
}

func (c *Cache[T]) WithMGetFallback(fn MGetFallbackFn[T]) MGetOpt[T] {
	return WithMGetFallback(fn)
}

func (c *Cache[T]) WithMGetUnmarshaler(um func([]byte, any) error) MGetOpt[T] {
	return WithMGetUnmarshaler[T](um)
}

const (
	mgetBgSetJobName = "misc.cache.mget.bgset"
)

func (c *Cache[T]) MGet(ctx context.Context, keys []string, opts ...MGetOpt[T]) (t map[string]T, err error) {
	opt := generics.MakeOpt(opts...)
	resp, err := c.r.MgetCtx(ctx, keys...)
	t = make(map[string]T, len(keys))

	setFn := func(ctx context.Context, m map[string]T) error {
		errPipex := c.r.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
			for k, v := range m {
				mv, merr := opt.marshaler(v)
				if merr == nil {
					p.Set(ctx, k, mv, time.Second*time.Duration(opt.fSec))
				} else {
					xlog.Msg("mget marshal err").Err(merr).Extras("v", v, "k", k).Infox(ctx)
				}
			}

			return nil
		})
		if errPipex != nil {
			xlog.Msg("mget setfn pipeline failed").Err(errPipex).Errorx(ctx)
		}

		return errPipex
	}

	if err != nil {
		if opt.fallback == nil {
			return
		}
		// 发生错误全都要fallback
		t, err = opt.fallback(ctx, keys)
		if err != nil {
			return
		}

		// set back to cache
		if opt.bgSet {
			concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
				Name: mgetBgSetJobName,
				Job: func(ctx context.Context) error {
					return setFn(ctx, t)
				},
			})
			return
		}
		setFn(ctx, t)
	} else {
		// 找出哪些需要调用fallback
		fallbackKeys := []string{}
		for idx, str := range resp {
			if str == "" {
				fallbackKeys = append(fallbackKeys, keys[idx])
				continue
			}

			var tmp T
			opt.unmarshaler(utils.StringToBytes(str), &tmp)
			t[keys[idx]] = tmp
		}

		if len(fallbackKeys) == 0 {
			return
		}

		// we need to fallback
		if opt.fallback != nil {
			var (
				ftm map[string]T
			)
			ftm, err = opt.fallback(ctx, fallbackKeys)
			if err != nil {
				return
			}

			maps.Copy(t, ftm)

			// set ftm back to cache
			if opt.bgSet {
				concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
					Name: mgetBgSetJobName,
					Job: func(ctx context.Context) error {
						return setFn(ctx, t)
					},
				})
				return
			}
			setFn(ctx, ftm)
		}
	}

	return
}

type setOpt[T any] struct {
	marshaler func(any) ([]byte, error)
}

func (o setOpt[T]) Default() setOpt[T] {
	return setOpt[T]{
		marshaler: json.Marshal,
	}
}

type SetOpt[T any] func(o *setOpt[T])

func WithSetMarshaler[T any](mr func(any) ([]byte, error)) SetOpt[T] {
	return func(o *setOpt[T]) {
		o.marshaler = mr
	}
}

// 设置对象近缓存中
func (c *Cache[T]) Set(ctx context.Context, key string, value T, opts ...SetOpt[T]) error {
	opt := generics.MakeOpt(opts...)
	content, err := opt.marshaler(value)
	if err != nil {
		return err
	}

	return c.r.SetCtx(ctx, key, utils.Bytes2String(content))
}

func (c *Cache[T]) Setex(ctx context.Context, key string, value T, seconds int, opts ...SetOpt[T]) error {
	opt := generics.MakeOpt(opts...)
	content, err := opt.marshaler(value)
	if err != nil {
		return err
	}

	return c.r.SetexCtx(ctx, key, utils.Bytes2String(content), seconds)
}

func (c *Cache[T]) Del(ctx context.Context, keys ...string) (int, error) {
	return c.r.DelCtx(ctx, keys...)
}

// smembers操作从缓存获取失败时执行函数
//
// 返回的三个参数为：
//
//	[]T 结果对象切片
//	err error
type SmembersFallbackFn[T any] func(ctx context.Context) ([]T, error)

type smembersOpt[T any] struct {
	fallback    SmembersFallbackFn[T]
	unmarshaler unmarshaler
}

func (o smembersOpt[T]) Default() smembersOpt[T] {
	return smembersOpt[T]{
		fallback:    nil,
		unmarshaler: json.Unmarshal,
	}
}

type SmembersOpt[T any] func(o *smembersOpt[T])

// smembers
//
// 不支持自动写入缓存
func (c *Cache[T]) Smembers(ctx context.Context, key string, opts ...SmembersOpt[T]) ([]T, error) {
	opt := generics.MakeOpt(opts...)
	res, err := c.r.SmembersCtx(ctx, key)
	if err != nil {
		// fallback
		if opt.fallback != nil {
			result, err := opt.fallback(ctx)
			if err != nil {
				return nil, err
			}
			return result, err
		}
	}

	// try to unmarshal
	var result = make([]T, 0, len(res))
	for _, r := range res {
		var e T
		err := opt.unmarshaler(utils.StringToBytes(r), &e)
		if err != nil {
			continue
		}
		result = append(result, e)
	}

	return result, nil
}
