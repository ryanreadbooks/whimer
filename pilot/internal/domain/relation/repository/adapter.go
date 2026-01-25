package repository

import (
	"context"
)

type RelationAdapter interface {
	BatchGetFollowStatus(ctx context.Context, uid int64, targets []int64) (map[int64]bool, error)
}
