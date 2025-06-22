package grpc

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/srv"
)

type ReplyServiceServer struct {
	commentv1.UnimplementedReplyServiceServer

	Svc *srv.Service
}

func NewReplyServiceServer(ctx *srv.Service) *ReplyServiceServer {
	return &ReplyServiceServer{
		Svc: ctx,
	}
}

// 发布评论
func (s *ReplyServiceServer) AddReply(ctx context.Context, in *commentv1.AddReplyRequest) (*commentv1.AddReplyResponse, error) {
	req := &model.AddReplyReq{
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

	res, err := s.Svc.CommentSrv.AddReply(ctx, req)
	if err != nil {
		return nil, err
	}

	return &commentv1.AddReplyResponse{
		ReplyId: res.ReplyId,
	}, nil
}

// 删除评论
func (s *ReplyServiceServer) DelReply(ctx context.Context, in *commentv1.DelReplyRequest) (*commentv1.DelReplyResponse, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	err := s.Svc.CommentSrv.DelReply(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}

	return &commentv1.DelReplyResponse{}, nil
}

func (s *ReplyServiceServer) LikeAction(ctx context.Context, in *commentv1.LikeActionRequest) (*commentv1.LikeActionResponse, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	if in.Action != commentv1.ReplyAction_REPLY_ACTION_DO &&
		in.Action != commentv1.ReplyAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.LikeReply(ctx, in.ReplyId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.LikeActionResponse{}, nil
}

func (s *ReplyServiceServer) DislikeAction(ctx context.Context, in *commentv1.DislikeActionRequest) (*commentv1.DislikeActionResponse, error) {
	if in.ReplyId <= 0 {
		return nil, global.ErrInvalidReplyId
	}

	if in.Action != commentv1.ReplyAction_REPLY_ACTION_DO &&
		in.Action != commentv1.ReplyAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.DislikeReply(ctx, in.ReplyId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.DislikeActionResponse{}, nil
}

// TODO 举报
func (s *ReplyServiceServer) ReportReply(ctx context.Context,
	in *commentv1.ReportReplyRequest) (*commentv1.ReportReplyResponse, error) {
	return &commentv1.ReportReplyResponse{}, nil
}

// 置顶
func (s *ReplyServiceServer) PinReply(ctx context.Context,
	in *commentv1.PinReplyRequest) (*commentv1.PinReplyResponse, error) {
	if in.Action != commentv1.ReplyAction_REPLY_ACTION_DO &&
		in.Action != commentv1.ReplyAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.PinReply(ctx, in.Oid, in.Rid, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.PinReplyResponse{}, nil
}

// 分页获取主评论
func (s *ReplyServiceServer) PageGetReply(ctx context.Context,
	in *commentv1.PageGetReplyRequest) (*commentv1.PageGetReplyResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetRootReplies(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetReplyResponse{
		Replies:    model.ItemsAsPbs(resp.Items),
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

// 分页获取子评论
func (s *ReplyServiceServer) PageGetSubReply(ctx context.Context,
	in *commentv1.PageGetSubReplyRequest) (*commentv1.PageGetSubReplyResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetSubReplies(ctx, in.Oid, in.RootId, in.Cursor)
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetSubReplyResponse{
		Replies:    model.ItemsAsPbs(resp.Items),
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

func (s *ReplyServiceServer) PageGetSubReplyV2(ctx context.Context,
	in *commentv1.PageGetSubReplyV2Request) (*commentv1.PageGetSubReplyV2Response, error) {
	if in.Page <= 0 {
		in.Page = 1
	}

	if in.Count >= 10 {
		in.Count = 10
	}

	resp, total, err := s.Svc.CommentSrv.PageListSubReplies(ctx, in.Oid, in.RootId, int(in.Page), int(in.Count))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetSubReplyV2Response{
		Replies: model.ItemsAsPbs(resp),
		Total:   total,
	}, nil
}

func (s *ReplyServiceServer) PageGetDetailedReply(ctx context.Context,
	in *commentv1.PageGetDetailedReplyRequest) (*commentv1.PageGetDetailedReplyResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetObjectReplies(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetDetailedReplyResponse{
		Replies:    model.DetailedItemsAsPbs(resp.Items),
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (s *ReplyServiceServer) PageGetDetailedReplyV2(ctx context.Context,
	in *commentv1.PageGetDetailedReplyV2Request) (*commentv1.PageGetDetailedReplyV2Response, error) {
	resp, err := s.Svc.CommentSrv.PageGetObjectRepliesV2(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetDetailedReplyV2Response{
		Replies:    model.DetailedItemsV2AsPbs(resp.Items),
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

// 获取置顶评论
func (s *ReplyServiceServer) GetPinnedReply(ctx context.Context,
	in *commentv1.GetPinnedReplyRequest) (*commentv1.GetPinnedReplyResponse, error) {
	resp, err := s.Svc.CommentSrv.GetPinnedReply(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.GetPinnedReplyResponse{
		Reply: resp.AsPb(),
	}, nil
}

// 获取评论数量
func (s *ReplyServiceServer) CountReply(ctx context.Context,
	in *commentv1.CountReplyRequest) (*commentv1.CountReplyResponse, error) {
	count, err := s.Svc.CommentSrv.GetReplyCount(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CountReplyResponse{NumReply: count}, nil
}

// 获取评论的点赞数量
func (s *ReplyServiceServer) GetReplyLikeCount(ctx context.Context,
	in *commentv1.GetReplyLikeCountRequest) (*commentv1.GetReplyLikeCountResponse, error) {
	res, err := s.Svc.CommentSrv.GetReplyLikesCount(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetReplyLikeCountResponse{
		ReplyId: in.ReplyId,
		Count:   res,
	}, nil
}

// 获取评论的点踩数量
func (s *ReplyServiceServer) GetReplyDislikeCount(ctx context.Context,
	in *commentv1.GetReplyDislikeCountRequest) (*commentv1.GetReplyDislikeCountResponse, error) {
	res, err := s.Svc.CommentSrv.GetReplyDislikesCount(ctx, in.ReplyId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetReplyDislikeCountResponse{
		ReplyId: in.ReplyId,
		Count:   res,
	}, nil
}

func (s *ReplyServiceServer) CheckUserOnObject(ctx context.Context,
	in *commentv1.CheckUserOnObjectRequest) (
	*commentv1.CheckUserOnObjectResponse, error) {
	ok, err := s.Svc.CommentSrv.CheckUserIsReplied(ctx, in.Uid, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CheckUserOnObjectResponse{
		Result: &commentv1.OidCommented{
			Oid:       in.Oid,
			Commented: ok,
		},
	}, nil
}

// 获取多个被评论对象的评论数
func (s *ReplyServiceServer) BatchCountReply(ctx context.Context, in *commentv1.BatchCountReplyRequest) (
	*commentv1.BatchCountReplyResponse, error) {
	oids := in.GetOids()
	resp, err := s.Svc.CommentSrv.BatchGetCountReply(ctx, oids)
	if err != nil {
		return nil, err
	}

	return &commentv1.BatchCountReplyResponse{Numbers: resp}, nil
}

func (s *ReplyServiceServer) BatchCheckUserOnObject(ctx context.Context,
	in *commentv1.BatchCheckUserOnObjectRequest) (
	*commentv1.BatchCheckUserOnObjectResponse, error) {

	var uidObjects = make(map[int64][]uint64, len(in.Mappings))
	for uid, m := range in.GetMappings() {
		uidObjects[uid] = append(uidObjects[uid], m.Oids...)
	}
	resp, err := s.Svc.CommentSrv.BatchCheckUserIsReplied(ctx, uidObjects)
	if err != nil {
		return nil, err
	}

	m := make(map[int64]*commentv1.OidCommentedList)
	for _, r := range resp {
		if _, ok := m[r.Uid]; !ok {
			m[r.Uid] = &commentv1.OidCommentedList{}
		}

		m[r.Uid].List = append(m[r.Uid].List, &commentv1.OidCommented{
			Oid:       r.Oid,
			Commented: r.Commented,
		})
	}

	return &commentv1.BatchCheckUserOnObjectResponse{Results: m}, nil
}
