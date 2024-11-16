package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/relation/sdk/v1"
)

type RelationServiceServer struct {
	v1.UnimplementedRelationServiceServer
}

func NewRelationServiceServer() *RelationServiceServer {
	s := &RelationServiceServer{}

	return s
}

func (s *RelationServiceServer) FollowUser(ctx context.Context, req *v1.FollowUserRequest) (*v1.FollowUserResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFanList(ctx context.Context, req *v1.GetUserFanListRequest) (*v1.GetUserFanListResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFollowingList(ctx context.Context, req *v1.GetUserFollowingListRequest) (*v1.GetUserFollowingListResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) RemoveUserFan(ctx context.Context, req *v1.RemoveUserFanRequest) (*v1.RemoveUserFanResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFanCount(ctx context.Context, req *v1.GetUserFanCountRequest) (*v1.GetUserFanCountResponse, error) {

	return nil, nil
}

func (s *RelationServiceServer) GetUserFollowingCount(ctx context.Context, req *v1.GetUserFollowingCountRequest) (*v1.GetUserFollowingCountResponse, error) {

	return nil, nil
}
