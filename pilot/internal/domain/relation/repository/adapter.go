package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
)

type RelationAdapter interface {
	BatchGetFollowStatus(ctx context.Context, uid int64, targets []int64) (map[int64]bool, error)
	FollowUser(ctx context.Context, follower, followee int64, action vo.FollowAction) error
	UpdateSettings(ctx context.Context, uid int64, showFans, showFollows bool) error

	// 检查单个用户关注状态
	CheckFollowed(ctx context.Context, uid, target int64) (bool, error)
	// 获取用户粉丝数量
	GetFanCount(ctx context.Context, uid int64) (int64, error)
	// 获取用户关注数量
	GetFollowingCount(ctx context.Context, uid int64) (int64, error)
	// 分页获取用户粉丝列表
	PageGetFanList(ctx context.Context, uid int64, page, count int32) ([]int64, int64, error)
	// 分页获取用户关注列表
	PageGetFollowingList(ctx context.Context, uid int64, page, count int32) ([]int64, int64, error)
	// 获取用户完整关注列表
	GetUserFollowingList(ctx context.Context, uid int64, offset int64, count int32) (*FollowingListResult, error)
}

type FollowingListResult struct {
	Followings  []int64
	FollowTimes []int64
	HasMore     bool
	NextOffset  int64
}
