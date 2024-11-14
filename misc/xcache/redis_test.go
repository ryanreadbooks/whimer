package xcache

import (
	"context"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Object struct {
	Name string
	Age  int
}

func TestGet(t *testing.T) {
	r := redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	ctx := context.Background()

	err := New[*Object](r).Set(ctx, "test-x-cache", &Object{Name: "hello", Age: 10})
	t.Log(err)
	o, err := New[*Object](r).Get(ctx, "test-x-cache")
	t.Log(o)
	t.Log(err)

	err = New[string](r).Set(ctx, "xcache-str", "hello world")
	t.Log(err)

	str, err := New[string](r).Get(ctx, "xcache-str")
	t.Log(str, err)

	obj, err := New[Object](r).Get(ctx, "fallback-test", WithGetFallback(func(ctx context.Context) (Object, int, error) {
		return Object{"fallback test here", 123}, 60, nil
	}))
	t.Log(err)
	t.Log(obj)
}
