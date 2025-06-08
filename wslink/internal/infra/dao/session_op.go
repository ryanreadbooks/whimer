package dao

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
	zeroredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// session are stored in redis in the following manner:
// uid -> set(sess_id1, sess_id2, sess_id3)
// sess_id -> hash(type struct [Session])

type SessionDao struct {
	cache *zeroredis.Redis
}

func getUidSessKey(uid int64) string {
	return fmt.Sprintf("wslink:uid:%d", uid)
}

func getSessKey(sid string) string {
	return fmt.Sprintf("wslink:sess:%s", sid)
}

func NewSessionDao(cache *zeroredis.Redis) *SessionDao {
	return &SessionDao{
		cache: cache,
	}
}

var (
	//go:embed lua/create_session.lua
	createLua    string
	createScript = zeroredis.NewScript(createLua)
)

// create new session
func (d *SessionDao) Create(ctx context.Context, sess *Session) error {
	var data map[string]any
	err := mapstructure.Decode(sess, &data)
	if err != nil {
		return xsql.ConvertError(err)
	}

	args := maps.KVs(data)

	sessKey := getSessKey(sess.Id)
	uidKey := getUidSessKey(sess.Uid)
	_, err = d.cache.ScriptRunCtx(ctx, createScript, []string{
		sessKey,
		uidKey,
	}, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// get session by session id, return ErrNoRecord if not found
func (d *SessionDao) GetById(ctx context.Context, id string) (*Session, error) {
	pipe, err := d.cache.TxPipeline()
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	var s Session
	cmd := pipe.HGetAll(ctx, getSessKey(id))
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	r, _ := cmd.Result()
	if len(r) == 0 {
		return nil, xsql.ErrNoRecord
	}
	err = cmd.Scan(&s)

	return &s, xsql.ConvertError(err)
}

var (
	//go:embed lua/delete_session.lua
	deleteLua    string
	deleteScript = zeroredis.NewScript(deleteLua)
)

// delete session by session id,
// uid session members will also be removed,
// if id does not exist, nothing will happen
func (d *SessionDao) DeleteById(ctx context.Context, id string) error {
	sess, err := d.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil
		}

		return err
	}

	sessKey := getSessKey(id)
	uidKey := getUidSessKey(sess.Uid)
	_, err = d.cache.ScriptRunCtx(ctx, deleteScript, []string{
		sessKey,
		uidKey,
	})

	return xsql.ConvertError(err)
}

// get user sessions by uid
func (d *SessionDao) GetByUid(ctx context.Context, uid int64) ([]*Session, error) {
	sessIds, err := d.cache.SmembersCtx(ctx, getUidSessKey(uid))
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	result := make([]*Session, 0, len(sessIds))
	if len(sessIds) == 0 {
		return result, nil
	}

	pipe, err := d.cache.TxPipeline()
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	cmds := make([]*redis.MapStringStringCmd, 0, len(sessIds))
	for _, sessId := range sessIds {
		cmd := pipe.HGetAll(ctx, sessId)
		cmds = append(cmds, cmd)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	for _, cmd := range cmds {
		r, _ := cmd.Result()
		if len(r) > 0 {
			var s Session
			err = cmd.Scan(&s)
			if err == nil {
				result = append(result, &s)
			}
		}
	}

	return result, nil
}

// delete all sessions belonging to uid
func (d *SessionDao) DeleteByUid(ctx context.Context, uid int64) error {
	uidKey := getUidSessKey(uid)
	sessIds, err := d.cache.SmembersCtx(ctx, uidKey)
	if err != nil {
		return xsql.ConvertError(err)
	}

	if len(sessIds) == 0 {
		return nil
	}

	delKeys := sessIds
	delKeys = append(delKeys, uidKey)

	_, err = d.cache.DelCtx(ctx, delKeys...)
	return xsql.ConvertError(err)
}

// update the last_active_time field in session with id
func (d *SessionDao) UpdateLastActiveTime(ctx context.Context, id string, t int64) error {
	err := d.cache.HsetCtx(ctx, getSessKey(id), "last_active_time", strconv.FormatInt(t, 10))
	return xsql.ConvertError(err)
}

// update the status field in session with id
func (d *SessionDao) UpdateStatus(ctx context.Context, id string, status ws.SessionStatus) error {
	err := d.cache.HsetCtx(ctx, getSessKey(id), "status", string(status))
	return xsql.ConvertError(err)
}

func (d *SessionDao) SetTTL(ctx context.Context, id string, sec int) error {
	err := d.cache.ExpireCtx(ctx, getSessKey(id), sec)
	return xsql.ConvertError(err)
}
