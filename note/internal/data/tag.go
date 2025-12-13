package data

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	tagdao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/tag"
)

// TagData 标签数据层 - 协调数据库和缓存操作
type TagData struct {
	repo  *tagdao.TagRepo
	cache *tagdao.TagCache
}

func NewTagData(repo *tagdao.TagRepo, cache *tagdao.TagCache) *TagData {
	return &TagData{
		repo:  repo,
		cache: cache,
	}
}

// Create 创建标签
func (d *TagData) Create(ctx context.Context, tag *tagdao.Tag) (int64, error) {
	return d.repo.Create(ctx, tag)
}

// FindByName 根据名称获取标签，支持缓存选项
func (d *TagData) FindByName(ctx context.Context, name string, opts ...GetOptionFunc) (*tagdao.Tag, error) {
	opt := ApplyOptions(opts...)

	// 尝试从缓存获取
	if opt.UseCache {
		if cached, err := d.cache.GetByName(ctx, name); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	tag, err := d.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}

	// 异步回填缓存
	if opt.SetCache {
		concurrent.SafeGo(func() {
			if err2 := d.cache.SetByName(context.WithoutCancel(ctx), tag); err2 != nil {
				xlog.Msg("tag data failed to set cache by name").
					Extras("name", name).Err(err2).Errorx(ctx)
			}
		})
	}

	return tag, nil
}

// FindById 根据ID获取标签，支持缓存选项
func (d *TagData) FindById(ctx context.Context, id int64, opts ...GetOptionFunc) (*tagdao.Tag, error) {
	opt := ApplyOptions(opts...)

	// 尝试从缓存获取
	if opt.UseCache {
		if cached, err := d.cache.GetById(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	tag, err := d.repo.FindById(ctx, id)
	if err != nil {
		return nil, err
	}

	// 异步回填缓存
	if opt.SetCache {
		concurrent.SafeGo(func() {
			if err2 := d.cache.SetById(context.WithoutCancel(ctx), tag); err2 != nil {
				xlog.Msg("tag data failed to set cache by id").
					Extras("id", id).Err(err2).Errorx(ctx)
			}
		})
	}

	return tag, nil
}

// BatchGetById 批量根据ID获取标签，支持缓存选项
func (d *TagData) BatchGetById(ctx context.Context, ids []int64, opts ...GetOptionFunc) ([]*tagdao.Tag, error) {
	if len(ids) == 0 {
		return []*tagdao.Tag{}, nil
	}

	opt := ApplyOptions(opts...)
	result := make([]*tagdao.Tag, 0, len(ids))
	idToTag := make(map[int64]*tagdao.Tag, len(ids))

	// 尝试从缓存获取
	var missingIds []int64
	if opt.UseCache {
		cached, missing, err := d.cache.GetMissingIds(ctx, ids)
		if err != nil {
			xlog.Msg("tag data failed to get from cache").Err(err).Infox(ctx)
			missingIds = ids
		} else {
			for id, tag := range cached {
				idToTag[id] = tag
			}
			missingIds = missing
		}
	} else {
		missingIds = ids
	}

	// 从数据库获取缺失的
	if len(missingIds) > 0 {
		tags, err := d.repo.BatchGetById(ctx, missingIds)
		if err != nil {
			return nil, err
		}

		for _, tag := range tags {
			idToTag[tag.Id] = tag
		}

		// 异步回填缓存
		if opt.SetCache && len(tags) > 0 {
			concurrent.SafeGo(func() {
				if err2 := d.cache.MSetByIds(context.WithoutCancel(ctx), tags); err2 != nil {
					xlog.Msg("tag data failed to mset cache").Err(err2).Errorx(ctx)
				}
			})
		}
	}

	// 按原始顺序返回结果
	for _, id := range ids {
		if tag, ok := idToTag[id]; ok {
			result = append(result, tag)
		}
	}

	return result, nil
}

