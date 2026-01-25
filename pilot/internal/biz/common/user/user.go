package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/recentcontact"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"

	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
)

type Biz struct {
	recentContact *recentcontact.Store
}

func NewBiz(c *config.Config) *Biz {
	return &Biz{
		recentContact: recentcontact.New(infra.Cache()),
	}
}

func (b *Biz) ListUsersV2(ctx context.Context, uids []int64) (map[int64]*userv1.UserInfo, error) {
	resp, err := dep.Userer().BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{
		Uids: uids,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUsers(), nil
}
