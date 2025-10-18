package dao

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xstring"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	msgpackv5 "github.com/vmihailenco/msgpack/v5"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// CommentAssetCache
type CommentAssetCache struct {
	cache *redis.Redis
}

func NewCommentAssetCache(cache *redis.Redis) *CommentAssetCache {
	return &CommentAssetCache{
		cache: cache,
	}
}

func (c *CommentAssetCache) GetByCommentId(ctx context.Context, cid int64) ([]*CommentAsset, error) {
	var ret []*CommentAsset
	key := getAssetListCacheKey(cid)

	// 外层限制了数量 此处可以全量拿
	cacheData, err := c.cache.LrangeCtx(ctx, key, 0, -1)
	if err == nil {
		for _, data := range cacheData {
			var item CommentAsset
			err = msgpackv5.Unmarshal(xstring.AsBytes(data), &item)
			if err == nil {
				ret = append(ret, &item)
			}
		}

		if len(ret) != 0 {
			return ret, nil
		}
	}

	return nil, xerror.Wrap(err)
}

func (c *CommentAssetCache) BatchGetByCommentIds(ctx context.Context, cids []int64) (map[int64][]*CommentAsset, error) {
	ret := make(map[int64][]*CommentAsset)
	if len(cids) == 0 {
		return ret, nil
	}

	var cacheAssets = make(map[int64][]*CommentAsset)
	pipe, err := c.cache.TxPipeline()
	cachedCmds := make([]*goredis.StringSliceCmd, 0, len(cids))
	if err == nil {
		for _, cid := range cids {
			key := getAssetListCacheKey(cid)
			cmd := pipe.LRange(ctx, key, 0, -1)
			cachedCmds = append(cachedCmds, cmd)
		}

		_, err = pipe.Exec(ctx)
		if err == nil {
			for idx, cmd := range cachedCmds {
				curCommentId := cids[idx]
				cacheDataList, err := cmd.Result()
				if err != nil {
					continue
				}

				for _, data := range cacheDataList {
					var item CommentAsset
					err = msgpackv5.Unmarshal(xstring.AsBytes(data), &item)
					if err != nil {
						break
					}

					cacheAssets[curCommentId] = append(cacheAssets[curCommentId], &item)
				}
			}
		}
	}

	return cacheAssets, nil
}

func (c *CommentAssetCache) SetByCommentId(ctx context.Context, cid int64, assets []*CommentAsset) error {
	if len(assets) == 0 {
		return nil
	}

	key := getAssetListCacheKey(cid)
	values := make([]any, 0, len(assets))
	for _, r := range assets {
		data, err := msgpackv5.Marshal(r)
		if err != nil {
			continue
		}
		values = append(values, data)
	}

	pipe, err := c.cache.TxPipeline()
	if err != nil {
		return xerror.Wrap(err)
	}

	pipe.LPush(ctx, key, values...)
	pipe.Expire(ctx, key, xtime.DayJitter(time.Minute*30))

	_, err = pipe.Exec(ctx)
	return xerror.Wrap(err)
}

func (c *CommentAssetCache) BatchSetByCommentIds(ctx context.Context, assetsMap map[int64][]*CommentAsset) error {
	if len(assetsMap) == 0 {
		return nil
	}

	pipe, err := c.cache.TxPipeline()
	if err != nil {
		return xerror.Wrap(err)
	}

	for cid, assets := range assetsMap {
		key := getAssetListCacheKey(cid)
		values := make([]any, 0, len(assets))
		for _, item := range assets {
			data, _ := msgpackv5.Marshal(item)
			values = append(values, data)
		}

		pipe.LPush(ctx, key, values...)
		pipe.Expire(ctx, key, xtime.DayJitter(time.Minute*30))
	}

	_, err = pipe.Exec(ctx)
	return xerror.Wrap(err)
}

func (c *CommentAssetCache) DeleteByCommentId(ctx context.Context, cid int64) error {
	key := getAssetListCacheKey(cid)
	_, err := c.cache.DelCtx(ctx, key)
	return xerror.Wrap(err)
}

func (c *CommentAssetCache) BatchDeleteByCommentIds(ctx context.Context, cids []int64) error {
	if len(cids) == 0 {
		return nil
	}

	keys := make([]string, 0, len(cids))
	for _, cid := range cids {
		keys = append(keys, getAssetListCacheKey(cid))
	}

	_, err := c.cache.DelCtx(ctx, keys...)
	return xerror.Wrap(err)
}

func (c *CommentAssetCache) DeleteByCommentIdAsync(ctx context.Context, cid int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "comment.assetdao.cache.delete.",
		Job: func(ctx context.Context) error {
			if err := c.DeleteByCommentId(ctx, cid); err != nil {
				xlog.Msg("comment asset dao del failed").Err(err).
					Extra("comment_id", cid).Errorx(ctx)
			}
			return nil
		},
	})
}

func (c *CommentAssetCache) BatchDeleteByCommentIdsAsync(ctx context.Context, cids []int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "comment.assetdao.cache.batchdelete.",
		Job: func(ctx context.Context) error {
			if err := c.BatchDeleteByCommentIds(ctx, cids); err != nil {
				xlog.Msg("comment asset dao del failed").Err(err).
					Extra("comment_ids", cids).Errorx(ctx)
			}
			return nil
		},
	})
}

func (c *CommentAssetCache) BatchSetByCommentIdsAsync(ctx context.Context, assetsMap map[int64][]*CommentAsset) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "commentasset.batchgetbycids.cache.set",
		Job: func(ctx context.Context) error {
			if err := c.BatchSetByCommentIds(ctx, assetsMap); err != nil {
				xlog.Msg("after comment asset dao exec failed").Err(err).Errorx(ctx)
			}
			return nil
		},
	})
}
