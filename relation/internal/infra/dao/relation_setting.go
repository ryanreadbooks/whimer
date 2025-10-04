package dao

import (
	"context"
	"encoding/json"
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

type Settings struct {
	DisplayFanList    bool `json:"display_fan_list"`    // 公开粉丝列表
	DisplayFollowList bool `json:"display_follow_list"` // 公开关注列表
}

func (s *Settings) Json() json.RawMessage {
	r, _ := json.Marshal(s)
	return r
}

var (
	DefaultSettings = &Settings{
		DisplayFanList:    true,
		DisplayFollowList: true,
	}

	DefaultSettingsJson = DefaultSettings.Json()
)

// 用户的关系设置
type RelationSetting struct {
	Uid      int64           `db:"uid" json:"uid"`
	Settings json.RawMessage `db:"settings" json:"settings"`
	Ctime    int64           `db:"ctime" json:"ctime"`
	Mtime    int64           `db:"mtime" json:"mtime"`
}

func (r *RelationSetting) ParseSettings() *Settings {
	var s Settings
	_ = json.Unmarshal(r.Settings, &s)
	return &s
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

func getSettingCacheKey(uid int64) string {
	return "relation:setting:uid:" + strconv.FormatInt(uid, 10)
}

// all sqls here
const (
	settingFields = "uid,settings,ctime,mtime"
)

var (
	sqlGetSettingByUid = fmt.Sprintf("SELECT %s FROM relation_setting WHERE uid=?", settingFields)
	sqlUpdateSetting   = "UPDATE relation_setting SET settings=?, mtime=? WHERE uid=?"
	sqlInsertSetting   = fmt.Sprintf("INSERT INTO relation_setting(%s) VALUES(?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE settings=val.settings, mtime=val.mtime", settingFields)
	sqlDelete = "DELETE FROM relation_setting WHERE uid=? LIMIT 1"
)

func (d *RelationSettingDao) Get(ctx context.Context, uid int64) (*RelationSetting, error) {
	return d.settingCache.Get(ctx, getSettingCacheKey(uid), xcache.WithGetFallback(
		func(ctx context.Context) (*RelationSetting, int, error) {
			r, err := d.getByUidDao(ctx, uid)
			if err != nil {
				if errors.Is(err, xsql.ErrNoRecord) {
					return &RelationSetting{
						Uid:      uid,
						Settings: DefaultSettingsJson,
					}, 0, nil
				}
				return nil, 0, err
			}

			return r, xtime.WeekJitterSec(time.Minute * 15), nil
		}))
}

func (d *RelationSettingDao) getByUidDao(ctx context.Context, uid int64) (*RelationSetting, error) {
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
	if s.Settings == nil {
		s.Settings = DefaultSettingsJson
	}
	_, err := d.db.ExecCtx(ctx, sqlUpdateSetting, s.Settings, s.Mtime, s.Uid)

	concurrent.SafeGo(func() {
		ctx2 := context.WithoutCancel(ctx)
		d.delCache(ctx2, s.Uid)
	})

	return xsql.ConvertError(err)
}

func (d *RelationSettingDao) Delete(ctx context.Context, uid int64) error {
	_, err := d.db.ExecCtx(ctx, sqlDelete, uid)
	if err != nil {
		return xsql.ConvertError(err)
	}

	concurrent.SafeGo(func() {
		ctx2 := context.WithoutCancel(ctx)
		d.delCache(ctx2, uid)
	})

	return nil
}

func (d *RelationSettingDao) Insert(ctx context.Context, s *RelationSetting) error {
	if s.Ctime == 0 {
		s.Ctime = time.Now().Unix()
		s.Mtime = s.Ctime
	}
	if s.Settings == nil {
		s.Settings = DefaultSettingsJson
	}
	_, err := d.db.ExecCtx(ctx, sqlInsertSetting, s.Uid, s.Settings, s.Ctime, s.Mtime)

	concurrent.SafeGo(func() {
		ctx2 := context.WithoutCancel(ctx)
		d.delCache(ctx2, s.Uid)
	})

	return xsql.ConvertError(err)
}

func (d *RelationSettingDao) delCache(ctx context.Context, uid int64) {
	if _, err := d.settingCache.Del(ctx, getSettingCacheKey(uid)); err != nil {
		xlog.Msg("relation setting dao failed to del cache when inserting").Extra("uid", uid).Errorx(ctx)
	}
}
