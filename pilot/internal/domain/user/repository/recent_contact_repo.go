package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

// 最近联系人仓储
type RecentContactRepository interface {
	// 原子追加最近联系人
	Append(ctx context.Context, uid int64, targets []int64) error

	// 获取所有最近联系人
	GetAll(ctx context.Context, uid int64) ([]vo.RecentContact, error)
}
