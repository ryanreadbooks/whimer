package note

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	noteCacheKey = "note:note:%d"

	noteCountByOwnerCacheKey       = "note:count:uid:"        // note:count:uid:%d
	notePublicCountByOwnerCacheKey = "note:public_count:uid:" // note:public_count:uid:%d
)

func getNoteCacheKey(nid int64) string {
	return fmt.Sprintf(noteCacheKey, nid)
}

func getNoteCountByOwnerCacheKey(uid int64) string {
	return noteCountByOwnerCacheKey + xconv.FormatInt(uid)
}

func getNotePublicCountByOwnerCacheKey(uid int64) string {
	return notePublicCountByOwnerCacheKey + xconv.FormatInt(uid)
}

// NoteCache 笔记缓存操作 - 纯缓存操作
type NoteCache struct {
	cache *redis.Redis
}

func NewNoteCache(cache *redis.Redis) *NoteCache {
	return &NoteCache{
		cache: cache,
	}
}

// GetNote 从缓存获取笔记
func (c *NoteCache) GetNote(ctx context.Context, nid int64) (*NotePO, error) {
	if c.cache == nil {
		return nil, fmt.Errorf("cache is nil")
	}

	res, err := c.cache.GetCtx(ctx, getNoteCacheKey(nid))
	if err != nil {
		return nil, err
	}

	if res == "" {
		return nil, nil
	}

	var ret NotePO
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

// SetNote 设置笔记缓存
func (c *NoteCache) SetNote(ctx context.Context, model *NotePO) error {
	if c.cache == nil {
		return nil
	}

	content, err := json.Marshal(model)
	if err != nil {
		return err
	}

	ttl := xtime.Week + xtime.JitterDuration(2*xtime.Hour)
	err = c.cache.SetexCtx(ctx, getNoteCacheKey(model.Id), utils.Bytes2String(content), int(ttl))
	return err
}

// DelNote 删除笔记缓存
func (c *NoteCache) DelNote(ctx context.Context, nid int64) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, getNoteCacheKey(nid))
	return err
}

// MGetNotes 批量获取笔记缓存
func (c *NoteCache) MGetNotes(ctx context.Context, nids []int64) (map[int64]*NotePO, error) {
	if c.cache == nil || len(nids) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(nids))
	keyToId := make(map[string]int64, len(nids))
	for _, nid := range nids {
		key := getNoteCacheKey(nid)
		keys = append(keys, key)
		keyToId[key] = nid
	}

	vals, err := c.cache.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*NotePO)
	for i, val := range vals {
		if val == "" {
			continue
		}

		var note NotePO
		if err := json.Unmarshal(utils.StringToBytes(val), &note); err == nil {
			result[keyToId[keys[i]]] = &note
		}
	}

	return result, nil
}

// MSetNotes 批量设置笔记缓存
func (c *NoteCache) MSetNotes(ctx context.Context, notes []*NotePO) error {
	if c.cache == nil || len(notes) == 0 {
		return nil
	}

	for _, note := range notes {
		if err := c.SetNote(ctx, note); err != nil {
			return err
		}
	}

	return nil
}

// DelKeys 删除指定的缓存key
func (c *NoteCache) DelKeys(ctx context.Context, keys ...string) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, keys...)
	return err
}

// GetNoteCountByOwner 获取用户笔记数量缓存
func (c *NoteCache) GetNoteCountByOwner(ctx context.Context, uid int64) (int64, bool, error) {
	if c.cache == nil {
		return 0, false, nil
	}

	res, err := c.cache.GetCtx(ctx, getNoteCountByOwnerCacheKey(uid))
	if err != nil {
		return 0, false, err
	}

	if res == "" {
		return 0, false, nil
	}

	cnt, _ := strconv.ParseInt(res, 10, 64)
	return cnt, true, nil
}

// SetNoteCountByOwner 设置用户笔记数量缓存
func (c *NoteCache) SetNoteCountByOwner(ctx context.Context, uid int64, count int64) error {
	if c.cache == nil {
		return nil
	}

	ttl := xtime.WeekJitterSec(2 * xtime.Hour)
	return c.cache.SetexCtx(ctx, getNoteCountByOwnerCacheKey(uid), xconv.FormatInt(count), ttl)
}

// DelNoteCountByOwner 删除用户笔记数量缓存
func (c *NoteCache) DelNoteCountByOwner(ctx context.Context, uid int64) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, getNoteCountByOwnerCacheKey(uid))
	return err
}

// GetPublicNoteCountByOwner 获取用户公开笔记数量缓存
func (c *NoteCache) GetPublicNoteCountByOwner(ctx context.Context, uid int64) (int64, bool, error) {
	if c.cache == nil {
		return 0, false, nil
	}

	res, err := c.cache.GetCtx(ctx, getNotePublicCountByOwnerCacheKey(uid))
	if err != nil {
		return 0, false, err
	}

	if res == "" {
		return 0, false, nil
	}

	cnt, _ := strconv.ParseInt(res, 10, 64)
	return cnt, true, nil
}

// SetPublicNoteCountByOwner 设置用户公开笔记数量缓存
func (c *NoteCache) SetPublicNoteCountByOwner(ctx context.Context, uid int64, count int64) error {
	if c.cache == nil {
		return nil
	}

	ttl := xtime.WeekJitterSec(2 * xtime.Hour)
	return c.cache.SetexCtx(ctx, getNotePublicCountByOwnerCacheKey(uid), xconv.FormatInt(count), ttl)
}

// DelPublicNoteCountByOwner 删除用户公开笔记数量缓存
func (c *NoteCache) DelPublicNoteCountByOwner(ctx context.Context, uid int64) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, getNotePublicCountByOwnerCacheKey(uid))
	return err
}

// DelNoteRelatedCache 删除笔记相关的缓存（笔记缓存 + 用户笔记数量缓存）
func (c *NoteCache) DelNoteRelatedCache(ctx context.Context, noteId int64, ownerId int64) error {
	if c.cache == nil {
		return nil
	}

	keys := []string{
		getNoteCacheKey(noteId),
		getNoteCountByOwnerCacheKey(ownerId),
		getNotePublicCountByOwnerCacheKey(ownerId),
	}

	_, err := c.cache.DelCtx(ctx, keys...)
	return err
}
