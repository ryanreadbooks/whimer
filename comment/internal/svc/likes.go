package svc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/external"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// 设置评论的点赞数量
func (s *CommentSvc) fillReplyLikes(ctx context.Context, replies []*commentv1.ReplyItem) error {
	if len(replies) == 0 {
		return nil
	}

	requests := make([]*counterv1.GetSummaryRequest, 0, 16)
	for _, reply := range replies {
		requests = append(requests, &counterv1.GetSummaryRequest{
			BizCode: global.CounterLikeBizcode,
			Oid:     reply.Id,
		})
	}
	resp, err := external.GetCounter().BatchGetSummary(ctx, &counterv1.BatchGetSummaryRequest{
		Requests: requests,
	})
	if err != nil {
		xlog.Msg("counter batch get summary failed").
			Err(err).
			Extra("len", len(replies)).
			Errorx(ctx)
		return err
	}

	type key struct {
		BizCode int32
		Oid     uint64
	}
	mapping := make(map[key]uint64, len(resp.Responses))
	for _, item := range resp.Responses {
		mapping[key{item.BizCode, item.Oid}] = item.Count
	}

	for _, reply := range replies {
		k := key{global.CounterLikeBizcode, reply.Id}
		if cnt, ok := mapping[k]; ok {
			reply.LikeCount = cnt
		}
	}

	return nil
}
