package dao

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// TODO 关注显式设置功能暂不实现

const (
	ShowFollowings    = 0 // 展示关注列表
	NotShowFollowings = 1 // 不展示关注列表

	NotShowFans = 0 // 不展示粉丝列表
	ShowFans    = 1 // 展示粉丝列表
)

// 用户的关系设置
type RelationSetting struct {
	Uid               uint64 `db:"uid"`
	NotShowFollowings int8   `db:"not_show_followings"` // 是否不展示关注的人 默认展示
	ShowFans          int8   `db:"show_fans"`           // 是否展示粉丝 默认不展示
	Ctime             int64  `db:"ctime"`
	Mtime             int64  `db:"mtime"`
}

type RelationSettingDao struct {
	db           *xsql.DB
	settingCache *xcache.Cache[*RelationSetting]
}

func NewRelationSettingDao(db *xsql.DB, c *redis.Redis) *RelationSettingDao {
	return &RelationSettingDao{
		db:           db,
		settingCache: xcache.New[*RelationSetting](c),
	}
}

func getSettingCacheKey(uid uint64) string {
	return "relation:setting:uid:" + strconv.FormatUint(uid, 10)
}

// all sqls here
const (
	settingFields = "uid,not_show_followings,show_fans,ctime,mtime"
)

var (
	sqlGetSettingByUid = fmt.Sprintf("SELECT %s FROM relation_setting WHERE uid=?", settingFields)
	sqlUpdateSetting   = "UPDATE relation_setting SET not_show_followings=?, show_fans=?, mtime=? WHERE uid=?"
	sqlInsertSetting   = fmt.Sprintf("INSERT INTO relation_setting(%s) VALUES(?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE not_show_followings=val.not_show_followings, show_fans=val.show_fans, mtime=val.mtime", settingFields)
)

func (d *RelationSettingDao) Get(ctx context.Context, uid uint64) (*RelationSetting, error) {
	return d.settingCache.Get(ctx, getSettingCacheKey(uid), xcache.WithGetFallback(
		func(ctx context.Context) (*RelationSetting, int, error) {
			r, err := d.getByUidDao(ctx, uid)
			if err != nil {
				if errors.Is(err, xsql.ErrNoRecord) {
					return &RelationSetting{
						Uid:               uid,
						NotShowFollowings: ShowFollowings,
						ShowFans:          NotShowFans,
					}, 0, nil
				}
				return nil, 0, err
			}

			return r, xtime.WeekJitterSec(time.Minute * 15), nil
		}))
}

func (d *RelationSettingDao) getByUidDao(ctx context.Context, uid uint64) (*RelationSetting, error) {
	var setting RelationSetting
	err := d.db.QueryRowCtx(ctx, &setting, sqlGetSettingByUid, uid)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &setting, nil
}

func (d *RelationSettingDao) Update(ctx context.Context, s *RelationSetting) error {
	if s.Mtime == 0 {
		s.Mtime = time.Now().Unix()
	}
	_, err := d.db.ExecCtx(ctx, sqlUpdateSetting, s.NotShowFollowings, s.ShowFans, s.Mtime, s.Uid)

	concurrent.SafeGo(func() {
		ctx2 := context.WithoutCancel(ctx)
		d.delCache(ctx2, s.Uid)
	})

	return xsql.ConvertError(err)
}

func (d *RelationSettingDao) Insert(ctx context.Context, s *RelationSetting) error {
	if s.Ctime == 0 {
		s.Ctime = time.Now().Unix()
		s.Mtime = s.Ctime
	}

	_, err := d.db.ExecCtx(ctx, sqlInsertSetting, s.Uid, s.NotShowFollowings, s.ShowFans, s.Ctime, s.Mtime)

	concurrent.SafeGo(func() {
		ctx2 := context.WithoutCancel(ctx)
		d.delCache(ctx2, s.Uid)
	})

	return xsql.ConvertError(err)
}

func (d *RelationSettingDao) delCache(ctx context.Context, uid uint64) {
	if _, err := d.settingCache.Del(ctx, getSettingCacheKey(uid)); err != nil {
		xlog.Msg("relation setting dao failed to del cache when inserting").Extra("uid", uid).Errorx(ctx)
	}
}
