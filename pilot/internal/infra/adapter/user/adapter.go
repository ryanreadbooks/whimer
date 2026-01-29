package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/user/convert"

	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
)

type UserAdapter struct {
	userServiceAdapter userv1.UserServiceClient
}

var _ repository.UserServiceAdapter = (*UserAdapter)(nil)

func NewUserAdapter(userServiceAdapter userv1.UserServiceClient) *UserAdapter {
	return &UserAdapter{
		userServiceAdapter: userServiceAdapter,
	}
}

func (a *UserAdapter) GetUser(ctx context.Context, uid int64) (*uservo.User, error) {
	resp, err := a.userServiceAdapter.GetUser(ctx,
		&userv1.GetUserRequest{Uid: uid})
	if err != nil {
		return nil, err
	}

	return convert.PbUserInfoToVoUser(resp.GetUser()), nil
}

func (a *UserAdapter) BatchGetUser(ctx context.Context, uids []int64) (map[int64]*uservo.User, error) {
	resp, err := a.userServiceAdapter.BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{Uids: uids})
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*uservo.User)
	for uid, user := range resp.GetUsers() {
		result[uid] = convert.PbUserInfoToVoUser(user)
	}

	return result, nil
}
