package biz

import (
	"context"
	"errors"
	"sync"
	"time"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/global"
	"github.com/ryanreadbooks/whimer/counter/internal/infra"
	recorddao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/record"
	summarydao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/summary"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
)

type CounterBiz struct {
	cursorObfuscator obfuscate.Obfuscate
}

func MustNewCounterBiz(c *config.Config) *CounterBiz {
	obs, err := obfuscate.NewConfuser(c.Obfuscate.Options()...)
	if err != nil {
		panic(err)
	}

	s := &CounterBiz{
		cursorObfuscator: obs,
	}

	return s
}

// 新增计数记录
func (s *CounterBiz) AddRecord(ctx context.Context,
	req *counterv1.AddRecordRequest) (*counterv1.AddRecordResponse, error) {
	var (
		biz = req.BizCode
		uid = req.Uid
		oid = req.Oid
	)

	data, err := infra.Dao().RecordRepo.Find(ctx, uid, oid, biz)
	if err != nil && !xsql.IsNoRecord(err) {
		xlog.Msg("add record find failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}
	if data != nil && data.Act == recorddao.ActDo {
		return nil, global.ErrAlreadyDo // 重复操作
	}

	newData := &recorddao.Record{
		BizCode: biz,
		Uid:     uid,
		Oid:     oid,
		Act:     recorddao.ActDo,
	}

	// 没有处理过，可以处理
	err = infra.Dao().RecordRepo.InsertUpdate(ctx, newData)
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

	// update cache
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "counter.biz.addrecord.cache.set",
		Job: func(ctx context.Context) error {
			err := infra.Dao().RecordCache.AddRecord(ctx, newData)
			if err != nil {
				xlog.Msg("counter biz add record cache failed").Err(err).Errorx(ctx)
			}

			return nil
		},
	})

	return &counterv1.AddRecordResponse{}, nil
}

// 取消计数记录
func (s *CounterBiz) CancelRecord(ctx context.Context,
	req *counterv1.CancelRecordRequest) (*counterv1.CancelRecordResponse, error) {
	var (
		biz = req.BizCode
		uid = req.Uid
		oid = req.Oid
	)
	data, err := infra.Dao().RecordRepo.Find(ctx, uid, oid, biz)
	if err != nil && !xsql.IsNoRecord(err) {
		xlog.Msg("cancel record find failed").
			Err(err).
			Extra("oid", oid).
			Extra("uid", uid).
			Extra("biz", biz).
			Errorx(ctx)
		return nil, global.ErrInternal
	}
	if data != nil && data.Act == recorddao.ActUndo {
		return nil, global.ErrAlreadyDo // 重复操作
	}

	newData := &recorddao.Record{
		BizCode: biz,
		Uid:     uid,
		Oid:     oid,
		Act:     recorddao.ActUndo,
	}

	err = infra.Dao().RecordRepo.InsertUpdate(ctx, newData)
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

	// update cache
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "counter.biz.cancelrecord.cache.set",
		Job: func(ctx context.Context) error {
			infra.Dao().RecordCache.AddRecord(ctx, newData)
			if err != nil {
				xlog.Msg("counter biz cancel record cache failed").Err(err).Errorx(ctx)
			}

			return nil
		},
	})

	return &counterv1.CancelRecordResponse{}, nil
}

func (s *CounterBiz) updateSummary(ctx context.Context, oid int64, biz int32, positive bool) {
	s.updateSummaryNow(ctx, oid, biz, positive)
}

func (s *CounterBiz) updateSummaryNow(ctx context.Context, oid int64, biz int32, positive bool) error {
	var err error
	if positive {
		err = infra.Dao().SummaryRepo.InsertOrIncr(ctx, int(biz), oid)
	} else {
		err = infra.Dao().SummaryRepo.InsertOrDecr(ctx, int(biz), oid)
	}
	if !errors.Is(err, xsql.ErrOutOfRange) && !xsql.IsMildErr(err) {
		xlog.Msg("update summary repo failed").
			Err(err).
			Extra("oid", oid).
			Extra("positive", positive).
			Extra("biz", biz).
			Errorx(ctx)
		return err
	}

	return err
}

