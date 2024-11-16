package xcache

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/utils"
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

type defOptFn[T any] func() T

type opter[T any] interface {
	~func(*T)
}

func injectOpt[T any, P opter[T]](def defOptFn[T], opts ...P) *T {
	opt := def()
	for _, o := range opts {
		o(&opt)
	}

	return &opt
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

// 从缓存中获取对象
// 如果T是一个对象，自动进行json序列化
func (c *Cache[T]) Get(ctx context.Context, key string, opts ...GetOpt[T]) (t T, err error) {
	opt := injectOpt[getOpt[T]](getOptDefault[T], opts...)

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
	opt := injectOpt[setOpt[T]](setOptDefault[T], opts...)
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

func smembersDefaultOpt[T any]() smembersOpt[T] {
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
	opt := injectOpt[smembersOpt[T]](smembersDefaultOpt[T], opts...)

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
