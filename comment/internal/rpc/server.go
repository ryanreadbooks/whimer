package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	sdk "github.com/ryanreadbooks/whimer/comment/sdk/v1"

	"github.com/bufbuild/protovalidate-go"
)

type ReplyServer struct {
	sdk.UnimplementedReplyServer
	validator *protovalidate.Validator

	Svc *svc.ServiceContext
}

func NewReplyServer(ctx *svc.ServiceContext) *ReplyServer {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	return &ReplyServer{
		Svc:       ctx,
		validator: validator,
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
		ReplyUid: in.GetReplyUid(),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	res, err := s.Svc.CommentSvc.ReplyAdd(ctx, req)
	if err != nil {
		return nil, err
	}

	return &sdk.AddReplyRes{
		ReplyId: res.ReplyId,
	}, nil
}

// 删除评论
func (s *ReplyServer) DelReply(ctx context.Context, in *sdk.DelReplyReq) (*sdk.DelReplyRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyDel(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}

	return &sdk.DelReplyRes{}, nil
}

// 点赞评论
func (s *ReplyServer) LikeAction(ctx context.Context, in *sdk.LikeActionReq) (*sdk.LikeActionRes, error) {
	return &sdk.LikeActionRes{}, nil
}

// 点踩
func (s *ReplyServer) DislikeAction(ctx context.Context, in *sdk.DislikeActionReq) (*sdk.DislikeActionRes, error) {
	return &sdk.DislikeActionRes{}, nil
}

// 举报
func (s *ReplyServer) ReportReply(ctx context.Context, in *sdk.ReportReplyReq) (*sdk.ReportReplyRes, error) {
	return &sdk.ReportReplyRes{}, nil
}

// 置顶
func (s *ReplyServer) PinReply(ctx context.Context, in *sdk.PinReplyReq) (*sdk.PinReplyRes, error) {
	err := s.Svc.CommentSvc.ReplyPin(ctx, in.Oid, in.Rid, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &sdk.PinReplyRes{}, nil
}

// 分页获取主评论
func (s *ReplyServer) PageGetReply(ctx context.Context, in *sdk.PageGetReplyReq) (*sdk.PageGetReplyRes, error) {
	return s.Svc.CommentSvc.PageGetReply(ctx, in)
}

// 分页获取子评论
func (s *ReplyServer) PageGetSubReply(ctx context.Context, in *sdk.PageGetSubReplyReq) (*sdk.PageGetSubReplyRes, error) {
	return s.Svc.CommentSvc.PageGetSubReply(ctx, in)
}

func (s *ReplyServer) PageGetDetailedReply(ctx context.Context, in *sdk.PageGetReplyReq) (*sdk.PageGetDetailedReplyRes, error) {
	return s.Svc.CommentSvc.PageGetObjectReplies(ctx, in)
}

// 获取置顶评论
func (s *ReplyServer) GetPinnedReply(ctx context.Context, in *sdk.GetPinnedReplyReq) (*sdk.GetPinnedReplyRes, error) {
	return s.Svc.CommentSvc.GetPinnedReply(ctx, in.Oid)
}

// 获取评论数量
func (s *ReplyServer) CountReply(ctx context.Context, in *sdk.CountReplyReq) (*sdk.CountReplyRes, error) {
	count, err := s.Svc.CommentSvc.CountReply(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &sdk.CountReplyRes{NumReply: count}, nil
}