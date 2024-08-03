package user

import (
	"context"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	"github.com/ryanreadbooks/whimer/passport/sdk/user"
)

const (
	maxBatchGetUserAllowed = 100 // 单次批量获取用户信息最大数目
)

type UserServer struct {
	user.UnimplementedUserServer
	Svc *svc.ServiceContext
}

func NewUserServer(s *svc.ServiceContext) *UserServer {
	return &UserServer{
		Svc: s,
	}
}

func (s *UserServer) BatchGetUser(ctx context.Context, in *user.BatchGetUserReq) (*user.BatchGetUserRes, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if len(in.Uids) > maxBatchGetUserAllowed {
		return nil, global.ErrArgs.Msg("数量太大")
	}

	resp, err := s.Svc.ProfileSvc.GetByUids(ctx, in.Uids)
	if err != nil {
		return nil, err
	}

	return &user.BatchGetUserRes{Users: resp}, nil
}
