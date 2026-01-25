package repository

import (
	"context"

	vo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

type UserServiceAdapter interface {
	GetUser(ctx context.Context, uid int64) (*vo.User, error)
	BatchGetUser(ctx context.Context, uids []int64) (map[int64]*vo.User, error)
}
