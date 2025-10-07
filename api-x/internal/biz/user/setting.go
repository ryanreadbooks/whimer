package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	v1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

func (b *Biz) GetSettings(ctx context.Context) (*model.UserSettings, error) {
	var (
		uid    = metadata.Uid(ctx)
		result model.UserSettings
	)

	relationSettings, err := infra.RelationServer().GetUserSettings(ctx, &v1.GetUserSettingsRequest{
		Uid: uid,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "failed to get relation settings").WithExtras("uid", uid).WithCtx(ctx)
	}
	result.ShowFanList = relationSettings.ShowFanList
	result.ShowFollowList = relationSettings.ShowFollowList

	return &result, nil
}
