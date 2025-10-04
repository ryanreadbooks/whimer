package v2

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/generics"
	"github.com/ryanreadbooks/whimer/misc/xstring"
)

func (c *Cache[T]) unmarshalSlice(cacheOpt *cacheOption, ss []string) ([]T) {
	var data = make([]T, 0, len(ss))
	for _, r := range ss {
		var e T
		err := cacheOpt.serializer.Unmarshal(xstring.AsBytes(r), &e)
		if err != nil {
			continue
		}
		data = append(data, e)
	}

	return data
}

func (c *Cache[T]) LRange(ctx context.Context, key string, start, stop int, opts ...Option) ([]T, error) {
	cacheOpt := generics.MakeOpt(opts...)

	result, err := c.r.LrangeCtx(ctx, key, start, stop)
	if err != nil {
		return nil, err
	}

	return c.unmarshalSlice(cacheOpt, result), nil
}
