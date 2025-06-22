package xlog

import (
	"context"
	"errors"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/metadata"
)

var (
	ctx = context.Background()
)

func TestMain(m *testing.M) {
	ctx = metadata.WithUid(ctx, 10010)
	m.Run()
}

func TestLog(t *testing.T) {
	Debugx(ctx, "hello test").doLog()
	Infox(ctx, "你好").Err(errors.New("new err")).doLog()
	Errorx(ctx, "failure").Err(nil).Extra("hello", "world").Extra("hi", "world").
		Field("trace", "test").Field("wanner", "back").doLog()
}

func TestLogV2(t *testing.T) {
	Err(errors.New("debug err")).Debug()
	Err(errors.New("debug err2")).Debugx(ctx)
	Err(errors.New("debug err")).Debug()
	Err(errors.New("debug err2")).Extra("oid", 100).Extra("nihao", "yes").Field("trace", "100").Error()
	Err(errors.New("debug err2")).Extra("oid", 100).Extra("nihao", "yes").Field("trace", "100").Errorx(ctx)
}
