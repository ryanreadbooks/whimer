package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xtime"
)

const (
	cacheUserBaseUidKey = "passport:user:base:uid:%d"
	cacheUserBaseTelKey = "passport:user:base:tel:%s"
)

func getCacheUserBaseUidKey(uid int64) string {
	return fmt.Sprintf(cacheUserBaseUidKey, uid)
}

func batchGetCacheUserBaseUidKeys(uids []int64) []string {
	keys := make([]string, 0, len(uids))
	for _, u := range uids {
		keys = append(keys, getCacheUserBaseUidKey(u))
	}
	return keys
}

func (d *UserDao) cacheGetUserBaseBy(ctx context.Context, key string) (*UserBase, error) {
	if d.cache == nil {
		return nil, nil
	}

	res, err := d.cache.GetCtx(ctx, key)
	if err != nil {
		return nil, err
	}

	var ret UserBase
	err = json.Unmarshal(utils.StringToBytes(res), &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (d *UserDao) cacheBatchGetUserBaseBy(ctx context.Context, keys []string) ([]*UserBase, error) {
	if d.cache == nil {
		return nil, nil
	}

	res, err := d.cache.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, err
	}

	ret := make([]*UserBase, len(keys))
	for idx, cache := range res {
		var u UserBase
		err := json.Unmarshal(utils.StringToBytes(cache), &u)
		if err != nil {
			continue
		}
		ret[idx] = &u
	}

	return ret, nil
}

func (d *UserDao) CacheGetUserBaseByUid(ctx context.Context, uid int64) (*UserBase, error) {
	return d.cacheGetUserBaseBy(ctx, getCacheUserBaseUidKey(uid))
}

func (d *UserDao) CacheBatchGetUserBaseByUids(ctx context.Context, uids []int64) ([]*UserBase, error) {
	keys := make([]string, len(uids))
	for idx, uid := range uids {
		keys[idx] = getCacheUserBaseUidKey(uid)
	}

	return d.cacheBatchGetUserBaseBy(ctx, keys)
}

func (d *UserDao) CacheSetUserBase(ctx context.Context, u *UserBase) error {
	if d.cache == nil {
		return nil
	}

	content, err := json.Marshal(u)
	if err != nil {
		return err
	}

	ttl := xtime.Week + xtime.JitterDuration(2*xtime.Hour)
	err = d.cache.SetexCtx(ctx, getCacheUserBaseUidKey(u.Uid), utils.Bytes2String(content), int(ttl))
	if err != nil {
		xlog.Msg("user dao cache failed to set uid key").Extra("uid", u.Uid).Infox(ctx)
	}

	return err
}

func (d *UserDao) CacheBatchSetUserBase(ctx context.Context, ubs []*UserBase) error {
	if len(ubs) == 0 || d.cache == nil {
		return nil
	}

	args := make([]any, 0, len(ubs)*2)
	keys := make([]string, 0, len(ubs))
	for _, u := range ubs {
		val, err := json.Marshal(u)
		if err == nil {
			key := getCacheUserBaseUidKey(u.Uid)
			keys = append(keys, key)
			args = append(args, key, val)
		}
	}

	pipe, err := d.cache.TxPipeline()
	if err != nil {
		return err
	}

	pipe.MSet(ctx, args...)
	for _, key := range keys {
		pipe.Expire(ctx, key, xtime.WeekJitter(time.Minute*30))
	}

	_, err = pipe.Exec(ctx)
	return err
}

func (d *UserDao) CacheDelUserBaseByUid(ctx context.Context, uid int64) error {
	if d.cache == nil {
		return nil
	}

	_, err := d.cache.DelCtx(ctx, getCacheUserBaseUidKey(uid))
	if err != nil {
		xlog.Msg("user dao cache failed to del uid key").Extra("uid", uid).Errorx(ctx)
	}

	return err
}
