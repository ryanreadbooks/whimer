package worker

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

func HandleImageProcess(ctx context.Context, task *worker.Task) worker.Result {
	xlog.Msg("processing image task").Extra("taskId", task.Id).Infox(ctx)

	// TODO: 实现图片处理逻辑

	return worker.Result{Output: "image-hello"}
}
