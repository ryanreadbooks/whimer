package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/passport/internal/biz"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

type UserSrv struct {
	parent *Service

	userBiz biz.UserBiz
}

func NewUserSrv(p *Service, biz biz.Biz) *UserSrv {
	s := &UserSrv{
		parent:  p,
		userBiz: biz.User,
	}

	return s
}

func (s *UserSrv) GetUser(ctx context.Context, uid int64) (*model.UserInfo, error) {
	return s.userBiz.GetUser(ctx, uid)
}

func (s *UserSrv) BatchGetUser(ctx context.Context, uids []int64) (map[int64]*model.UserInfo, error) {
	return s.userBiz.BatchGetUser(ctx, uids)
}

func (s *UserSrv) UpdateUser(ctx context.Context, req *model.UpdateUserRequest) (*model.UserInfo, error) {
	return s.userBiz.UpdateUser(ctx, req)
}

func (s *UserSrv) UpdateUserAvatar(ctx context.Context, req *model.AvatarInfoRequest) (string, error) {
	var (
		user = model.CtxGetUserInfo(ctx)
	)

	ret, _, err := s.userBiz.UpdateAvatar(ctx, user.Uid, req)
	return ret, err
}
