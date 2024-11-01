package svc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	noteCacheKey = "note:note:%d"
)

func getNoteCacheKey(nid uint64) string {
	return fmt.Sprintf(noteCacheKey, nid)
}

type NoteAdminCache struct {
	rd *redis.Redis
}

func NewNoteAdminCache(rd *redis.Redis) *NoteAdminCache {
	return &NoteAdminCache{
		rd: rd,
	}
}

func (c *NoteAdminCache) GetNote(ctx context.Context, nid uint64) (*noterepo.Model, error) {
	res, err := c.rd.GetCtx(ctx, getNoteCacheKey(nid))
	if err != nil {
		return nil, err
	}

	var ret noterepo.Model
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (c *NoteAdminCache) DelNote(ctx context.Context, nid uint64) error {
	_, err := c.rd.HdelCtx(ctx, getNoteCacheKey(nid))
	return err
}

func (c *NoteAdminCache) SetNote(ctx context.Context, model *noterepo.Model) error {
	content, err := json.Marshal(model)
	if err != nil {
		return err
	}

	ttl := xtime.Week + xtime.JitterDuration(2*xtime.Hour)
	err = c.rd.SetexCtx(ctx, getNoteCacheKey(model.Id), utils.Bytes2String(content), int(ttl))
	return err
}
