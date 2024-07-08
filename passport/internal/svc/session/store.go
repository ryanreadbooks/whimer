package session

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type Store interface {
	// 获取id为key的session
	// 返回找到的session, 如果found为true, 表示没有对应的key的session
	// err不为nil时, 表示发生了系统错误
	Get(ctx context.Context, key string) (sess *model.Session, found bool, err error)
	// 设置id为key的session
	// 存在则覆盖
	Set(ctx context.Context, key string, sess *model.Session) error
	// 立即删除id为key的session
	// 如果不存在 则操作为no-op
	Del(ctx context.Context, key string) error
}

const (
	defaultPrefix = "whimer:sess:rs:"
)

type RedisStoreOpt func(*RedisStore)

func WithPrefix(px string) RedisStoreOpt {
	return func(rs *RedisStore) {
		rs.prefix = px
	}
}

func WithSerializer(ser model.SessionSerializer) RedisStoreOpt {
	return func(rs *RedisStore) {
		rs.ser = ser
	}
}

type RedisStore struct {
	cache  *redis.Redis
	ser    model.SessionSerializer
	prefix string
}

func NewRedisStore(cache *redis.Redis, opts ...RedisStoreOpt) Store {
	rs := &RedisStore{
		cache:  cache,
		ser:    model.JsonSessionSerializer{},
		prefix: defaultPrefix,
	}

	for _, opt := range opts {
		opt(rs)
	}

	return rs
}

func (r *RedisStore) getKey(key string) string {
	return r.prefix + key
}

func (r *RedisStore) Get(ctx context.Context, key string) (sess *model.Session, found bool, err error) {
	result, err := r.cache.GetCtx(ctx, r.getKey(key))
	if err != nil {
		return
	}

	// result为空 表示没有对应的key
	if len(result) == 0 {
		found = false
		err = nil
		return
	}

	data, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return
	}

	sess, err = r.ser.Deserialize(data)
	if err != nil {
		return
	}

	found = true
	return
}

func (r *RedisStore) Set(ctx context.Context, key string, sess *model.Session) error {
	data, err := r.ser.Serialize(sess)
	if err != nil {
		return err
	}

	// 过期时间设置
	// <=0 -> 没有过期时间
	// >0 -> 过期时间戳 单位second
	expiry := sess.Meta.ExpireAt
	var ttl int64
	if expiry > 0 {
		now := time.Now().Unix()
		if expiry > now {
			// 计算ttl
			ttl = expiry - now
		}
	}

	value := base64.StdEncoding.EncodeToString(data)

	return r.cache.SetexCtx(ctx, r.getKey(key), value, int(ttl))
}

func (r *RedisStore) Del(ctx context.Context, key string) error {
	_, err := r.cache.DelCtx(ctx, r.getKey(key))
	return err
}
