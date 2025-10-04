package v2

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/generics"
)

type SortedSetFetcher[T any] func(ctx context.Context) ([]T, error)

func (c *Cache[T]) SmembersOrFetch(ctx context.Context,
	key string,
	fetcher SortedSetFetcher[T],
	opts ...Option) ([]T, error) {

	cacheOpt := generics.MakeOpt(opts...)
	res, err := c.r.SmembersCtx(ctx, key)
	if err != nil {
		// fallback
		if fetcher != nil {
			result, err := fetcher(ctx)
			return result, err
		}
	}

	return c.unmarshalSlice(cacheOpt, res), nil
}
