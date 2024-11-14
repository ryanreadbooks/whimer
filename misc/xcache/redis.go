package xcache

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// var (
// 	r *redis.Redis
// )

// func Init(rd *redis.Redis) {
// 	r = rd
// }

// 从缓存处获取失败时执行函数
// 返回的三个参数为：
//
//	T 结果对象
//	int 结果对象存入缓存的过期时间
//	err error
type FallbackFn[T any] func(ctx context.Context) (t T, sec int, err error)

type getOpt[T any] struct {
	unmarshaler func([]byte, any) error
	marshaler   func(any) ([]byte, error)
	fallback    FallbackFn[T]
	bgSet       bool
}

func getOptDefault[T any]() getOpt[T] {
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

func WithGetFallback[T any](fn FallbackFn[T]) GetOpt[T] {
	return func(o *getOpt[T]) {
		o.fallback = fn
	}
}

func WithGetBgSet[T any](b bool) GetOpt[T] {
	return func(o *getOpt[T]) {
		o.bgSet = b
	}
}

// 从缓存中获取对象
// 如果T是一个对象，自动进行json序列化
func (c *Cache[T]) Get(ctx context.Context, key string, opts ...GetOpt[T]) (t T, err error) {
	opt := getOptDefault[T]()
	for _, o := range opts {
		o(&opt)
	}

	resp, err := c.r.GetCtx(ctx, key)
	if err != nil || resp == "" {
		if opt.fallback != nil {
			var sec int
			t, sec, err = opt.fallback(ctx)
			if err != nil {
				return
			}

			// we can put result back to cache here
			sf := func() { _ = c.Setex(ctx, key, t, sec, WithSetMarshaler[T](opt.marshaler)) }
			if opt.bgSet {
				concurrent.SafeGo(sf)
				return
			}
			sf()
		}

		return
	}

	err = opt.unmarshaler(utils.StringToBytes(resp), &t)
	if err != nil {
		return
	}
	return
}

type setOpt[T any] struct {
	marshaler func(any) ([]byte, error)
}

func setOptDefault[T any]() setOpt[T] {
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
	opt := setOptDefault[T]()
	for _, o := range opts {
		o(&opt)
	}

	content, err := opt.marshaler(value)
	if err != nil {
		return err
	}

	return c.r.SetCtx(ctx, key, utils.Bytes2String(content))
}

func (c *Cache[T]) Setex(ctx context.Context, key string, value T, seconds int, opts ...SetOpt[T]) error {
	opt := setOptDefault[T]()
	for _, o := range opts {
		o(&opt)
	}

	content, err := opt.marshaler(value)
	if err != nil {
		return err
	}

	return c.r.SetexCtx(ctx, key, utils.Bytes2String(content), seconds)
}

type Cache[T any] struct {
	r *redis.Redis
}

func New[T any](rd *redis.Redis) *Cache[T] {
	return &Cache[T]{
		r: rd,
	}
}

func (c *Cache[T]) Del(ctx context.Context, keys ...string) (int, error) {
	return c.r.DelCtx(ctx, keys...)
}
