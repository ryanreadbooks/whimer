package svc

import (
	"context"
	"errors"
	"math/bits"
	"strconv"
	"strings"

	gcache "github.com/patrickmn/go-cache"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/global"
	"github.com/ryanreadbooks/whimer/counter/internal/repo"
	"github.com/ryanreadbooks/whimer/counter/internal/repo/record"
	"github.com/ryanreadbooks/whimer/counter/internal/repo/summary"
	v1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CounterSvc struct {
	c          *config.Config
	root       *ServiceContext
	repo       *repo.Repo
	cache      *redis.Redis
	sumcounter *gcache.Cache
}

func NewCounterSvc(ctx *ServiceContext, repo *repo.Repo, cache *redis.Redis) *CounterSvc {
	s := &CounterSvc{
		root:       ctx,
		repo:       repo,
		cache:      cache,
		c:          ctx.Config,
		sumcounter: gcache.New(0, 0),
	}

	return s
}

func (s *CounterSvc) summaryKey(oid uint64, bizcode int) string {
	// summary:biz:oid
	return "summary:" + strconv.Itoa(bizcode) + ":" + strconv.FormatUint(oid, 10)
}

// 新增计数记录
func (s *CounterSvc) AddRecord(ctx context.Context,
	req *v1.AddRecordRequest) (*v1.AddRecordResponse, error) {
	var (
		biz = req.BizCode
		uid = req.Uid
		oid = req.Oid
	)

	data, err := s.repo.RecordRepo.Find(ctx, uid, oid, int(biz))
	if err != nil && !xsql.IsNotFound(err) {
		xlog.Msg("add record find failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}
	if data != nil && data.Act == record.ActDo {
		return nil, global.ErrAlreadyDo // 重复操作
	}

	// 没有处理过，可以处理
	err = s.repo.RecordRepo.InsertUpdate(ctx, &record.Model{
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

	// handle summary data
	s.updateSummary(ctx, oid, biz, true)

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
	data, err := s.repo.RecordRepo.Find(ctx, uid, oid, int(biz))
	if err != nil && !xsql.IsNotFound(err) {
		xlog.Msg("cancel record find failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}
	if data != nil && data.Act == record.ActUndo {
		return nil, global.ErrAlreadyDo // 重复操作
	}

	err = s.repo.RecordRepo.InsertUpdate(ctx, &record.Model{
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

	// handle summary data
	s.updateSummary(ctx, oid, biz, false)

	return &v1.CancelRecordResponse{}, nil
}

func (s *CounterSvc) updateSummary(ctx context.Context, oid uint64, biz int32, positive bool) {
	// TODO determine to use updateSummaryNow or cacheSummary based on database overload
	s.cacheSummary(ctx, oid, biz, positive)
}

func (s *CounterSvc) updateSummaryNow(ctx context.Context, oid uint64, biz int32, positive bool) {
	var err error
	if positive {
		err = s.repo.SummaryRepo.Incr(ctx, int(biz), oid)
	} else {
		err = s.repo.SummaryRepo.Decr(ctx, int(biz), oid)
	}
	if !errors.Is(err, xsql.ErrOutOfRange) && !xsql.IsMildErr(err) {
		xlog.Msg("update summary repo failed").
			Err(err).
			Extra("oid", oid).
			Extra("biz", biz).
			Errorx(ctx)
		return
	}
}

func (s *CounterSvc) cacheSummary(ctx context.Context, oid uint64, biz int32, positive bool) {
	k := s.summaryKey(oid, int(biz))
	if _, ok := s.sumcounter.Get(k); ok {
		if positive {
			_, err := s.sumcounter.IncrementInt64(k, 1)
			if err != nil {
				xlog.Msg("record sumcounter incr failed").Err(err).Extra("key", k).Errorx(ctx)
			}
		} else {
			_, err := s.sumcounter.DecrementInt64(k, 1)
			if err != nil {
				xlog.Msg("record sumcounter decr failed").Err(err).Extra("key", k).Errorx(ctx)
			}
		}
	} else {
		initval := 1
		if !positive {
			initval = -1
		}
		s.sumcounter.Set(k, int64(initval), -1)
	}
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

// 获取某个oid的计数
func (s *CounterSvc) GetSummary(ctx context.Context, req *v1.GetSummaryRequest) (*v1.GetSummaryResponse, error) {
	// 直接从数据库拿
	var (
		biz = req.BizCode
		oid = req.Oid
	)

	number, err := s.repo.SummaryRepo.Get(ctx, int(biz), oid)
	if err != nil && !xsql.IsNotFound(err) {
		xlog.Msg("get summary repo failed").Err(err).
			Extra("oid", oid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}

	return &v1.GetSummaryResponse{
		BizCode: req.BizCode,
		Oid:     req.Oid,
		Count:   number,
	}, nil
}

// 同步增量数据到数据库
func (s *CounterSvc) SyncCacheSummary(ctx context.Context) {
	var (
		batchsize = 500
	)

	items := s.sumcounter.Items()
	s.sumcounter.Flush() // 重新开始计数
	type delta struct {
		Biz   int
		Oid   uint64
		Delta int64
	}
	deltas := make([]*delta, 0, len(items))

	for k, v := range items {
		seps := strings.Split(k, ":")
		biz, err := strconv.Atoi(seps[1])
		if err != nil {
			continue
		}
		oid, err := strconv.ParseUint(seps[2], 10, 64)
		if err != nil {
			continue
		}
		num, ok := v.Object.(int64)
		if !ok {
			continue
		}

		deltas = append(deltas, &delta{
			Biz:   biz,
			Oid:   oid,
			Delta: num,
		})
	}

	slices.BatchExec(deltas, batchsize, func(start, end int) error {
		tmps := deltas[start:end]
		keys := make(summary.PrimaryKeyList, 0, len(tmps))
		deltaMaps := make(map[summary.PrimaryKey]int64)

		for _, delta := range tmps {
			k := &summary.PrimaryKey{BizCode: delta.Biz, Oid: delta.Oid}
			keys = append(keys, k)
			deltaMaps[*k] = delta.Delta
		}

		// 先查出来
		result, err := s.repo.SummaryRepo.Gets(ctx, keys)
		if err != nil {
			xlog.Msg("sync summary repo gets failed").Err(err).Error()
			return err
		}

		// 再设置回去
		newVals := make(map[summary.PrimaryKey]uint64)
		for key, cur := range result {
			num := deltaMaps[key] // > or < or == 0
			// cur为当前计数 num为需要变化的计数
			var newCur uint64
			if num >= 0 {
				sum, overflow := bits.Add64(cur, uint64(num), 0)
				if overflow != 0 {
					// overflow.
					xlog.Msg("sync summary bits.Add64 overflow").Extra("cur", cur).Extra("num", num).Error()
					newCur = cur // stays the same
				} else {
					newCur = sum
				}
			} else {
				num = -num // abs, > 0
				diff, underflow := bits.Sub64(cur, uint64(num), 0)
				if underflow != 0 {
					xlog.Msg("sync summary bits.Sub64 underflow").Extra("cur", cur).Extra("num", num).Error()
					newCur = 0
				} else {
					newCur = diff
				}
			}

			newVals[key] = newCur
		}

		datas := make([]*summary.Model, 0, len(newVals))
		for k, v := range newVals {
			datas = append(datas, &summary.Model{BizCode: k.BizCode, Oid: k.Oid, Cnt: v})
		}

		if err := s.repo.SummaryRepo.BatchInsert(ctx, datas); err != nil {
			xlog.Msg("sync summary repo batch insert failed").
				Err(err).
				Errorx(ctx)
			return err
		}

		return nil
	})
}