func (s *CounterBiz) GetRecord(ctx context.Context,
	req *counterv1.GetRecordRequest) (*counterv1.GetRecordResponse, error) {

	cacheRecord, err := infra.Dao().RecordCache.GetRecord(ctx, req.BizCode, req.Uid, req.Oid)
	if err == nil && cacheRecord != nil {
		return &counterv1.GetRecordResponse{Record: pbRecordFromDaoRecord(cacheRecord)}, nil
	}

	isDataFake := false
	data, err := infra.Dao().RecordRepo.Find(ctx, req.Uid, req.Oid, req.BizCode)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			xlog.Msg("cancel record repo insert update failed").
				Err(err).
				Extra("oid", req.Oid).
				Extra("uid", req.Uid).
				Extra("biz", req.BizCode).
				Errorx(ctx)
			return nil, global.ErrInternal
		} else {
			// 没有找到记录 也就是没有产生过计数动作和取消计数的动作 可以填一条假数据到缓存中
			now := time.Now().Unix()
			data = &recorddao.Record{
				Act:     recorddao.ActUnspecified,
				BizCode: req.BizCode,
				Uid:     req.Uid,
				Oid:     req.Oid,
				Ctime:   now,
				Mtime:   now,
			}
			isDataFake = true
		}
	}

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "counter.biz.getrecord.cache.set",
		Job: func(ctx context.Context) error {
			opts := []recorddao.CacheOption{}
			if isDataFake {
				// 假数据的缓存时间设置比真数据短
				opts = append(opts, recorddao.WithExpire(xtime.NHourJitter(2, time.Minute*15)))
			}
			if err := infra.Dao().RecordCache.AddRecord(ctx, data, opts...); err != nil {
				xlog.Msg("counter biz add record after get record failed").Err(err).Errorx(ctx)
			}
			return nil
		},
	})

	return &counterv1.GetRecordResponse{Record: pbRecordFromDaoRecord(data)}, nil
}

// 检查是否有正向计数记录
func (s *CounterBiz) CheckHasActDo(ctx context.Context, req *counterv1.CheckHasActDoRequest) (
	bool, error,
) {
	cacheHas, err := infra.Dao().RecordCache.CounterListExistsOid(ctx, req.BizCode, req.Uid, req.Oid)
	if err == nil && cacheHas {
		return cacheHas, nil
	}

	if err != nil {
		xlog.Msg("recode cache exists oid err").Err(err).Errorx(ctx)
	}

	record, err := s.GetRecord(ctx, &counterv1.GetRecordRequest{
		BizCode: req.BizCode,
		Uid:     req.Uid,
		Oid:     req.Oid,
	})
	if err != nil {
		return false, xerror.Wrapf(err, "counter biz failed to get record").
			WithExtras("biz_code", req.BizCode, "uid", req.Uid, "oid", req.Oid).
			WithCtx(ctx)
	}

	has := record.GetRecord().GetAct() == counterv1.RecordAct_RECORD_ACT_ADD
	if has {
		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "counter.biz.checkhasact.recordcache.add",
			Job: func(ctx context.Context) error {
				if err := infra.Dao().RecordCache.CounterListAdd(ctx, req.BizCode, req.Uid,
					&recorddao.CacheRecord{
						Act:   recorddao.ActDo,
						Oid:   req.Oid,
						Mtime: record.Record.Mtime,
					}); err != nil {
					xlog.Msg("background record cache batch add failed").
						Err(err).
						Extras("biz", req.BizCode, "uid", req.Uid, "oid", req.Oid).
						Errorx(ctx)
				}

				return nil
			},
		})
	}

	return has, nil
}

func (s *CounterBiz) BatchGetRecord(ctx context.Context, uidOids map[int64][]int64, biz int32) (
	map[int64][]*counterv1.Record, error) {

	var (
		finalResult  = make(map[int64][]*counterv1.Record, len(uidOids))
		fallbackKeys = make(map[int64][]int64, len(uidOids))
	)

	for uid, oids := range uidOids {
		cacheKeys := make([]recorddao.CacheKey, 0, len(oids))
		for _, oid := range oids {
			cacheKeys = append(cacheKeys, recorddao.CacheKey{
				BizCode: biz,
				Uid:     uid,
				Oid:     oid,
			})
		}

		// query cache first
		existingCache, err := infra.Dao().RecordCache.BatchGetRecord(ctx, cacheKeys)
		if err != nil {
			continue
		}

		// biz+uid+oid
		for _, record := range existingCache {
			// pre-fill final results
			finalResult[uid] = append(finalResult[uid], pbRecordFromDaoRecord(record))
		}

		for uid, oids := range uidOids {
			for _, oid := range oids {
				if _, ok := existingCache[recorddao.CacheKey{BizCode: biz, Uid: uid, Oid: oid}]; !ok {
					// we can not find uid+oid in cache, need to find them in db
					fallbackKeys[uid] = append(fallbackKeys[uid], oid)
				}
			}
		}
	}

	if len(fallbackKeys) == 0 {
		return finalResult, nil
	}

	datas, err := infra.Dao().RecordRepo.BatchFind(ctx, fallbackKeys, biz)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return nil, xerror.Wrapf(err, "batch find failed")
		}
	}

	// 那些既没有计数动作又没有取消计数动作的造一些假数据放进缓存
	tmpDatasMap := xslice.MakeMap(datas, func(v *recorddao.Record) recorddao.CacheKey {
		return recorddao.CacheKey{BizCode: biz, Uid: v.Uid, Oid: v.Oid}
	})

	fakeDatas := make([]*recorddao.Record, 0, len(datas))
	now := time.Now().Unix()
	for uid, oids := range fallbackKeys {
		for _, oid := range oids {
			_, ok := tmpDatasMap[recorddao.CacheKey{BizCode: biz, Uid: uid, Oid: oid}]
			if !ok {
				fakeDatas = append(fakeDatas, &recorddao.Record{
					BizCode: biz,
					Uid:     uid,
					Oid:     oid,
					Act:     recorddao.ActUnspecified,
					Ctime:   now,
					Mtime:   now,
				})
			}
		}
	}

	for _, data := range datas {
		finalResult[data.Uid] = append(finalResult[data.Uid], pbRecordFromDaoRecord(data))
	}

	// set back to cache
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "counter.biz.batchgetrecord.cache.set",
		Job: func(ctx context.Context) error {
			if err := infra.Dao().RecordCache.BatchAddRecord(ctx, datas); err != nil {
				xlog.Msg("counter biz add record cache after batch get record failed").Err(err).Errorx(ctx)
			}
			// 假数据的缓存时间要短
			if err := infra.Dao().RecordCache.BatchAddRecord(ctx,
				fakeDatas,
				recorddao.WithExpire(xtime.NHourJitter(2, time.Minute*15))); err != nil {
				xlog.Msg("counter biz add fake record cache after batch get record failed").Err(err).Errorx(ctx)
			}

			return nil
		},
	})

	return finalResult, nil
}

