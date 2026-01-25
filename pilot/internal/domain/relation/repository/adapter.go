package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
)

type RelationAdapter interface {
	BatchGetFollowStatus(ctx context.Context, uid int64, targets []int64) (map[int64]bool, error)
	FollowUser(ctx context.Context, follower, followee int64, action vo.FollowAction) error
	UpdateSettings(ctx context.Context, uid int64, showFans, showFollows bool) error
}
