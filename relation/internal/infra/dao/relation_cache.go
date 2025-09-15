package dao

import "github.com/zeromicro/go-zero/core/stores/redis"

// 用户关注关系的缓存 缓存结构定义如下
// 
// 
type RelationCache struct {
	r *redis.Redis
}

func NewRelationCache(r *redis.Redis) *RelationCache {
	return &RelationCache{
		r: r,
	}
}