// 批量检查是否有正向计数记录
func (s *CounterBiz) BatchCheckHasActDo(ctx context.Context, uidOids map[int64][]int64, biz int32) (
	map[int64][]*counterv1.BatchCheckHasActDoResponse_Item, error,
) {
	resp := make(map[int64][]int64, 0)
	// 需要补偿查库的部分 因为缓存中不是全量数据 可能会被裁剪
	compensating := make(map[int64][]int64, 0)
	for uid, oids := range uidOids {
		oidsCounted, err := infra.Dao().RecordCache.CounterListBatchExistsOid(ctx, biz, uid, oids...)
		if err == nil {
			for _, oid := range oids {
				if _, ok := oidsCounted[oid]; !ok {
					compensating[uid] = append(compensating[uid], oid)
				} else {
					resp[uid] = append(resp[uid], oid) // 断定为true的
				}
			}
		}
	}

	final := make(map[int64][]*counterv1.BatchCheckHasActDoResponse_Item)
	for uid, oids := range resp {
		for _, oid := range oids {
			final[uid] = append(final[uid], &counterv1.BatchCheckHasActDoResponse_Item{
				Do:  true,
				Oid: oid,
			})
		}
	}

	// 检查是否需要查库
	if len(compensating) == 0 {
		return final, nil
	}

	// 需要查库补偿
	compensatedResult, err := s.BatchGetRecord(ctx, compensating, biz)
	if err != nil {
		return nil, xerror.Wrapf(err, "counter biz batch get record failed")
	}

	fillingResulsts := make([]*counterv1.Record, 0, len(compensatedResult))
	for uid, compensated := range compensatedResult {
		for _, record := range compensated {
			item := &counterv1.BatchCheckHasActDoResponse_Item{
				Oid: record.Oid,
				Do:  false,
			}
			if record.GetAct() == counterv1.RecordAct_RECORD_ACT_ADD {
				item.Do = true
				fillingResulsts = append(fillingResulsts, record)
			}
			final[uid] = append(final[uid], item)
		}
	}

	// set back to cache
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "counter.biz.batchcheckhasact.recordcache.batchadd",
		Job: func(ctx context.Context) error {
			batches := make(map[int64][]*counterv1.Record)
			for _, r := range fillingResulsts {
				batches[r.Uid] = append(batches[r.Uid], r)
			}

			for uid, oids := range batches {
				records := make([]*recorddao.CacheRecord, 0, len(oids))
				for _, record := range oids {
					records = append(records, &recorddao.CacheRecord{
						Act:   recorddao.ActDo,
						Oid:   record.Oid,
						Mtime: record.Mtime,
					})
				}

				if err := infra.Dao().RecordCache.CounterListBatchAdd(ctx, biz, uid, records); err != nil {
					xlog.Msg("background record cache batch add failed").
						Err(err).
						Extras("biz", biz, "uid", uid).
						Errorx(ctx)
				}
			}
			return nil
		},
	})

	return nil, nil
}

