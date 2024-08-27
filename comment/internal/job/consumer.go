package job

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type Job struct {
	Svc *svc.ServiceContext
}

func New(svc *svc.ServiceContext) *Job {
	j := &Job{
		Svc: svc,
	}

	return j
}

// Job需要实现kq.ConsumerHandler接口
func (j *Job) Consume(ctx context.Context, key, value string) error {
	xlog.Msg("job Consume").Extra("key", key).Extra("value", value).Debugx(ctx)
	var data queue.Data
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		xlog.Error("job Consume json.Unmarshal err").Err(err).Errorx(ctx)
		return err
	}

	switch data.Action {
	case queue.ActAddReply:
		return j.Svc.CommentSvc.ConsumeAddReplyEv(ctx, data.AddReplyData)
	case queue.ActDelReply:
		return j.Svc.CommentSvc.ConsumeDelReplyEv(ctx, data.DelReplyData)
	case queue.ActLikeReply, queue.ActDislikeReply:
		return j.Svc.CommentSvc.ConsumeLikeDislikeEv(ctx, data.LikeReplyData)
	case queue.ActPinReply:
		return j.Svc.CommentSvc.ConsumePinEv(ctx, data.PinReplyData)
	default:
		xlog.Msg("job Consumer got unsupported action type").Extra("type", data.Action).Debugx(ctx)
		return global.ErrInternal
	}
}
