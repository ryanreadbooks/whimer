package grpc

import (
	"context"
	"strconv"

	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"
	user "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
)

const (
	maxBatchGetUserAllowed = 800 // 单次批量获取用户信息最大数目
)

type UserServiceServer struct {
	user.UnimplementedUserServiceServer
	Svc *srv.Service
}

func NewUserServiceServer(s *srv.Service) *UserServiceServer {
	return &UserServiceServer{
		Svc: s,
	}
}

func (s *UserServiceServer) BatchGetUser(ctx context.Context, in *user.BatchGetUserRequest) (*user.BatchGetUserResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if len(in.Uids) > maxBatchGetUserAllowed {
		return nil, global.ErrArgs.Msg("数量太大")
	}

	if len(in.Uids) == 0 {
		return &user.BatchGetUserResponse{}, nil
	}

	resp, err := s.Svc.UserSrv.BatchGetUser(ctx, in.Uids)
	if err != nil {
		return nil, err
	}

	users := make(map[string]*user.UserInfo, len(resp))
	for _, r := range resp {
		users[strconv.FormatUint(r.Uid, 10)] = r.ToPb()
	}

	return &user.BatchGetUserResponse{Users: users}, nil
}

func (s *UserServiceServer) GetUser(ctx context.Context, in *user.GetUserRequest) (*user.GetUserResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	resp, err := s.Svc.UserSrv.GetUser(ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	return &user.GetUserResponse{User: resp.ToPb()}, nil
}
