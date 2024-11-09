package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/passport/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// convenient time units definition
	second = time.Second
	minute = time.Minute
	hour   = time.Hour
	day    = 24 * hour

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

func (m *Manager) MarshalUserBase(user *dao.UserBase) (string, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return "", xerror.Wrapf(global.ErrInternal.Msg(err.Error()), "json marshal userbase failed")
	}

	return utils.Bytes2String(data), nil
}

func (m *Manager) UnmarshalUserBase(data string) (*dao.UserBase, error) {
	var res = new(dao.UserBase)
	err := json.Unmarshal(utils.StringToBytes(data), res)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrInternal.Msg(err.Error()), "json unmarshal userbase failed")
	}

	return res, nil
}

// Get 获取Session
func (m *Manager) GetSession(ctx context.Context, sessId string) (*model.Session, error) {
	sess, found, err := m.store.Get(ctx, sessId)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrInternal.Msg(err.Error()), "store get failed")
	}

	// 找不到或者过期都当作登录过期
	if !found || sess.Meta.ExpireAt < time.Now().Unix() {
		return nil, xerror.Wrap(global.ErrSessInvalidated)
	}

	return sess, nil
}

// 创建新的session并且存储
func (m *Manager) NewSession(ctx context.Context, user *dao.UserBase, platform string) (*model.Session, error) {
	sessId, err := getToken()
	if err != nil {
		return nil, xerror.Wrap(global.ErrCheckIn)
	}

	expireAt := time.Now().Add(defaultSessionTTL) // 15天过期时间
	session := model.NewSession(sessId, expireAt.Unix())
	session.Uid = user.Uid
	session.Platform = platform
	detail, err := m.MarshalUserBase(user)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrCheckIn, "marshal userbase failed")
	}
	session.Detail = detail

	err = m.store.Set(ctx, sessId, session)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrCheckIn, "store set failed")
	}

	return session, nil
}

// 立即令session过期
func (m *Manager) InvalidateSession(ctx context.Context, sessId string) error {
	if len(sessId) == 0 {
		return nil
	}

	if err := m.store.Del(ctx, sessId); err != nil {
		return xerror.Wrapf(global.ErrInternal.Msg(err.Error()), "store del failed")
	}

	return nil
}

// 给一个已经存在的session续期
func (m *Manager) RenewSession(ctx context.Context, sessId string) error {
	if len(sessId) == 0 {
		return nil
	}

	sess, found, err := m.store.Get(ctx, sessId)
	if err != nil {
		return xerror.Wrapf(global.ErrSessRenewal, "store get failed")
	}

	if !found {
		return nil
	}

	// 更新过期时间
	sess.Meta.ExpireAt = time.Now().Add(defaultSessionTTL).Unix()
	err = m.store.Set(ctx, sessId, sess)
	if err != nil {
		return xerror.Wrapf(global.ErrSessRenewal, "cache get failed")
	}

	return nil
}
