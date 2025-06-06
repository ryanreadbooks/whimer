package dao

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// session are stored in redis in the following manner:
// uid -> set(sess_id1, sess_id2, sess_id3)
// sess_id -> json.Marshal(Session)

type SessionDao struct {
	cache *redis.Redis
}

func getUidSessKey(uid int64) string {
	return fmt.Sprintf("wslink:uid:%d", uid)
}

func getSessKey(sid string) string {
	return fmt.Sprintf("wslink:sess:%s", sid)
}

func NewSessionDao(cache *redis.Redis) *SessionDao {
	return &SessionDao{
		cache: cache,
	}
}

var (
	//go:embed lua/create_session.lua
	createLua    string
	createScript = redis.NewScript(createLua)
)

// create new session
func (d *SessionDao) Create(ctx context.Context, sess *Session) error {
	sessData, err := json.Marshal(sess)
	if err != nil {
		return xsql.ConvertError(err)
	}

	sessKey := getSessKey(sess.Id)
	uidKey := getUidSessKey(sess.Uid)
	_, err = d.cache.ScriptRunCtx(ctx, createScript, []string{
		sessKey,
		uidKey,
	}, utils.Bytes2String(sessData))
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// get session by session id, return ErrNoRecord if not found
func (d *SessionDao) GetById(ctx context.Context, id string) (*Session, error) {
	res, err := d.cache.GetCtx(ctx, getSessKey(id))
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	if res == "" {
		return nil, xsql.ErrNoRecord
	}

	var s Session
	err = json.Unmarshal(utils.StringToBytes(res), &s)
	return &s, xsql.ConvertError(err)
}

var (
	//go:embed lua/delete_session.lua
	deleteLua    string
	deleteScript = redis.NewScript(deleteLua)
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
