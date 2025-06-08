package dao

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Dao struct {
	cache *redis.Redis

	SessionDao *SessionDao
}

func New(cache *redis.Redis) *Dao {
	return &Dao{
		cache: cache,
		SessionDao: NewSessionDao(cache),
	}
}
