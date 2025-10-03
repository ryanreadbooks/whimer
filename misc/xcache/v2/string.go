package v2

import (
	"context"
	"maps"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/generics"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

type Fetcher[T any] func(ctx context.Context) (T, time.Duration, error)

type MFetcher[T any] func(ctx context.Context, keys []string) (map[string]T, error)

func (c *Cache[T]) setTFn(ctx context.Context, opt *cacheOption,
	key string, t T, ttl time.Duration) error {

	cc, err := opt.serializer.Marshal(t)
	if err != nil {
		return err
	}

	return c.r.SetexCtx(ctx, key, xstring.FromBytes(cc), int(ttl.Seconds()))
}

func (c *Cache[T]) Get(ctx context.Context, key string, opts ...Option) (t T, err error) {
	return c.GetOrFetch(ctx, key, nil, opts...)
}

func (c *Cache[T]) GetOrFetch(ctx context.Context, key string, fetcher Fetcher[T], opts ...Option) (t T, err error) {
	opt := generics.MakeOpt(opts...)

	resp, err := c.r.GetCtx(ctx, key)
	if err != nil {
		if fetcher != nil {
			var ttl time.Duration
			t, ttl, err = fetcher(ctx)
			if err != nil {
				return
			}

			if ttl == 0 {
				ttl = time.Duration(opt.ttlSec) * time.Second
			}

			// we can put result back to cache here
			c.setCacheBack(ctx, opt, func(ctx context.Context) error {
				return c.setTFn(ctx, opt, key, t, ttl)
			})
		}

		return
	}

	// err == nil
	err = opt.serializer.Unmarshal(xstring.AsBytes(resp), &t)
	if err != nil {
		return
	}

	return
}

func (c *Cache[T]) Set(ctx context.Context, key string, val T, opts ...Option) error {
	opt := generics.MakeOpt(opts...)
	content, err := opt.serializer.Marshal(val)
	if err != nil {
		return err
	}

	return c.r.SetCtx(ctx, key, xstring.FromBytes(content))
}

func (c *Cache[T]) Setex(ctx context.Context, key string, val T, seconds int, opts ...Option) error {
	opt := generics.MakeOpt(opts...)
	content, err := opt.serializer.Marshal(val)
	if err != nil {
		return err
	}

	return c.r.SetexCtx(ctx, key, xstring.FromBytes(content), seconds)
}

type ctxFn func(ctx context.Context) error

func (c *Cache[T]) setCacheBack(ctx context.Context, opt *cacheOption, f ctxFn) {
	if opt.bgSet {
		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "misc.cachev2.bgset",
			Job: func(ctx context.Context) error {
				return f(ctx)
			},
		})
		return
	}

	f(ctx)
}

func (c *Cache[T]) setMapTFn(ctx context.Context, opt *cacheOption, m map[string]T) error {
	pipe, err := c.r.TxPipeline()
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(m))
	args := make([]any, 0, len(m)*2)
	for key, val := range m {
		sv, err := opt.serializer.Marshal(val)
		if err == nil {
			keys = append(keys, key)
			args = append(args, key, sv)
		}
	}

	pipe.MSet(ctx, args...)
	for _, key := range keys {
		pipe.Expire(ctx, key, time.Duration(opt.ttlSec)*time.Second)
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache[T]) mgetTotalFallback(ctx context.Context,
	keys []string,
	fetcher MFetcher[T],
	opt *cacheOption) (t map[string]T, err error) {

	t, err = fetcher(ctx, keys)
	if err != nil {
		return
	}

	// set back to cache
	c.setCacheBack(ctx, opt, func(ctx context.Context) error { return c.setMapTFn(ctx, opt, t) })

	return t, nil
}

func (c *Cache[T]) mgetPartialFallback(ctx context.Context,
	keys []string,
	curResult []string,
	fetcher MFetcher[T],
	opt *cacheOption) (t map[string]T, err error) {

	t = make(map[string]T, len(keys))
	missings := []string{}
	for idx, str := range curResult { // len(keys) should be equal to len(curResult)
		if str == "" {
			missings = append(missings, keys[idx])
			continue
		}

		var tmp T
		err = opt.serializer.Unmarshal(xstring.AsBytes(str), &tmp)
		if err != nil {
			missings = append(missings, keys[idx])
			continue
		}
		t[keys[idx]] = tmp
	}

	if len(missings) == 0 {
		return
	}

	// we need to partial fallback
	if fetcher != nil {
		var (
			ftm map[string]T
		)
		ftm, err = fetcher(ctx, missings)
		if err != nil {
			return
		}

		maps.Copy(t, ftm)

		// set ftm back to cache
		c.setCacheBack(ctx, opt, func(ctx context.Context) error { return c.setMapTFn(ctx, opt, ftm) })
		return
	}

	return
}

func (c *Cache[T]) MGetOrFetch(ctx context.Context,
	keys []string,
	fetcher MFetcher[T],
	opts ...Option) (t map[string]T, err error) {

	opt := generics.MakeOpt(opts...)

	resp, err := c.r.MgetCtx(ctx, keys...)
	if err != nil {
		if fetcher == nil {
			return
		}

		// 发生错误全都要fallback
		return c.mgetTotalFallback(ctx, keys, fetcher, opt)
	} else {
		// 部分fallback
		return c.mgetPartialFallback(ctx, keys, resp, fetcher, opt)
	}
}
