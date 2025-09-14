package biz

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/counter/internal/infra"
	recorddao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/record"
	summarydao "github.com/ryanreadbooks/whimer/counter/internal/infra/dao/summary"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
)

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
