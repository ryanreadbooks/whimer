package svc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
)

type commentSvcProxy struct {
	proxy *CommentSvc
}

func (s *commentSvcProxy) AddReply(ctx context.Context, data *comm.Model) error {
	return s.proxy.ConsumeAddReplyEvent(ctx, (*queue.AddReplyData)(data))
}

func (s *commentSvcProxy) DelReply(ctx context.Context, rid uint64, reply *comm.Model) error {
	return s.proxy.ConsumeDelReplyEvent(ctx, &queue.DelReplyData{
		ReplyId: rid,
		Reply:   reply,
	})
}

func (s *commentSvcProxy) LikeReply(ctx context.Context, rid, uid uint64) error {
	return s.proxy.ConsumeLikeDislikeEvent(ctx, &queue.BinaryReplyData{
		Uid:     uid,
		ReplyId: rid,
		Action:  queue.ActionDo,
		Type:    queue.LikeType,
	})
}

func (s *commentSvcProxy) UnLikeReply(ctx context.Context, rid, uid uint64) error {
	return s.proxy.ConsumeLikeDislikeEvent(ctx, &queue.BinaryReplyData{
		Uid:     uid,
		ReplyId: rid,
		Action:  queue.ActionUndo,
		Type:    queue.LikeType,
	})
}

func (s *commentSvcProxy) DisLikeReply(ctx context.Context, rid, uid uint64) error {
	return s.proxy.ConsumeLikeDislikeEvent(ctx, &queue.BinaryReplyData{
		Uid:     uid,
		ReplyId: rid,
		Action:  queue.ActionDo,
		Type:    queue.DisLikeType,
	})
}

func (s *commentSvcProxy) UnDisLikeReply(ctx context.Context, rid, uid uint64) error {
	return s.proxy.ConsumeLikeDislikeEvent(ctx, &queue.BinaryReplyData{
		Uid:     uid,
		ReplyId: rid,
		Action:  queue.ActionUndo,
		Type:    queue.DisLikeType,
	})
}

func (s *commentSvcProxy) PinReply(ctx context.Context, oid, rid uint64) error {
	return s.proxy.ConsumePinEvent(ctx, &queue.PinReplyData{
		ReplyId: rid,
		Action:  queue.ActionDo,
		Oid:     oid,
	})
}

func (s *commentSvcProxy) UnPinReply(ctx context.Context, oid, rid uint64) error {
	return s.proxy.ConsumePinEvent(ctx, &queue.PinReplyData{
		ReplyId: rid,
		Action:  queue.ActionUndo,
		Oid:     oid,
	})
}
