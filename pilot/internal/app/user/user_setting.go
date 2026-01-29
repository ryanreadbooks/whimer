package user

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/user/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/entity"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

// 获取用户设置
func (s *Service) GetSettings(ctx context.Context) (*dto.UserSettings, error) {
	uid := metadata.Uid(ctx)

	fullSetting, err := s.userSettingRepo.GetFullSetting(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get full setting failed").WithCtx(ctx)
	}

	return &dto.UserSettings{
		ShowFanList:    fullSetting.ShowFanList,
		ShowFollowList: fullSetting.ShowFollowList,
		ShowNoteLikes:  fullSetting.ShowNoteLikes,
	}, nil
}

// 设置笔记展示相关设置
func (s *Service) SetNoteShowSettings(ctx context.Context, uid int64, cmd *dto.SetNoteShowSettingReq) error {
	now := time.Now().Unix()

	setting, err := s.userSettingRepo.GetLocalSettingForUpdate(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "get local setting for update failed").WithCtx(ctx)
	}

	setting.SetShowNoteLikes(cmd.ShowNoteLikes)
	setting.Ctime = now
	setting.Utime = now

	if err := s.userSettingRepo.UpsertLocalSetting(ctx, setting); err != nil {
		return xerror.Wrapf(err, "upsert local setting failed").WithCtx(ctx)
	}

	return nil
}

// 获取整体用户设置
func (s *Service) GetIntegralUserSettings(ctx context.Context, uid int64) (*entity.UserSetting, error) {
	setting, err := s.userSettingRepo.GetLocalSetting(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get local setting failed").WithCtx(ctx)
	}

	return setting, nil
}
