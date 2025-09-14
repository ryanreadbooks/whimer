package record

import (
	"time"
)

type cacheOption struct {
	KeyPrefix string
	Expire    time.Duration
}

type CacheOption func(*cacheOption)

func WithExpire(d time.Duration) CacheOption {
	return func(co *cacheOption) {
		co.Expire = d
	}
}

func WithKeyPrefix(s string) CacheOption {
	return func(co *cacheOption) {
		co.KeyPrefix = s
	}
}
