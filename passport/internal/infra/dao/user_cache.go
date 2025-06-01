package dao

import (
	"context"
	"encoding/json"
	"fmt"

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

func (d *UserDao) CacheGetUserBaseByUid(ctx context.Context, uid int64) (*UserBase, error) {
	return d.cacheGetUserBaseBy(ctx, getCacheUserBaseUidKey(uid))
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
