package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dao"
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

	if !st.DisplayFollowingList {
		return xerror.Wrap(global.ErrFollowingListHidden)
	}

	return nil
}
