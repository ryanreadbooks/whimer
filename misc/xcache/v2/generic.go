package v2

import "context"

func (c *Cache[T]) Del(ctx context.Context, keys ...string) error {
	_, err := c.r.DelCtx(ctx, keys...)
	return err
}
