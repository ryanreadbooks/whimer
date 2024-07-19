package session

import (
	"context"
	"encoding/base64"
	"fmt"
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
	defaultSessPrefix = "whimer:sess:rs:"
	defaultUidPrefix  = "whimer:uid:sessid:"
)

type RedisStoreOpt func(*RedisStore)

func WithSessPrefix(px string) RedisStoreOpt {
	return func(rs *RedisStore) {
		rs.sessPrefix = px
	}
}

func WithUidPrefix(px string) RedisStoreOpt {
	return func(rs *RedisStore) {
		rs.uidPrefix = px
	}
}

func WithSerializer(ser model.SessionSerializer) RedisStoreOpt {
	return func(rs *RedisStore) {
		rs.ser = ser
	}
}

type RedisStore struct {
	cache      *redis.Redis
	ser        model.SessionSerializer
	sessPrefix string
	uidPrefix  string
}

func NewRedisStore(cache *redis.Redis, opts ...RedisStoreOpt) Store {
	rs := &RedisStore{
		cache:      cache,
		ser:        model.JsonSessionSerializer{},
		sessPrefix: defaultSessPrefix,
		uidPrefix:  defaultUidPrefix,
	}

	for _, opt := range opts {
		opt(rs)
	}

	return rs
}

func (r *RedisStore) getSessKey(key string) string {
	return r.sessPrefix + key
}

func (r *RedisStore) getUidKey(uid uint64) string {
	return fmt.Sprintf("%s%d", r.uidPrefix, uid)
}

func (r *RedisStore) Get(ctx context.Context, key string) (sess *model.Session, found bool, err error) {
	result, err := r.cache.GetCtx(ctx, r.getSessKey(key))
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

	// 设置session
	// uid -> (sess_id1, sess_id2, ...)
	// 一个uid可以有多个session, uid采用set结构，每个member为sessid
	// 需要获取具体的session内容，需要通过sessid取获取
	err = r.cache.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		p.SAdd(ctx, r.getUidKey(sess.Uid), key)                                // member中的key不带前缀
		p.SetEx(ctx, r.getSessKey(key), value, time.Second*time.Duration(ttl)) // 作为string-key的带前缀

		res, err := p.Exec(ctx)
		if err != nil {
			return err
		}

		for _, cmd := range res {
			if cmd.Err() != nil {
				return cmd.Err()
			}
		}

		return nil
	})

	return err
}

func (r *RedisStore) Del(ctx context.Context, key string) error {
	// 两次RT
	sess, found, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	// 在pipeline中移除
	err = r.cache.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		p.Del(ctx, r.getSessKey(key))
		uid := sess.Uid
		p.SRem(ctx, r.getUidKey(uid), r.getSessKey(key))

		res, err := p.Exec(ctx)
		if err != nil {
			return err
		}

		for _, cmd := range res {
			if cmd.Err() != nil {
				return err
			}
		}

		return nil
	})

	return err
}
