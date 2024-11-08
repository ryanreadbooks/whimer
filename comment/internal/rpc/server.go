package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"

	"github.com/bufbuild/protovalidate-go"
)

type ReplyServiceServer struct {
	commentv1.UnimplementedReplyServiceServer
	validator *protovalidate.Validator

	Svc *svc.ServiceContext
}

func NewReplyServiceServer(ctx *svc.ServiceContext) *ReplyServiceServer {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	return &ReplyServiceServer{
		Svc:       ctx,
		validator: validator,
	}
}

// 发布评论
func (s *ReplyServiceServer) AddReply(ctx context.Context, in *commentv1.AddReplyReq) (*commentv1.AddReplyRes, error) {
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
func (s *ReplyServiceServer) DelReply(ctx context.Context, in *commentv1.DelReplyReq) (*commentv1.DelReplyRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyDel(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}

	return &commentv1.DelReplyRes{}, nil
}

func (s *ReplyServiceServer) LikeAction(ctx context.Context, in *commentv1.LikeActionReq) (*commentv1.LikeActionRes, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSvc.ReplyLike(ctx, in.ReplyId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.LikeActionRes{}, nil
}

func (s *ReplyServiceServer) DislikeAction(ctx context.Context, in *commentv1.DislikeActionReq) (*commentv1.DislikeActionRes, error) {
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
func (s *ReplyServiceServer) ReportReply(ctx context.Context,
	in *commentv1.ReportReplyReq) (*commentv1.ReportReplyRes, error) {
	return &commentv1.ReportReplyRes{}, nil
}

// 置顶
func (s *ReplyServiceServer) PinReply(ctx context.Context,
	in *commentv1.PinReplyReq) (*commentv1.PinReplyRes, error) {
	err := s.Svc.CommentSvc.ReplyPin(ctx, in.Oid, in.Rid, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.PinReplyRes{}, nil
}

// 分页获取主评论
func (s *ReplyServiceServer) PageGetReply(ctx context.Context,
	in *commentv1.PageGetReplyReq) (*commentv1.PageGetReplyRes, error) {
	return s.Svc.CommentSvc.PageGetReply(ctx, in)
}

// 分页获取子评论
func (s *ReplyServiceServer) PageGetSubReply(ctx context.Context,
	in *commentv1.PageGetSubReplyReq) (*commentv1.PageGetSubReplyRes, error) {
	return s.Svc.CommentSvc.PageGetSubReply(ctx, in)
}

func (s *ReplyServiceServer) PageGetDetailedReply(ctx context.Context,
	in *commentv1.PageGetReplyReq) (*commentv1.PageGetDetailedReplyRes, error) {
	return s.Svc.CommentSvc.PageGetObjectReplies(ctx, in)
}

// 获取置顶评论
func (s *ReplyServiceServer) GetPinnedReply(ctx context.Context,
	in *commentv1.GetPinnedReplyReq) (*commentv1.GetPinnedReplyRes, error) {
	return s.Svc.CommentSvc.GetPinnedReply(ctx, in.Oid)
}

// 获取评论数量
func (s *ReplyServiceServer) CountReply(ctx context.Context,
	in *commentv1.CountReplyReq) (*commentv1.CountReplyRes, error) {
	count, err := s.Svc.CommentSvc.CountReply(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CountReplyRes{NumReply: count}, nil
}

// 获取评论的点赞数量
func (s *ReplyServiceServer) GetReplyLikeCount(ctx context.Context,
	in *commentv1.GetReplyLikeCountReq) (*commentv1.GetReplyLikeCountRes, error) {
	res, err := s.Svc.CommentSvc.GetReplyLikesCount(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetReplyLikeCountRes{
		ReplyId: in.ReplyId,
		Count:   res,
	}, nil
}

// 获取评论的点踩数量
func (s *ReplyServiceServer) GetReplyDislikeCount(ctx context.Context,
	in *commentv1.GetReplyDislikeCountReq) (*commentv1.GetReplyDislikeCountRes, error) {
	res, err := s.Svc.CommentSvc.GetReplyDislikesCount(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetReplyDislikeCountRes{
		ReplyId: in.ReplyId,
		Count:   res,
	}, nil
}

func (s *ReplyServiceServer) CheckUserCommentOnObject(ctx context.Context,
	in *commentv1.CheckUserCommentOnObjectRequest) (
	*commentv1.CheckUserCommentOnObjectResponse, error) {
	ok, err := s.Svc.CommentSvc.CheckUserCommentOnObject(ctx, in.Uid, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CheckUserCommentOnObjectResponse{Commented: ok}, nil
}
