package tag

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	tagCacheNameKey = "tag:name:%s"
	tagCacheIdKey   = "tag:id:%d"
)

func getTagCacheNameKey(name string) string {
	return fmt.Sprintf(tagCacheNameKey, name)
}

func getTagCacheIdKey(id int64) string {
	return fmt.Sprintf(tagCacheIdKey, id)
}

// TagCache 标签缓存操作 - 纯缓存操作
type TagCache struct {
	cache *redis.Redis
}

func NewTagCache(cache *redis.Redis) *TagCache {
	return &TagCache{
		cache: cache,
	}
}

// GetByName 根据名称从缓存获取标签
func (c *TagCache) GetByName(ctx context.Context, name string) (*Tag, error) {
	if c.cache == nil {
		return nil, nil
	}

	res, err := c.cache.GetCtx(ctx, getTagCacheNameKey(name))
	if err != nil {
		return nil, err
	}

	if res == "" {
		return nil, nil
	}

	var tag Tag
	if err := json.Unmarshal(utils.StringToBytes(res), &tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// SetByName 根据名称设置标签缓存
func (c *TagCache) SetByName(ctx context.Context, tag *Tag) error {
	if c.cache == nil {
		return nil
	}

	content, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	ttl := xtime.WeekJitterSec(xtime.Hour)
	return c.cache.SetexCtx(ctx, getTagCacheNameKey(tag.Name), utils.Bytes2String(content), ttl)
}

// GetById 根据ID从缓存获取标签
func (c *TagCache) GetById(ctx context.Context, id int64) (*Tag, error) {
	if c.cache == nil {
		return nil, nil
	}

	res, err := c.cache.GetCtx(ctx, getTagCacheIdKey(id))
	if err != nil {
		return nil, err
	}

	if res == "" {
		return nil, nil
	}

	var tag Tag
	if err := json.Unmarshal(utils.StringToBytes(res), &tag); err != nil {
		return nil, err
	}

	return &tag, nil
}

// SetById 根据ID设置标签缓存
func (c *TagCache) SetById(ctx context.Context, tag *Tag) error {
	if c.cache == nil {
		return nil
	}

	content, err := json.Marshal(tag)
	if err != nil {
		return err
	}

	ttl := xtime.WeekJitterSec(xtime.Hour)
	return c.cache.SetexCtx(ctx, getTagCacheIdKey(tag.Id), utils.Bytes2String(content), ttl)
}

// MGetByIds 批量根据ID从缓存获取标签
func (c *TagCache) MGetByIds(ctx context.Context, ids []int64) (map[int64]*Tag, error) {
	if c.cache == nil || len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(ids))
	keyToId := make(map[string]int64, len(ids))
	for _, id := range ids {
		key := getTagCacheIdKey(id)
		keys = append(keys, key)
		keyToId[key] = id
	}

	vals, err := c.cache.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*Tag)
	for i, val := range vals {
		if val == "" {
			continue
		}

		var tag Tag
		if err := json.Unmarshal(utils.StringToBytes(val), &tag); err == nil {
			result[keyToId[keys[i]]] = &tag
		}
	}

	return result, nil
}

// MSetByIds 批量根据ID设置标签缓存
func (c *TagCache) MSetByIds(ctx context.Context, tags []*Tag) error {
	if c.cache == nil || len(tags) == 0 {
		return nil
	}

	for _, tag := range tags {
		if err := c.SetById(ctx, tag); err != nil {
			return err
		}
	}

	return nil
}

// DelById 根据ID删除标签缓存
func (c *TagCache) DelById(ctx context.Context, id int64) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, getTagCacheIdKey(id))
	return err
}

// DelByName 根据名称删除标签缓存
func (c *TagCache) DelByName(ctx context.Context, name string) error {
	if c.cache == nil {
		return nil
	}

	_, err := c.cache.DelCtx(ctx, getTagCacheNameKey(name))
	return err
}

// DelTag 删除标签的所有缓存
func (c *TagCache) DelTag(ctx context.Context, tag *Tag) error {
	if c.cache == nil || tag == nil {
		return nil
	}

	keys := []string{
		getTagCacheIdKey(tag.Id),
		getTagCacheNameKey(tag.Name),
	}

	_, err := c.cache.DelCtx(ctx, keys...)
	return err
}

// GetMissingIds 获取缓存中缺失的ID列表
func (c *TagCache) GetMissingIds(ctx context.Context, ids []int64) (cached map[int64]*Tag, missing []int64, err error) {
	cached, err = c.MGetByIds(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	if cached == nil {
		cached = make(map[int64]*Tag)
	}

	for _, id := range ids {
		if _, ok := cached[id]; !ok {
			missing = append(missing, id)
		}
	}

	return cached, missing, nil
}
