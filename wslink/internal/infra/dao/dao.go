package dao

import "github.com/zeromicro/go-zero/core/stores/redis"

type SessionDao struct {
	cache *redis.Redis
}

func NewSessionDao(cache *redis.Redis) *SessionDao {
	return &SessionDao{
		cache: cache,
	}
}
