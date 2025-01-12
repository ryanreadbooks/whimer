package grpc

import (
	"context"

	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/srv"
)

const (
	maxLimit = 20
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
	if req.Cond.Count < 0 {
		req.Cond.Count = maxLimit
	}
	req.Cond.Count = min(req.Cond.Count, maxLimit)
	fans, res, err := s.Srv.RelationSrv.GetUserFanList(ctx, req.Uid, req.Cond.Offset, int(req.Cond.Count))
	if err != nil {
		return nil, err
	}

	return &relationv1.GetUserFanListResponse{
		Fans:       fans,
		NextOffset: res.NextOffset,
		HasMore:    res.HasMore}, nil
}

func (s *RelationServiceServer) GetUserFollowingList(ctx context.Context, req *relationv1.GetUserFollowingListRequest) (
	*relationv1.GetUserFollowingListResponse, error) {
	if req.Cond.Count < 0 {
		req.Cond.Count = maxLimit
	}
	req.Cond.Count = min(req.Cond.Count, maxLimit)
	followings, res, err := s.Srv.RelationSrv.GetUserFollowingList(ctx, req.Uid, req.Cond.Offset, int(req.Cond.Count))
	if err != nil {
		return nil, err
	}

	return &relationv1.GetUserFollowingListResponse{
		Followings: followings,
		NextOffset: res.NextOffset,
		HasMore:    res.HasMore}, nil
}

func (s *RelationServiceServer) RemoveUserFan(ctx context.Context, req *relationv1.RemoveUserFanRequest) (
	*relationv1.RemoveUserFanResponse, error) {

	return &relationv1.RemoveUserFanResponse{}, nil
}

func (s *RelationServiceServer) GetUserFanCount(ctx context.Context, req *relationv1.GetUserFanCountRequest) (
	*relationv1.GetUserFanCountResponse, error) {
	cnt, err := s.Srv.RelationSrv.GetUserFanCount(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	return &relationv1.GetUserFanCountResponse{Count: cnt}, nil
}

func (s *RelationServiceServer) GetUserFollowingCount(ctx context.Context, req *relationv1.GetUserFollowingCountRequest) (
	*relationv1.GetUserFollowingCountResponse, error) {
	cnt, err := s.Srv.RelationSrv.GetUserFollowingCount(ctx, req.Uid)
	if err != nil {
		return nil, err
	}
	return &relationv1.GetUserFollowingCountResponse{Count: cnt}, nil
}

func (s *RelationServiceServer) BatchCheckUserFollowed(ctx context.Context, req *relationv1.BatchCheckUserFollowedRequest) (
	*relationv1.BatchCheckUserFollowedResponse, error) {
	res, err := s.Srv.RelationSrv.BatchCheckUserFollowStatus(ctx, req.Uid, req.Targets)
	if err != nil {
		return nil, err
	}

	return &relationv1.BatchCheckUserFollowedResponse{Status: res}, nil
}

func (s *RelationServiceServer) CheckUserFollowed(ctx context.Context, req *relationv1.CheckUserFollowedRequest) (
	*relationv1.CheckUserFollowedResponse, error) {
	res, err := s.Srv.RelationSrv.CheckUserFollowStatus(ctx, req.Uid, req.Other)
	if err != nil {
		return nil, err
	}

	return &relationv1.CheckUserFollowedResponse{Followed: res}, nil
}
