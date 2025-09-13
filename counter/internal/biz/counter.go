package biz

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/global"
	"github.com/ryanreadbooks/whimer/counter/internal/infra"
	recorddao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/record"
	summarydao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/summary"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xstring"
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

	data, err := infra.Dao().RecordRepo.Find(ctx, uid, oid, int(biz))
	if err != nil && !xsql.IsNotFound(err) {
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

	// 没有处理过，可以处理
	err = infra.Dao().RecordRepo.InsertUpdate(ctx, &recorddao.Record{
		BizCode: biz,
		Uid:     uid,
		Oid:     oid,
		Act:     recorddao.ActDo,
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
	data, err := infra.Dao().RecordRepo.Find(ctx, uid, oid, int(biz))
	if err != nil && !xsql.IsNotFound(err) {
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

	err = infra.Dao().RecordRepo.InsertUpdate(ctx, &recorddao.Record{
		BizCode: biz,
		Uid:     uid,
		Oid:     oid,
		Act:     recorddao.ActUndo,
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

	data, err := infra.Dao().RecordRepo.Find(ctx, req.Uid, req.Oid, int(req.BizCode))
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
			return &counterv1.GetRecordResponse{Record: &counterv1.Record{
				Act: counterv1.RecordAct_RECORD_ACT_UNSPECIFIED,
			}}, nil // 找不到记录不当作错误
		}
	}

	return &counterv1.GetRecordResponse{Record: &counterv1.Record{
		BizCode: int32(data.BizCode),
		Uid:     data.Uid,
		Oid:     data.Oid,
		Act:     counterv1.RecordAct(data.Act),
		Ctime:   data.Ctime,
		Mtime:   data.Mtime,
	}}, nil
}

func (s *CounterBiz) BatchGetRecord(ctx context.Context, uidOids map[int64][]int64, biz int) (
	map[int64][]*counterv1.Record, error) {
	datas, err := infra.Dao().RecordRepo.BatchFind(ctx, uidOids, biz)
	var uidRecords = make(map[int64][]*counterv1.Record, len(datas))
	if err != nil {
		if !xsql.IsNotFound(err) {
			return nil, xerror.Wrapf(err, "batch find failed")
		} else {
			// 找不到不当作错误
			return uidRecords, nil
		}
	}

	for _, data := range datas {
		uidRecords[data.Uid] = append(uidRecords[data.Uid], &counterv1.Record{
			BizCode: int32(data.BizCode),
			Uid:     data.Uid,
			Oid:     data.Oid,
			Act:     counterv1.RecordAct(data.Act),
			Ctime:   data.Ctime,
			Mtime:   data.Mtime,
		})
	}

	return uidRecords, nil
}

// 获取某个oid的计数
func (s *CounterBiz) GetSummary(ctx context.Context, req *counterv1.GetSummaryRequest) (*counterv1.GetSummaryResponse, error) {
	// 直接从数据库拿
	var (
		biz = req.BizCode
		oid = req.Oid
	)

	number, err := infra.Dao().SummaryRepo.Get(ctx, int(biz), oid)
	if err != nil && !xsql.IsNotFound(err) {
		xlog.Msg("get summary repo failed").Err(err).
			Extra("oid", oid).
			Extra("biz", biz).
			Errorx(ctx)
		// TODO 可以尝试直接查record表

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

// 全表扫描 从record表更新summary的数据
func (s *CounterBiz) SyncSummaryFromRecords(ctx context.Context) error {
	total, err := infra.Dao().RecordRepo.CountAll(ctx)
	if err != nil {
		xlog.Msg("record repo count all failed").Err(err).Errorx(ctx)
		return err
	}

	xlog.Msg(fmt.Sprintf("record repo count all result: total = %d", total)).Info()
	// 点赞的数量
	actDoSum, err := infra.Dao().RecordRepo.GetSummary(ctx, recorddao.ActDo)
	if err != nil {
		xlog.Msg("record repo get actdo summary failed").Err(err).Errorx(ctx)
		return err
	}

	// 取消点赞的数量
	actUndoSum, err := infra.Dao().RecordRepo.GetSummary(ctx, recorddao.ActUndo)
	if err != nil {
		xlog.Msg("record repo get act undo summary failed").Err(err).Errorx(ctx)
		return err
	}

	if len(actDoSum) == 0 {
		return nil
	}

	keyFn := func(r *recorddao.Summary) string {
		return fmt.Sprintf("%d-%d", r.BizCode, r.Oid)
	}

	// 结合点赞和取消点赞修正最终的点赞数
	actUndoSumMap := make(map[string]*recorddao.Summary, len(actUndoSum))
	for _, undoSum := range actUndoSum {
		actUndoSumMap[keyFn(undoSum)] = undoSum
	}
	actDoSumMap := make(map[string]*recorddao.Summary, len(actDoSum))
	for _, doSum := range actDoSum {
		actDoSumMap[keyFn(doSum)] = doSum
	}

	datas := make([]*recorddao.Summary, 0, len(actDoSum))

	// 存在一种情况为: 被全部取消点赞，cnt需要为0
	for k, undoSum := range actUndoSumMap {
		if _, ok := actDoSumMap[k]; !ok {
			// 全部都是取消点赞数据，那么数据取值为0
			actDoSumMap[k] = &recorddao.Summary{
				BizCode: undoSum.BizCode,
				Oid:     undoSum.Oid,
				Cnt:     0,
			}
		}
	}
	for _, v := range actDoSumMap {
		datas = append(datas, &recorddao.Summary{
			BizCode: v.BizCode,
			Oid:     v.Oid,
			Cnt:     v.Cnt,
		})
	}

	batchsize := 500

	err = xslice.BatchExec(datas, batchsize, func(start, end int) error {
		data := datas[start:end]
		if len(data) == 0 {
			return nil
		}

		summaryModels := make([]*summarydao.Model, 0, len(data))
		for _, sub := range data {
			summaryModels = append(summaryModels, &summarydao.Model{
				BizCode: sub.BizCode,
				Oid:     sub.Oid,
				Cnt:     sub.Cnt,
			})
		}
		if err := infra.Dao().SummaryRepo.BatchInsert(ctx, summaryModels); err != nil {
			xlog.Msg("sync summary from records repo batch insert failed").
				Err(err).
				Errorx(ctx)
			return err
		}
		return nil
	})

	if err != nil {
		xlog.Msg("batch exec update summary failed").
			Err(err).
			Extra("len", len(actDoSum)).
			Errorx(ctx)
	}

	return err
}

type PageListOrder int8

const (
	PageListDescOrder PageListOrder = 0
	PageListAscOrder  PageListOrder = 1
)

type PageListRecordsParam struct {
	Cursor string
	Count  int32
	Order  PageListOrder
}

func (r *PageListRecordsParam) ParseCursor(obs obfuscate.Obfuscate) (mtime, id int64, err error) {
	raw, err := base64.RawStdEncoding.DecodeString(r.Cursor)
	if err != nil {
		return
	}

	s := xstring.FromBytes(raw)
	unpacked := strings.SplitN(s, ":", 2)
	if len(unpacked) != 2 {
		err = fmt.Errorf("%s is invalid cursor", s)
		return
	}

	mtimeStr := unpacked[0]
	mixIdStr := unpacked[1]
	mtime, err = obs.DeMix(mtimeStr)
	if err != nil {
		err = fmt.Errorf("invalid mtime: %w", err)
		return
	}

	id, err = obs.DeMix(mixIdStr)
	if err != nil {
		err = fmt.Errorf("invalid id: %w", err)
	}

	return
}

func (PageListRecordsParam) FormatCursor(mtime, id int64, obs obfuscate.Obfuscate) string {
	mtimeMix, _ := obs.Mix(mtime)
	idMix, _ := obs.Mix(id)
	cursor := mtimeMix + ":" + idMix
	return base64.RawStdEncoding.EncodeToString(xstring.AsBytes(cursor))
}

type PageListRecordsNextRequest struct {
	NextCursor string
	HasNext    bool
}

func (b *CounterBiz) PageListRecords(ctx context.Context, bizCode int32, uid int64, param PageListRecordsParam) (
	[]*counterv1.Record, PageListRecordsNextRequest, error) {

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

	var nextRequest PageListRecordsNextRequest

	if err != nil {
		return nil, nextRequest, xerror.Wrapf(err, "counter biz failed to page get by uid").
			WithExtras("uid", uid, "biz_code", bizCode).WithCtx(ctx)
	}

	gotLen := len(records)
	if gotLen == int(count) {
		// has more
		nextRequest.HasNext = true
		records = records[0 : gotLen-1]
		// we calculate the next cursor
		nextCursor := records[len(records)-1]
		nextRequest.NextCursor = PageListRecordsParam{}.FormatCursor(nextCursor.Mtime, nextCursor.Id, b.cursorObfuscator)
	} else {
		nextRequest.HasNext = false
	}

	var resp = make([]*counterv1.Record, 0, len(records))
	for _, r := range records {
		resp = append(resp, NewPbRecord(r))
	}

	return resp, nextRequest, nil
}

func NewPbRecord(r *recorddao.Record) *counterv1.Record {
	act := counterv1.RecordAct_RECORD_ACT_UNSPECIFIED
	switch r.Act {
	case recorddao.ActDo:
		act = counterv1.RecordAct_RECORD_ACT_ADD
	case recorddao.ActUndo:
		act = counterv1.RecordAct_RECORD_ACT_UNADD
	}
	return &counterv1.Record{
		BizCode: r.BizCode,
		Uid:     r.Uid,
		Oid:     r.Oid,
		Act:     act,
		Ctime:   r.Ctime,
		Mtime:   r.Mtime,
	}
}
