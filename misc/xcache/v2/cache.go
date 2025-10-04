package v2

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Cache[T any] struct {
	r *redis.Redis
}

func New[T any](r *redis.Redis) *Cache[T] {
	return &Cache[T]{r: r}
}

type cacheOption struct {
	serializer Serializer
	bgSet      bool
	ttlSec     int
}

func (cacheOption) Default() cacheOption {
	return cacheOption{
		serializer: JSONSerializer{},
		bgSet:      false,
		ttlSec:     3600,
	}
}

type Option func(*cacheOption)

func WithSerializer(ser Serializer) Option {
	return func(co *cacheOption) {
		co.serializer = ser
	}
}

func WithBgSet(b bool) Option {
	return func(co *cacheOption) {
		co.bgSet = b
	}
}

func WithTTL(dur time.Duration) Option {
	return func(co *cacheOption) {
		co.ttlSec = int(dur.Seconds())
	}
}
