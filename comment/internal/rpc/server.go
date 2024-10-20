package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"

	"github.com/bufbuild/protovalidate-go"
)

type ReplyServer struct {
	commentv1.UnimplementedReplyServiceServer
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
func (s *ReplyServer) AddReply(ctx context.Context, in *commentv1.AddReplyReq) (*commentv1.AddReplyRes, error) {
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

	return &commentv1.AddReplyRes{
		ReplyId: res.ReplyId,
	}, nil
}

// 删除评论
func (s *ReplyServer) DelReply(ctx context.Context, in *commentv1.DelReplyReq) (*commentv1.DelReplyRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyDel(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}

	return &commentv1.DelReplyRes{}, nil
}

func (s *ReplyServer) LikeAction(ctx context.Context, in *commentv1.LikeActionReq) (*commentv1.LikeActionRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyLike(ctx, in.ReplyId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.LikeActionRes{}, nil
}

func (s *ReplyServer) DislikeAction(ctx context.Context, in *commentv1.DislikeActionReq) (*commentv1.DislikeActionRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyDislike(ctx, in.ReplyId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.DislikeActionRes{}, nil
}

// TODO 举报
func (s *ReplyServer) ReportReply(ctx context.Context,
	in *commentv1.ReportReplyReq) (*commentv1.ReportReplyRes, error) {
	return &commentv1.ReportReplyRes{}, nil
}

// 置顶
func (s *ReplyServer) PinReply(ctx context.Context,
	in *commentv1.PinReplyReq) (*commentv1.PinReplyRes, error) {
	err := s.Svc.CommentSvc.ReplyPin(ctx, in.Oid, in.Rid, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.PinReplyRes{}, nil
}

// 分页获取主评论
func (s *ReplyServer) PageGetReply(ctx context.Context,
	in *commentv1.PageGetReplyReq) (*commentv1.PageGetReplyRes, error) {
	return s.Svc.CommentSvc.PageGetReply(ctx, in)
}

// 分页获取子评论
func (s *ReplyServer) PageGetSubReply(ctx context.Context,
	in *commentv1.PageGetSubReplyReq) (*commentv1.PageGetSubReplyRes, error) {
	return s.Svc.CommentSvc.PageGetSubReply(ctx, in)
}

func (s *ReplyServer) PageGetDetailedReply(ctx context.Context,
	in *commentv1.PageGetReplyReq) (*commentv1.PageGetDetailedReplyRes, error) {
	return s.Svc.CommentSvc.PageGetObjectReplies(ctx, in)
}

// 获取置顶评论
func (s *ReplyServer) GetPinnedReply(ctx context.Context,
	in *commentv1.GetPinnedReplyReq) (*commentv1.GetPinnedReplyRes, error) {
	return s.Svc.CommentSvc.GetPinnedReply(ctx, in.Oid)
}

// 获取评论数量
func (s *ReplyServer) CountReply(ctx context.Context,
	in *commentv1.CountReplyReq) (*commentv1.CountReplyRes, error) {
	count, err := s.Svc.CommentSvc.CountReply(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CountReplyRes{NumReply: count}, nil
}

// 获取评论的点赞数量
func (s *ReplyServer) GetReplyLikeCount(ctx context.Context,
	in *commentv1.GetReplyLikeCountReq) (*commentv1.GetReplyLikeCountRes, error) {
	s.Svc.CommentSvc.GetReplyLikesCount(ctx, in.ReplyId)
	return &commentv1.GetReplyLikeCountRes{}, nil
}
