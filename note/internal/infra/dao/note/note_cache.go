package note

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xtime"
)

const (
	noteCacheKey = "note:note:%d"
)

func getNoteCacheKey(nid int64) string {
	return fmt.Sprintf(noteCacheKey, nid)
}

func (d *NoteDao) CacheGetNote(ctx context.Context, nid int64) (*Note, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	res, err := d.cache.GetCtx(ctx, getNoteCacheKey(nid))
	if err != nil {
		return nil, err
	}

	var ret Note
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (d *NoteDao) CacheDelNote(ctx context.Context, nid int64) error {
	if d.cache == nil {
		return nil
	}

	_, err := d.cache.DelCtx(ctx, getNoteCacheKey(nid))
	return err
}

func (d *NoteDao) CacheSetNote(ctx context.Context, model *Note) error {
	if d.cache == nil {
		return nil
	}

	content, err := json.Marshal(model)
	if err != nil {
		return err
	}

	ttl := xtime.Week + xtime.JitterDuration(2*xtime.Hour)
	err = d.cache.SetexCtx(ctx, getNoteCacheKey(model.Id), utils.Bytes2String(content), int(ttl))
	return err
}