// 获取某个oid的计数
func (s *CounterBiz) GetSummary(ctx context.Context, req *counterv1.GetSummaryRequest) (
	*counterv1.GetSummaryResponse, error) {

	// 直接从数据库拿
	var (
		biz = req.BizCode
		oid = req.Oid
	)

	number, err := infra.Dao().SummaryRepo.Get(ctx, int(biz), oid)
	if err != nil && !xsql.IsNoRecord(err) {
		xlog.Msg("get summary repo failed").Err(err).
			Extra("oid", oid).
			Extra("biz", biz).
			Errorx(ctx)

		return nil, global.ErrInternal
	}

	return &counterv1.GetSummaryResponse{
		BizCode: req.BizCode,
		Oid:     req.Oid,
		Count:   number,
	}, nil
}

// 批量获取某个oid的计数
func (s *CounterBiz) BatchGetSummary(ctx context.Context, req *counterv1.BatchGetSummaryRequest) (
	*counterv1.BatchGetSummaryResponse, error) {
	const batchsize = 200

	var (
		summaryRes = make([]map[summarydao.PrimaryKey]int64, 0)
		wg         sync.WaitGroup
		mu         sync.Mutex
	)

	err := xslice.BatchAsyncExec(&wg, req.Requests, batchsize, func(start, end int) error {
		reqs := req.Requests[start:end]
		conds := make(summarydao.PrimaryKeyList, 0, len(reqs))
		for _, req := range reqs {
			conds = append(conds, &summarydao.PrimaryKey{
				BizCode: req.BizCode,
				Oid:     req.Oid,
			})
		}
		res, err := infra.Dao().SummaryRepo.Gets(ctx, conds)
		if err != nil {
			return global.ErrCountSummary
		}

		mu.Lock()
		defer mu.Unlock()
		summaryRes = append(summaryRes, res)

		return nil
	})

	if err != nil {
		xlog.Msg("batch get summary failed").Err(err).Errorx(ctx)
		return nil, global.ErrCountSummary
	}

	// 整理结果
	merged := make(map[summarydao.PrimaryKey]int64, len(summaryRes))
	for _, sumRes := range summaryRes {
		for k, v := range sumRes {
			merged[k] = v
		}
	}

	responses := make([]*counterv1.GetSummaryResponse, 0, len(summaryRes))
	for k, v := range merged {
		responses = append(responses, &counterv1.GetSummaryResponse{
			BizCode: k.BizCode,
			Oid:     k.Oid,
			Count:   v,
		})
	}

	return &counterv1.BatchGetSummaryResponse{Responses: responses}, nil
}

func (b *CounterBiz) PageListRecords(ctx context.Context, bizCode int32, uid int64, param PageListRecordsParam) (
	[]*counterv1.Record, PageResult, error) {

	var (
		sortOrder   = recorddao.Desc
		cursorMtime int64
		cursorId    int64
		err         error

		records []*recorddao.Record
	)

	if param.Order == PageListAscOrder {
		sortOrder = recorddao.Asc
	}

	hasCursor := false
	if param.Cursor != "" {
		var errParse error
		cursorMtime, cursorId, errParse = param.ParseCursor(b.cursorObfuscator)
		if errParse == nil {
			hasCursor = true
		}
	}

	var count = param.Count + 1 // fetch one more record

	if !hasCursor {
		records, err = infra.Dao().RecordRepo.PageGetByUidOrderByMtime(ctx,
			bizCode,
			recorddao.PageGetByUidOrderByMtimeParam{
				Uid:   uid,
				Count: count,
				Order: sortOrder,
			})
	} else {
		records, err = infra.Dao().RecordRepo.PageGetByUidOrderByMtimeWithCursor(ctx,
			bizCode,
			recorddao.PageGetByUidOrderByMtimeParam{
				Uid:   uid,
				Count: count,
				Order: sortOrder,
			},
			recorddao.PageGetByUidOrderByMtimeCursor{
				Mtime: cursorMtime,
				Id:    cursorId,
			})
	}

	var nextPage PageResult

	if err != nil {
		return nil, nextPage, xerror.Wrapf(err, "counter biz failed to page get by uid").
			WithExtras("uid", uid, "biz_code", bizCode).WithCtx(ctx)
	}

	gotLen := len(records)
	if gotLen == int(count) {
		// has more
		nextPage.HasNext = true
		records = records[0 : gotLen-1]
		// we calculate the next cursor
		nextCursor := records[len(records)-1]
		nextPage.NextCursor = PageListRecordsParam{}.FormatCursor(nextCursor.Mtime, nextCursor.Id, b.cursorObfuscator)
	} else {
		nextPage.HasNext = false
	}

	var resp = make([]*counterv1.Record, 0, len(records))
	for _, r := range records {
		resp = append(resp, NewPbRecord(r))
	}

	return resp, nextPage, nil
}
