package worker

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

func HandleVideoProcess(ctx context.Context, task *worker.Task) worker.Result {
	xlog.Msg("processing video task").Extra("taskId", task.Id).Infox(ctx)

	// TODO: 实现视频处理逻辑

	return worker.Result{Output: "hello"}
}
