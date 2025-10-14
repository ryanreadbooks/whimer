package relation

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/relation/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

type Biz struct {
}

func NewBiz() *Biz { return &Biz{} }

func (b *Biz) FollowOrUnfollow(ctx context.Context, uid int64, req *model.FollowReq) error {
	_, err := dep.RelationServer().FollowUser(ctx, &relationv1.FollowUserRequest{
		Follower: uid,
		Followee: req.Target,
		Action:   relationv1.FollowUserRequest_Action(req.Action),
	})

	if err != nil {
		return xerror.Wrapf(err, "remote relation server follow user failed")
	}

	return err
}

// 检查uid是否关注了target
func (b *Biz) CheckUserFollows(ctx context.Context, uid, target int64) (bool, error) {
	resp, err := dep.RelationServer().BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     uid,
			Targets: []int64{target},
		})

	if err != nil {
		return false, xerror.Wrapf(err, "remote relation server batch check user followed failed")
	}

	return resp.GetStatus()[target], nil
}

func (b *Biz) UpdateRelationSettings(ctx context.Context, uid int64, req *model.UpdateSettingReq) error {
	_, err := dep.RelationServer().UpdateUserSettings(ctx,
		&relationv1.UpdateUserSettingsRequest{
			TargetUid:      uid,
			ShowFanList:    req.ShowFans,
			ShowFollowList: req.ShowFollows,
		})

	if err != nil {
		return xerror.Wrapf(err, "remote relation server update user settings failed")
	}

	return err
}
