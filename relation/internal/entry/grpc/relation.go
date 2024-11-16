package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/srv"
	relationv1 "github.com/ryanreadbooks/whimer/relation/sdk/v1"
)

type RelationServiceServer struct {
	relationv1.UnimplementedRelationServiceServer

	Srv *srv.Service
}

func NewRelationServiceServer(srv *srv.Service) *RelationServiceServer {
	s := &RelationServiceServer{
		Srv: srv,
	}

	return s
}

func (s *RelationServiceServer) FollowUser(ctx context.Context, req *relationv1.FollowUserRequest) (*relationv1.FollowUserResponse, error) {
	var err error
	if req.Action == relationv1.FollowUserRequest_ACTION_FOLLOW {
		err = s.Srv.RelationSrv.FollowUser(ctx, req.Follower, req.Followee)
	} else if req.Action == relationv1.FollowUserRequest_ACTION_UNFOLLOW {
		err = s.Srv.RelationSrv.UnfollowUser(ctx, req.Follower, req.Followee)
	} else {
		err = global.ErrUnSupported
	}

	if err != nil {
		return nil, err
	}
	return &relationv1.FollowUserResponse{}, nil
}

func (s *RelationServiceServer) GetUserFanList(ctx context.Context, req *relationv1.GetUserFanListRequest) (*relationv1.GetUserFanListResponse, error) {
	return nil, nil
}

func (s *RelationServiceServer) GetUserFollowingList(ctx context.Context, req *relationv1.GetUserFollowingListRequest) (
	*relationv1.GetUserFollowingListResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) RemoveUserFan(ctx context.Context, req *relationv1.RemoveUserFanRequest) (
	*relationv1.RemoveUserFanResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFanCount(ctx context.Context, req *relationv1.GetUserFanCountRequest) (
	*relationv1.GetUserFanCountResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFollowingCount(ctx context.Context, req *relationv1.GetUserFollowingCountRequest) (
	*relationv1.GetUserFollowingCountResponse, error) {

	return nil, nil
}
