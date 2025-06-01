package session

import (
	"context"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Store只提供存储功能
type Store interface {
	// 获取id为key的session
	// 返回找到的session, 如果found为true, 表示没有对应的key的session
	// err不为nil时, 表示发生了系统错误
	Get(ctx context.Context, key string) (sess *model.Session, found bool, err error)
	// 批量获取
	BatchGet(ctx context.Context, keys []string) ([]*model.Session, error)
	// 获取uid所有的session
	GetUid(ctx context.Context, uid int64) ([]*model.Session, error)
	// 设置id为key的session 存在则覆盖
	Set(ctx context.Context, key string, sess *model.Session) error
	// 立即删除id为key的session, 如果不存在 则操作为no-op
	Del(ctx context.Context, key string) error
	// 批量删除sessId
	BatchDel(ctx context.Context, keys []string) error
	// 删除uid的所有session, 如果不存在, 则操作为no-op
	DelUid(ctx context.Context, uid int64) error
}

const (
	defaultSessPrefix = "passport:sess:rs:"
	defaultUidPrefix  = "passport:uid:sessid:"
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
		ser:        model.JsonSessionSerializer{}, // Json序列化
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

func (r *RedisStore) getUidKey(uid int64) string {
	return r.uidPrefix + strconv.FormatInt(uid, 10)
}

func (r *RedisStore) parseGetResult(result string) (*model.Session, error) {
	data, err := base64.StdEncoding.DecodeString(result)
	if err != nil {
		return nil, err
	}

	sess, err := r.ser.Deserialize(data)
	if err != nil {
		return nil, err
	}

	return sess, nil
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

	sess, err = r.parseGetResult(result)
	if err != nil {
		return
	}

	found = true
	return
}

func (r *RedisStore) BatchGet(ctx context.Context, keys []string) ([]*model.Session, error) {
	l := len(keys)
	var res = make([]*model.Session, 0, l)
	if l == 0 {
		return res, nil
	}

	sessKeys := make([]string, 0, l)
	for _, k := range keys {
		if k != "" {
			sessKeys = append(sessKeys, r.getSessKey(k))
		}
	}

	sessDatas, err := r.cache.MgetCtx(ctx, sessKeys...)
	if err != nil {
		return nil, err
	}

	for _, sessData := range sessDatas {
		if len(sessData) > 0 {
			sess, err := r.parseGetResult(sessData)
			if err == nil {
				res = append(res, sess)
			}
		}
	}

	return res, nil
}

func (r *RedisStore) GetUid(ctx context.Context, uid int64) ([]*model.Session, error) {
	res, err := r.cache.SmembersCtx(ctx, r.getUidKey(uid))
	if err != nil {
		return nil, err
	}

	return r.BatchGet(ctx, res)
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
	// sess_id -> session data (string)
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
		p.SRem(ctx, r.getUidKey(uid), key)

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

func (r *RedisStore) BatchDel(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	sesses, err := r.BatchGet(ctx, keys)
	if err != nil {
		return err
	}

	// 在pipeline中删除
	err = r.cache.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		targets := make(map[int64][]string)
		for _, sess := range sesses {
			targets[sess.Uid] = append(targets[sess.Uid], sess.Meta.Id)
		}

		for uid, rawSessKeys := range targets {
			// 删除 uid set中的members
			// 删除 member 对应的key
			members := make([]any, 0, len(rawSessKeys))
			for _, rawSessKey := range rawSessKeys {
				p.Del(ctx, r.getSessKey(rawSessKey))
				members = append(members, rawSessKey)
			}
			p.SRem(ctx, r.getUidKey(uid), members...)
		}

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

func (r *RedisStore) DelUid(ctx context.Context, uid int64) error {
	// 2RT
	sessIds, err := r.cache.SmembersCtx(ctx, r.getUidKey(uid))
	if err != nil {
		return err
	}

	keysToRemove := make([]string, 0, len(sessIds)+1)
	keysToRemove = append(keysToRemove, r.getUidKey(uid))
	for _, sessId := range sessIds {
		keysToRemove = append(keysToRemove, r.getSessKey(sessId))
	}

	_, err = r.cache.DelCtx(ctx, keysToRemove...)
	if err != nil {
		return err
	}

	return nil
}
