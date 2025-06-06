package dao

import (
	"context"
	_ "embed"
	"encoding/json"
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

//go:embed lua/create_session.lua
var createLua string

var (
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
