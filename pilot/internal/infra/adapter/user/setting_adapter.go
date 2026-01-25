package user

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	usersettingdao "github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/database/usersetting"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"golang.org/x/sync/errgroup"
)

var _ repository.UserSettingRepository = (*UserSettingAdapter)(nil)

type UserSettingAdapter struct {
	dao         *usersettingdao.Dao
	relationCli relationv1.RelationServiceClient
}

func NewUserSettingAdapter(
	dao *usersettingdao.Dao,
	relationCli relationv1.RelationServiceClient,
) *UserSettingAdapter {
	return &UserSettingAdapter{
		dao:         dao,
		relationCli: relationCli,
	}
}

// GetLocalSetting 获取本地存储的用户设置
func (a *UserSettingAdapter) GetLocalSetting(ctx context.Context, uid int64) (*entity.UserSetting, error) {
	po, err := a.dao.GetByUid(ctx, uid, false)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return &entity.UserSetting{Uid: uid}, nil
		}
		return nil, xerror.Wrap(err)
	}

	return settingPoToEntity(po), nil
}

// GetLocalSettingForUpdate 获取本地存储的用户设置
func (a *UserSettingAdapter) GetLocalSettingForUpdate(ctx context.Context, uid int64) (
	*entity.UserSetting, error,
) {
	po, err := a.dao.GetByUid(ctx, uid, true)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return &entity.UserSetting{Uid: uid}, nil
		}
		return nil, xerror.Wrap(err)
	}

	return settingPoToEntity(po), nil
}

// UpsertLocalSetting 更新或创建本地用户设置
func (a *UserSettingAdapter) UpsertLocalSetting(ctx context.Context, setting *entity.UserSetting) error {
	po := entityToSettingPo(setting)
	return a.dao.Upsert(ctx, po)
}

// GetRelationSetting 获取关系服务的用户设置
func (a *UserSettingAdapter) GetRelationSetting(ctx context.Context, uid int64) (*vo.RelationSetting, error) {
	resp, err := a.relationCli.GetUserSettings(ctx,
		&relationv1.GetUserSettingsRequest{
			Uid: uid,
		},
	)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &vo.RelationSetting{
		ShowFanList:    resp.GetShowFanList(),
		ShowFollowList: resp.GetShowFollowList(),
	}, nil
}

// GetFullSetting 获取完整用户设置（聚合本地+远程）
func (a *UserSettingAdapter) GetFullSetting(ctx context.Context, uid int64) (*vo.FullUserSetting, error) {
	var (
		localSetting    *entity.UserSetting
		relationSetting *vo.RelationSetting
	)

	eg, ctx := errgroup.WithContext(ctx)

	// 并发获取本地设置
	eg.Go(recovery.DoV2(func() error {
		var err error
		localSetting, err = a.GetLocalSetting(ctx, uid)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				xlog.Msg("get local user setting failed").Extra("uid", uid).Err(err).Errorx(ctx)
			}
			// 本地设置获取失败不影响整体，使用默认值
			localSetting = &entity.UserSetting{Uid: uid}
		}
		return nil
	}))

	// 并发获取远程设置
	eg.Go(recovery.DoV2(func() error {
		var err error
		relationSetting, err = a.GetRelationSetting(ctx, uid)
		if err != nil {
			return xerror.Wrapf(err, "failed to get relation settings").WithExtras("uid", uid).WithCtx(ctx)
		}
		return nil
	}))

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &vo.FullUserSetting{
		ShowNoteLikes:  localSetting.ShouldShowNoteLikes(),
		ShowFanList:    relationSetting.ShowFanList,
		ShowFollowList: relationSetting.ShowFollowList,
	}, nil
}

func settingPoToEntity(po *usersettingdao.UserSettingPO) *entity.UserSetting {
	if po == nil {
		return nil
	}
	return &entity.UserSetting{
		Uid:   po.Uid,
		Flags: po.Flags,
		Ext:   po.Ext,
		Ctime: po.Ctime,
		Utime: po.Utime,
	}
}

func entityToSettingPo(e *entity.UserSetting) *usersettingdao.UserSettingPO {
	if e == nil {
		return nil
	}
	return &usersettingdao.UserSettingPO{
		Uid:   e.Uid,
		Flags: e.Flags,
		Ext:   e.Ext,
		Ctime: e.Ctime,
		Utime: e.Utime,
	}
}
