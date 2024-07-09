package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// convenient time units definition
	second = time.Second
	minute = time.Minute
	hour   = time.Hour
	day    = 24 * hour
)

const (
	defaultSessionTTL = 15 * day
)

// 管理session
type Manager struct {
	store Store
}

func NewManager(cache *redis.Redis) *Manager {
	return &Manager{
		store: NewRedisStore(cache),
	}
}

func getToken() (string, error) {
	var buf = make([]byte, 64)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (m *Manager) MarshalUserBasic(user *userbase.Basic) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	return utils.Bytes2String(data), nil
}

func (m *Manager) UnmarshalUserBasic(data string) (*userbase.Basic, error) {
	var res = new(userbase.Basic)
	err := json.Unmarshal(utils.StringToBytes(data), res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Get 获取Session
func (m *Manager) GetSession(ctx context.Context, sessId string) (*model.Session, error) {
	sess, found, err := m.store.Get(ctx, sessId)
	if err != nil {
		logx.Errorf("store get session err: %v, sessId: %s", err, sessId)
		return nil, global.ErrInternal
	}

	// 找不到或者过期都当作登录过期
	if !found || sess.Meta.ExpireAt < time.Now().Unix() {
		return nil, global.ErrSessInvalidated
	}

	return sess, nil
}

// 创建新的session并且存储
func (m *Manager) NewSession(ctx context.Context, user *userbase.Basic, platform string) (*model.Session, error) {
	sessId, err := getToken()
	if err != nil {
		logx.Errorf("getToken err: %v", err)
		return nil, global.ErrSignIn
	}

	expireAt := time.Now().Add(defaultSessionTTL) // 15天过期时间
	session := model.NewSession(sessId, expireAt.Unix())
	session.Uid = user.Uid
	session.Platform = platform
	detail, err := m.MarshalUserBasic(user)
	if err != nil {
		logx.Errorf("marshalUserBasic err: %v", err)
		return nil, global.ErrSignIn
	}
	session.Detail = detail

	err = m.store.Set(ctx, sessId, session)
	if err != nil {
		logx.Errorf("store set sessiong err: %v, sessId: %s", err, sessId)
		return nil, global.ErrSignIn
	}

	return session, nil
}

// 立即令session过期
func (m *Manager) InvalidateSession(ctx context.Context, sessId string) error {
	if len(sessId) == 0 {
		return nil
	}

	return m.store.Del(ctx, sessId)
}

// 给一个已经存在的session续期
func (m *Manager) RenewSession(ctx context.Context, sessId string) error {
	if len(sessId) == 0 {
		return nil
	}

	sess, found, err := m.store.Get(ctx, sessId)
	if err != nil {
		logx.Errorf("store get session err: %v, sessId: %s", err, sessId)
		return global.ErrSessRenewal
	}

	if !found {
		return nil
	}

	// 更新过期时间
	sess.Meta.ExpireAt = time.Now().Add(defaultSessionTTL).Unix()
	err = m.store.Set(ctx, sessId, sess)
	if err != nil {
		logx.Errorf("store set session err: %v, sessId: %s", err, sessId)
		return global.ErrSessRenewal
	}

	return nil
}
