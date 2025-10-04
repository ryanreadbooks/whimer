package grpc

import (
	"context"
	"errors"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/srv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer

	Svc *srv.Service
}

func NewCommentServiceServer(ctx *srv.Service) *CommentServiceServer {
	return &CommentServiceServer{
		Svc: ctx,
	}
}

// 发布评论
func (s *CommentServiceServer) AddComment(ctx context.Context, in *commentv1.AddCommentRequest) (*commentv1.AddCommentResponse, error) {
	req := &model.AddCommentReq{
		Type:     model.CommentType(in.GetType()),
		Oid:      in.GetOid(),
		RootId:   in.GetRootId(),
		ParentId: in.GetParentId(),
		Content:  in.GetContent(),
		ReplyUid: in.GetReplyUid(),
		Images:   in.GetImages(),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	res, err := s.Svc.CommentSrv.AddComment(ctx, req)
	if err != nil {
		return nil, err
	}

	return &commentv1.AddCommentResponse{
		CommentId: res.CommentId,
	}, nil
}

// 删除评论
func (s *CommentServiceServer) DelComment(ctx context.Context, in *commentv1.DelCommentRequest) (*commentv1.DelCommentResponse, error) {
	if in.CommentId <= 0 {
		return nil, global.ErrInvalidCommentId
	}
	if in.Oid <= 0 {
		return nil, global.ErrObjectIdEmpty
	}

	err := s.Svc.CommentSrv.DelComment(ctx, in.Oid, in.CommentId)
	if err != nil {
		return nil, err
	}

	return &commentv1.DelCommentResponse{}, nil
}

func (s *CommentServiceServer) LikeAction(ctx context.Context, in *commentv1.LikeActionRequest) (*commentv1.LikeActionResponse, error) {
	if in.CommentId <= 0 {
		return nil, global.ErrInvalidCommentId
	}

	if in.Action != commentv1.CommentAction_REPLY_ACTION_DO &&
		in.Action != commentv1.CommentAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.LikeComment(ctx, in.CommentId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.LikeActionResponse{}, nil
}

func (s *CommentServiceServer) DislikeAction(ctx context.Context, in *commentv1.DislikeActionRequest) (*commentv1.DislikeActionResponse, error) {
	if in.CommentId <= 0 {
		return nil, global.ErrInvalidCommentId
	}

	if in.Action != commentv1.CommentAction_REPLY_ACTION_DO &&
		in.Action != commentv1.CommentAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.DislikeComment(ctx, in.CommentId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.DislikeActionResponse{}, nil
}

// TODO 举报
func (s *CommentServiceServer) ReportComment(ctx context.Context,
	in *commentv1.ReportCommentRequest) (*commentv1.ReportCommentResponse, error) {
	return &commentv1.ReportCommentResponse{}, nil
}

// 置顶
func (s *CommentServiceServer) PinComment(ctx context.Context,
	in *commentv1.PinCommentRequest) (*commentv1.PinCommentResponse, error) {
	if in.Action != commentv1.CommentAction_REPLY_ACTION_DO &&
		in.Action != commentv1.CommentAction_REPLY_ACTION_UNDO {
		return nil, global.ErrUnsupportedAction
	}

	err := s.Svc.CommentSrv.PinComment(ctx, in.Oid, in.CommentId, int8(in.Action))
	if err != nil {
		return nil, err
	}

	return &commentv1.PinCommentResponse{}, nil
}

// 分页获取主评论
func (s *CommentServiceServer) PageGetComment(ctx context.Context,
	in *commentv1.PageGetCommentRequest) (*commentv1.PageGetCommentResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetRootComments(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetCommentResponse{
		Comments:   model.ItemsAsPbs(resp.Items),
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

// 分页获取子评论
func (s *CommentServiceServer) PageGetSubComment(ctx context.Context,
	in *commentv1.PageGetSubCommentRequest) (*commentv1.PageGetSubCommentResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetSubComments(ctx, in.Oid, in.RootId, in.Cursor)
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetSubCommentResponse{
		Comments:   model.ItemsAsPbs(resp.Items),
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

func (s *CommentServiceServer) PageGetSubCommentV2(ctx context.Context,
	in *commentv1.PageGetSubCommentV2Request) (*commentv1.PageGetSubCommentV2Response, error) {
	if in.Page <= 0 {
		in.Page = 1
	}

	if in.Count >= 10 {
		in.Count = 10
	}

	resp, total, err := s.Svc.CommentSrv.PageListSubComments(ctx, in.Oid, in.RootId, int(in.Page), int(in.Count))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetSubCommentV2Response{
		Comments: model.ItemsAsPbs(resp),
		Total:    total,
	}, nil
}

func (s *CommentServiceServer) PageGetDetailedComment(ctx context.Context,
	in *commentv1.PageGetDetailedCommentRequest) (*commentv1.PageGetDetailedCommentResponse, error) {
	resp, err := s.Svc.CommentSrv.PageGetObjectComments(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetDetailedCommentResponse{
		Comments:   model.DetailedItemsAsPbs(resp.Items),
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (s *CommentServiceServer) PageGetDetailedCommentV2(ctx context.Context,
	in *commentv1.PageGetDetailedCommentV2Request) (*commentv1.PageGetDetailedCommentV2Response, error) {
	resp, err := s.Svc.CommentSrv.PageGetObjectCommentsV2(ctx, in.Oid, in.Cursor, int8(in.SortBy))
	if err != nil {
		return nil, err
	}

	return &commentv1.PageGetDetailedCommentV2Response{
		RootComments: model.DetailedItemsV2AsPbs(resp.Items),
		NextCursor:   resp.NextCursor,
		HasNext:      resp.HasNext,
	}, nil
}

// 获取置顶评论
func (s *CommentServiceServer) GetPinnedComment(ctx context.Context,
	in *commentv1.GetPinnedCommentRequest) (*commentv1.GetPinnedCommentResponse, error) {
	resp, err := s.Svc.CommentSrv.GetPinnedComment(ctx, in.Oid)
	if err != nil {
		if errors.Is(err, global.ErrNoPinComment) {
			return &commentv1.GetPinnedCommentResponse{}, nil
		}

		return nil, err
	}

	return &commentv1.GetPinnedCommentResponse{
		Item: resp.AsPb(),
	}, nil
}

// 获取评论数量
func (s *CommentServiceServer) CountComment(ctx context.Context,
	in *commentv1.CountCommentRequest) (*commentv1.CountCommentResponse, error) {
	count, err := s.Svc.CommentSrv.GetCommentCount(ctx, in.Oid)
	if err != nil {
		return nil, err
	}

	return &commentv1.CountCommentResponse{Count: count}, nil
}

// 获取评论的点赞数量
func (s *CommentServiceServer) GetCommentLikeCount(ctx context.Context,
	in *commentv1.GetCommentLikeCountRequest) (*commentv1.GetCommentLikeCountResponse, error) {
	res, err := s.Svc.CommentSrv.GetCommentLikesCount(ctx, in.CommentId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetCommentLikeCountResponse{
		CommentId: in.CommentId,
		Count:     res,
	}, nil
}

// 获取评论的点踩数量
func (s *CommentServiceServer) GetCommentDislikeCount(ctx context.Context,
	in *commentv1.GetCommentDislikeCountRequest) (*commentv1.GetCommentDislikeCountResponse, error) {
	res, err := s.Svc.CommentSrv.GetCommentDislikesCount(ctx, in.CommentId)
	if err != nil {
		return nil, err
	}
	return &commentv1.GetCommentDislikeCountResponse{
		CommentId: in.CommentId,
		Count:     res,
	}, nil
}

func (s *CommentServiceServer) CheckUserOnObject(ctx context.Context,
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
func (s *CommentServiceServer) BatchCountComment(ctx context.Context, in *commentv1.BatchCountCommentRequest) (
	*commentv1.BatchCountCommentResponse, error) {
	oids := in.GetOids()
	resp, err := s.Svc.CommentSrv.BatchGetCountComment(ctx, oids)
	if err != nil {
		return nil, err
	}

	return &commentv1.BatchCountCommentResponse{Numbers: resp}, nil
}

func (s *CommentServiceServer) BatchCheckUserOnObject(ctx context.Context,
	in *commentv1.BatchCheckUserOnObjectRequest) (
	*commentv1.BatchCheckUserOnObjectResponse, error) {

	var uidObjects = make(map[int64][]int64, len(in.Mappings))
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

func (s *CommentServiceServer) BatchCheckUserLikeComment(ctx context.Context,
	in *commentv1.BatchCheckUserLikeCommentRequest) (*commentv1.BatchCheckUserLikeCommentResponse, error) {

	l := len(in.Mappings)
	if l > 50 {
		return nil, global.ErrArgs.Msg("请求参数太多")
	}

	var req = make(map[int64][]int64)
	for uid, ids := range in.Mappings {
		if len(ids.GetIds()) > 50 {
			return nil, global.ErrArgs.Msg("请求子参数太多")
		}
		req[uid] = ids.GetIds()
	}

	resp, err := s.Svc.CommentSrv.BatchCheckUserLikeStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	results := make(map[int64]*commentv1.BatchCheckUserLikeCommentResponse_CommentLikedList, len(resp))
	for uid, statuses := range resp {
		likeStatuses := make([]*commentv1.CommentLiked, 0, len(statuses))
		for _, status := range statuses {
			likeStatuses = append(likeStatuses, &commentv1.CommentLiked{
				CommentId: status.CommentId,
				Liked:     status.Liked,
			})
		}
		results[uid] = &commentv1.BatchCheckUserLikeCommentResponse_CommentLikedList{
			List: likeStatuses,
		}
	}

	return &commentv1.BatchCheckUserLikeCommentResponse{Results: results}, nil
}

// 获取上传图片评论凭证
func (s *CommentServiceServer) UploadCommentImages(ctx context.Context, in *commentv1.UploadCommentImagesRequest) (
	*commentv1.UploadCommentImagesResponse, error) {
	if in.RequestedCount <= 0 {
		return &commentv1.UploadCommentImagesResponse{}, nil
	}

	if in.RequestedCount > model.MaxCommentImageCount {
		return nil, xerror.ErrInvalidArgs.Msg("request count too large")
	}

	auths, err := s.Svc.CommentSrv.GetCommentImagesUploadAuth(ctx, in.RequestedCount)
	if err != nil {
		return nil, err
	}

	return &commentv1.UploadCommentImagesResponse{
		StoreKeys:   auths.ImageIds,
		CurrentTime: auths.CurrentTime,
		ExpireTime:  auths.ExpireTime,
		Token:       auths.Token,
		UploadAddr:  auths.UploadAddr,
	}, nil
}
