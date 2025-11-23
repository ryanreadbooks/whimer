package user

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/database/usersetting"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	v1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"golang.org/x/sync/errgroup"
)

func (b *Biz) GetSettings(ctx context.Context) (*model.UserSettings, error) {
	var (
		uid       = metadata.Uid(ctx)
		settingPo = &usersetting.UserSettingPO{} // 默认全0

		result           model.UserSettings
		relationSettings *v1.GetUserSettingsResponse
	)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(recovery.DoV2(func() error {
		var err error
		relationSettings, err = dep.RelationServer().GetUserSettings(ctx, &v1.GetUserSettingsRequest{
			Uid: uid,
		})
		if err != nil {
			return xerror.Wrapf(err, "failed to get relation settings").WithExtras("uid", uid).WithCtx(ctx)
		}

		return nil
	}))

	eg.Go(recovery.DoV2(func() error {
		// local user setting from dao
		var err error
		tmp, err := dao.Database().UserSettingDao.GetByUid(ctx, uid, false)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				// 报错不用退出
				err = xerror.Wrapf(err, "failed to get local user setting").WithExtras("uid", uid).WithCtx(ctx)
				xlog.Msg("user setting dao get by uid failed").Err(err).Errorx(ctx)
			}
		}

		if tmp != nil {
			settingPo = tmp
		}

		return nil
	}))

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	result.ShowFanList = relationSettings.ShowFanList
	result.ShowFollowList = relationSettings.ShowFollowList
	result.ShowNoteLikes = model.ShouldShowNoteLikes(settingPo.Flags)

	return &result, nil
}

func (b *Biz) GetIntegralUserSettings(ctx context.Context, uid int64) (*model.IntegralUserSetting, error) {
	settingPo, err := dao.Database().UserSettingDao.GetByUid(ctx, uid, false)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			// 报错不用退出
			err = xerror.Wrapf(err, "failed to get local user setting").WithExtras("uid", uid).WithCtx(ctx)
			xlog.Msg("user setting dao get by uid failed").Err(err).Errorx(ctx)
		}
	}

	if settingPo != nil {
		flags := settingPo.Flags
		return &model.IntegralUserSetting{
			IntegralNoteShowSetting: &model.IntegralNoteShowSetting{
				ShowNoteLikes: model.ShouldShowNoteLikes(flags),
			},
		}, nil
	}

	return &model.IntegralUserSetting{}, nil
}

func (b *Biz) SetNoteShowSettings(ctx context.Context, uid int64, settings *model.IntegralNoteShowSetting) error {
	now := time.Now().Unix()

	// 简单起见 直接加锁改
	err := dao.Database().Transact(ctx, func(ctx context.Context) error {
		settingPo, err := dao.Database().UserSettingDao.GetByUid(ctx, uid, true)
		if err != nil {
			if !xsql.IsNoRecord(err) {
				return xerror.Wrap(err)
			}
		}

		var (
			curFlag int64
			ext     json.RawMessage
		)

		if settingPo != nil {
			ext = settingPo.Ext
			curFlag = settingPo.Flags
		}

		newFlag := model.UpdateFlagsBit(curFlag, model.ShowNoteLikesSettingMask, settings.ShowNoteLikes)

		newSetting := &usersetting.UserSettingPO{
			Uid:   uid,
			Ctime: now,
			Utime: now,
			Ext:   ext,
			Flags: newFlag,
		}

		err = dao.Database().UserSettingDao.Upsert(ctx, newSetting)
		if err != nil {
			return xerror.Wrap(err)
		}

		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "tx set user setting failed").WithCtx(ctx)
	}

	return nil
}
