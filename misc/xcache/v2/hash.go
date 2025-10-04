package v2

import (
	"context"
	"reflect"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/ryanreadbooks/whimer/misc/generics"
)

func (c *Cache[T]) hgetallSetCacheBack(ctx context.Context,
	cacheOpt *cacheOption,
	key string,
	t T,
	ttl time.Duration) {
	if ttl == 0 {
		ttl = time.Second * time.Duration(cacheOpt.ttlSec)
	}

	c.setCacheBack(ctx, cacheOpt, func(ctx context.Context) error {
		return c.r.PipelinedCtx(ctx, func(p goredis.Pipeliner) error {
			p.HSet(ctx, key, t) // hset accepts struct with redis field tag
			p.Expire(ctx, key, ttl)
			return nil
		})
	})
}

// you need `redis:"field"` tag in T to make this method work.
// T must be a struct
func (c *Cache[T]) HGetAllOrFetch(ctx context.Context,
	key string,
	fetcher Fetcher[T],
	opt ...Option) (t T, err error) {

	var (
		cacheOpt = generics.MakeOpt(opt...)
		ttl      time.Duration
	)

	res, err := c.r.HgetallCtx(ctx, key)
	if err != nil || len(res) == 0 {
		// fetcher
		if fetcher != nil {
			t, ttl, err = fetcher(ctx)
			if err != nil {
				return
			}
			// fetcher err is nil
			c.hgetallSetCacheBack(ctx, cacheOpt, key, t, ttl)
			return
		}

		// hgetall returned err is not nil and fetcher is not set
		return
	}

	t, err = c.mapStringStringUnmarshal(res)
	if err == nil || fetcher == nil {
		return
	}

	// if we can not scan res into T, we invoke fetcher again
	t, ttl, err = fetcher(ctx)
	if err != nil {
		return
	}

	c.hgetallSetCacheBack(ctx, cacheOpt, key, t, ttl)
	return
}

func (c *Cache[T]) mapStringStringUnmarshal(res map[string]string) (t T, err error) {
	valOfT := reflect.ValueOf(t)
	var dest any
	// if T is a pointer, we need to alloc a new instance
	if valOfT.Kind() == reflect.Pointer {
		// T is *struct, so typeOfT.Elem() is struct
		typeOfT := valOfT.Type()
		// reflect.New (simply a new operator, like new(struct)) returns *struct
		// Interface() return the interface{} type
		t = reflect.New(typeOfT.Elem()).Interface().(T)
		dest = t
	} else {
		// t is not a pointer, we diretly use pointer of t as dest
		dest = &t
	}

	// hgetall returned err is nil, now try to parse result into T
	err = goredis.NewMapStringStringResult(res, err).Scan(dest)
	if err == nil {
		return
	}

	return
}
