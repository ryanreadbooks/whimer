package xlog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

var (
	ctx = context.Background()
)

func TestMain(m *testing.M) {
	ctx = metadata.WithUid(ctx, 10010)
	m.Run()
}

func TestLog(t *testing.T) {
	xlog.Debugx(ctx, "hello test").Do()
	xlog.Infox(ctx, "你好").Err(errors.New("new err")).Do()
	xlog.Errorx(ctx, "failure").Err(nil).Extra("hello", "world").Extra("hi", "world").
		Field("trace", "test").Field("wanner", "back").Do()
}

func TestLogV2(t *testing.T) {
	xlog.Err(errors.New("debug err")).Debug()
	xlog.Err(errors.New("debug err2")).Debugx(ctx)
	xlog.Err(errors.New("debug err")).Debug()
	xlog.Err(errors.New("debug err2")).Extra("oid", 100).Extra("nihao", "yes").Field("trace", "100").Error()
	xlog.Err(errors.New("debug err2")).Extra("oid", 100).Extra("nihao", "yes").Field("trace", "100").Errorx(ctx)
}