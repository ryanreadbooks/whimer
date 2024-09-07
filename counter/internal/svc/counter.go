package svc

import (
	"context"

	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/global"
	"github.com/ryanreadbooks/whimer/counter/internal/repo"
	"github.com/ryanreadbooks/whimer/counter/internal/repo/record"
	v1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CounterSvc struct {
	c     *config.Config
	root  *ServiceContext
	repo  *repo.Repo
	cache *redis.Redis
}

func NewCounterSvc(ctx *ServiceContext, repo *repo.Repo, cache *redis.Redis) *CounterSvc {
	return &CounterSvc{
		root:  ctx,
		repo:  repo,
		cache: cache,
		c:     ctx.Config,
	}
}

// 新增计数记录
func (s *CounterSvc) AddRecord(ctx context.Context,
	req *v1.AddRecordRequest) (*v1.AddRecordResponse, error) {
	var (
		biz = req.BizCode
		uid = req.Uid
		oid = req.Oid
	)

	err := s.repo.RecordRepo.InsertUpdate(ctx, &record.Model{
		BizCode: int(biz),
		Uid:     uid,
		Oid:     oid,
		Act:     record.ActDo,
	})
	if xsql.IsCriticalErr(err) {
		xlog.Msg("add record repo insert update failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}

	return &v1.AddRecordResponse{}, nil
}

// 取消计数记录
func (s *CounterSvc) CancelRecord(ctx context.Context,
	req *v1.CancelRecordRequest) (*v1.CancelRecordResponse, error) {
	var (
		biz = req.BizCode
		uid = req.Uid
		oid = req.Oid
	)

	err := s.repo.RecordRepo.InsertUpdate(ctx, &record.Model{
		BizCode: int(biz),
		Uid:     uid,
		Oid:     oid,
		Act:     record.ActUndo,
	})
	if xsql.IsCriticalErr(err) {
		xlog.Msg("cancel record repo insert update failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}

	return &v1.CancelRecordResponse{}, nil
}

func (s *CounterSvc) GetRecord(ctx context.Context,
	req *v1.GetRecordRequest) (*v1.GetRecordResponse, error) {

	data, err := s.repo.RecordRepo.Find(ctx, req.Uid, req.Oid, int(req.BizCode))
	if err != nil {
		if !xsql.IsNotFound(err) {
			xlog.Msg("cancel record repo insert update failed").
				Err(err).
				Extra("oid", req.Oid).
				Extra("uid", req.Uid).
				Extra("biz", req.BizCode).
				Errorx(ctx)
			return nil, global.ErrInternal
		} else {
			return nil, global.ErrNoRecord
		}
	}

	return &v1.GetRecordResponse{Record: &v1.Record{
		BizCode: int32(data.BizCode),
		Uid:     data.Uid,
		Oid:     data.Oid,
		Act:     v1.RecordAct(data.Act),
		Ctime:   data.Ctime,
		Mtime:   data.Mtime,
	}}, nil
}
