package biz

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/relation/internal/model"
)

type RelationSettingBiz struct {
}

func NewRelationSettingBiz() *RelationSettingBiz {
	b := &RelationSettingBiz{}

	return b
}

func (b *RelationSettingBiz) getSetting(ctx context.Context, visitor, target int64) (*dao.Settings, error) {
	setting, err := infra.Dao().RelationSettingDao.Get(ctx, target)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz can not get settings").
			WithExtras("visitor", visitor, "target", target).WithCtx(ctx)
	}

	return setting.ParseSettings(), nil
}

// 是否允许visitor获取target的粉丝列表
func (b *RelationSettingBiz) CanVisitFanList(ctx context.Context, visitor, target int64) error {
	if visitor == target {
		return nil
	}

	st, err := b.getSetting(ctx, visitor, target)
	if err != nil {
		return err
	}

	if !st.DisplayFanList {
		return xerror.Wrap(global.ErrFanListHidden)
	}

	return nil
}

// 是否允许visitor获取target的关注列表
func (b *RelationSettingBiz) CanVisitFollowingList(ctx context.Context, visitor, target int64) error {
	if visitor == target {
		return nil
	}

	st, err := b.getSetting(ctx, visitor, target)
	if err != nil {
		return err
	}

	if !st.DisplayFollowList {
		return xerror.Wrap(global.ErrFollowingListHidden)
	}

	return nil
}

func newSettingsPoFrom(s *model.RelationSettings) *dao.Settings {
	return &dao.Settings{
		DisplayFanList:    s.ShowFanList,
		DisplayFollowList: s.ShowFollowList,
	}
}

func (b *RelationSettingBiz) UpdateSettings(ctx context.Context, uid int64, s *model.RelationSettings) error {
	data, err := json.Marshal(newSettingsPoFrom(s))
	if err != nil {
		return xerror.Wrapf(err, "can not marshal settings").WithExtra("req", s).WithCtx(ctx)
	}

	err = infra.Dao().RelationSettingDao.Insert(ctx, &dao.RelationSetting{
		Uid:      uid,
		Settings: data,
	})
	if err != nil {
		return xerror.Wrapf(err, "can not insert relation settings").WithCtx(ctx)
	}

	return nil
}

func (b *RelationSettingBiz) GetSettings(ctx context.Context, uid int64) (*model.RelationSettings, error) {
	data, err := infra.Dao().RelationSettingDao.Get(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get settings failed").WithExtras("uid", uid).WithCtx(ctx)
	}

	settings := data.ParseSettings()

	return &model.RelationSettings{
		ShowFanList:    settings.DisplayFanList,
		ShowFollowList: settings.DisplayFollowList,
	}, nil
}
