package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	"github.com/ryanreadbooks/whimer/comment/sdk"
)

type ReplyServer struct {
	sdk.UnimplementedReplyServer

	Svc *svc.ServiceContext
}

func NewReplyServer(ctx *svc.ServiceContext) *ReplyServer {
	return &ReplyServer{
		Svc: ctx,
	}
}

// 发布评论
func (s *ReplyServer) AddReply(ctx context.Context, in *sdk.AddReplyReq) (*sdk.AddReplyRes, error) {
	req := &model.ReplyReq{
		Type:     model.ReplyType(in.GetReplyType()),
		Oid:      in.GetOid(),
		RootId:   in.GetRootId(),
		ParentId: in.GetParentId(),
		Content:  in.GetContent(),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Svc.CommentSvc.ReplyAdd(ctx, req)
	if err != nil {
		return nil, err
	}

	return &sdk.AddReplyRes{}, nil
}

// 删除评论
func (s *ReplyServer) DelReply(ctx context.Context, in *sdk.DelReplyReq) (*sdk.DelReplyRes, error) {

	return &sdk.DelReplyRes{}, nil
}

func (s *ReplyServer) LikeAction(ctx context.Context, in *sdk.LikeActionReq) (*sdk.LikeActionRes, error) {

	return &sdk.LikeActionRes{}, nil
}

func (s *ReplyServer) DislikeAction(ctx context.Context, in *sdk.DislikeActionReq) (*sdk.DislikeActionRes, error) {

	return &sdk.DislikeActionRes{}, nil
}

func (s *ReplyServer) ReportReply(ctx context.Context, in *sdk.ReportReplyReq) (*sdk.ReportReplyRes, error) {

	return &sdk.ReportReplyRes{}, nil
}
