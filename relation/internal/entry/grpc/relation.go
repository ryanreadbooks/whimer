package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/model"
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
	switch req.Action {
	case relationv1.FollowUserRequest_ACTION_FOLLOW:
		err = s.Srv.RelationSrv.FollowUser(ctx, req.Follower, req.Followee)
	case relationv1.FollowUserRequest_ACTION_UNFOLLOW:
		err = s.Srv.RelationSrv.UnfollowUser(ctx, req.Follower, req.Followee)
	default:
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

// 分页获取某个用户的粉丝列表
func (s *RelationServiceServer) PageGetUserFanList(ctx context.Context,
	req *relationv1.PageGetUserFanListRequest) (*relationv1.PageGetUserFanListResponse, error) {
	var resp = &relationv1.PageGetUserFanListResponse{}
	if req.Page <= 0 || req.Count <= 0 {
		return resp, nil
	}

	if req.Count >= 30 {
		req.Count = 30
	}

	fansId, total, err := s.Srv.RelationSrv.PageGetUserFanList(ctx, req.Target, req.Page, req.Count)
	if err != nil {
		return nil, err
	}

	resp.FansId = fansId
	resp.Total = total

	return resp, nil
}

// 分页获取某个用户的关注列表
func (s *RelationServiceServer) PageGetUserFollowingList(ctx context.Context,
	req *relationv1.PageGetUserFollowingListRequest) (*relationv1.PageGetUserFollowingListResponse, error) {
	var resp = &relationv1.PageGetUserFollowingListResponse{}
	if req.Page <= 0 || req.Count <= 0 {
		return resp, nil
	}
	if req.Count >= 30 {
		req.Count = 30
	}

	fansId, total, err := s.Srv.RelationSrv.PageGetUserFollowingList(ctx, req.Target, req.Page, req.Count)
	if err != nil {
		return nil, err
	}

	resp.FollowingsId = fansId
	resp.Total = total

	return resp, nil
}

// 关注设置
func (s *RelationServiceServer) UpdateUserSettings(ctx context.Context, in *relationv1.UpdateUserSettingsRequest) (
	*relationv1.UpdateUserSettingsResponse, error) {
	if in.TargetUid == 0 {
		return &relationv1.UpdateUserSettingsResponse{}, nil
	}

	err := s.Srv.RelationSrv.UpdateUserSettings(ctx, in.TargetUid, &model.RelationSettings{
		ShowFanList:    in.ShowFanList,
		ShowFollowList: in.ShowFollowList,
	})
	if err != nil {
		return nil, err
	}
	return &relationv1.UpdateUserSettingsResponse{}, nil
}

func (s *RelationServiceServer) GetUserSettings(ctx context.Context, in *relationv1.GetUserSettingsRequest) (
	*relationv1.GetUserSettingsResponse, error) {
	if in.Uid == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	resp, err := s.Srv.RelationSrv.GetUserSettings(ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	return &relationv1.GetUserSettingsResponse{
		ShowFanList:    resp.ShowFanList,
		ShowFollowList: resp.ShowFollowList,
	}, nil
}
