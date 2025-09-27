package grpc

import (
	"context"
	"errors"
	"strconv"

	user "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"
)

const (
	maxBatchGetUserAllowed = 100 // 单次批量获取用户信息最大数目
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

func (s *UserServiceServer) batchGetUser(ctx context.Context, uids []int64) (map[int64]*model.UserInfo, error) {
	if len(uids) > maxBatchGetUserAllowed {
		return nil, global.ErrArgs.Msg("batch get too many users")
	}

	if len(uids) == 0 {
		return map[int64]*model.UserInfo{}, nil
	}

	resp, err := s.Svc.UserSrv.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *UserServiceServer) BatchGetUser(ctx context.Context, in *user.BatchGetUserRequest) (*user.BatchGetUserResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	resp, err := s.batchGetUser(ctx, in.Uids)
	if err != nil {
		return nil, err
	}

	users := make(map[string]*user.UserInfo, len(resp))
	for _, r := range resp {
		users[strconv.FormatInt(r.Uid, 10)] = r.ToPb()
	}

	return &user.BatchGetUserResponse{Users: users}, nil
}

func (s *UserServiceServer) BatchGetUserV2(ctx context.Context, in *user.BatchGetUserV2Request) (*user.BatchGetUserV2Response, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	resp, err := s.batchGetUser(ctx, in.Uids)
	if err != nil {
		return nil, err
	}

	users := make(map[int64]*user.UserInfo, len(resp))
	for _, r := range resp {
		users[r.Uid] = r.ToPb()
	}

	return &user.BatchGetUserV2Response{Users: users}, nil
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

func (s *UserServiceServer) HasUser(ctx context.Context, in *user.HasUserRequest) (*user.HasUserResponse, error) {
	resp, err := s.Svc.UserSrv.GetUser(ctx, in.Uid)
	if err != nil {
		if errors.Is(err, global.ErrUserNotFound) {
			return &user.HasUserResponse{Has: false}, nil
		}
		return nil, err
	}

	return &user.HasUserResponse{Has: resp.Uid == in.Uid}, nil
}
