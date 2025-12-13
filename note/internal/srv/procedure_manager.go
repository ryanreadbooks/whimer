package srv

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/shard"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv/assetprocess"

	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	minSlotGapSec = 5       // 5s
	maxFutureSec  = 30 * 60 // 30min

	defaultPoolCount    = 4
	defaultPoolCapacity = 150
)

// ProcedureManager 流程管理器
// 负责流程的创建、执行、回调和后台重试
type ProcedureManager struct {
	c *config.Config

	// 分片管理
	shardMgr *shard.Manager

	// 业务依赖
	bizz             *biz.Biz
	noteBiz          *biz.NoteBiz
	noteProcedureBiz *biz.NoteProcedureBiz
	noteCreatorBiz   *biz.NoteCreatorBiz

	// 协程池组，按 protype hash 选择
	pools []*ants.Pool

	wg     sync.WaitGroup
	quitCh chan struct{}
}

func NewProcedureManager(c *config.Config, bizz *biz.Biz) *ProcedureManager {
	mgr := shard.NewManager(infra.Etcd().GetClient(), c.Etcd.Key, global.GetHostname())

	pools := make([]*ants.Pool, defaultPoolCount)
	for i := range pools {
		pool, _ := ants.NewPool(defaultPoolCapacity)
		pools[i] = pool
	}

	return &ProcedureManager{
		c: c,

		shardMgr:         mgr,
		bizz:             bizz,
		noteBiz:          bizz.Note,
		noteProcedureBiz: bizz.Procedure,
		noteCreatorBiz:   bizz.Creator,

		pools:  pools,
		quitCh: make(chan struct{}),
	}
}

func (m *ProcedureManager) selectPool(protype model.ProcedureType) *ants.Pool {
	h := fnv.New32a()
	h.Write([]byte(protype))
	idx := h.Sum32() % uint32(len(m.pools))
	return m.pools[idx]
}

// Create 初始化流程
// 创建流程记录并标记笔记状态为处理中
func (m *ProcedureManager) Create(ctx context.Context, note *model.Note, protype model.ProcedureType) error {
	// taskId 先留空后续再填充
	err := m.noteProcedureBiz.CreateRecord(ctx, &biz.CreateProcedureRecordReq{
		NoteId:      note.NoteId,
		Protype:     protype,
		TaskId:      "",
		MaxRetryCnt: 3,
	})
	if err != nil {
		return xerror.Wrapf(err, "procedure manager create record failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// 标记开始处理
	err = m.noteCreatorBiz.SetNoteStateProcessing(ctx, note.NoteId)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager set note state processing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return nil
}

// Execute 执行流程
// 根据笔记类型调度处理任务
func (m *ProcedureManager) Execute(ctx context.Context, note *model.Note) (string, error) {
	assetProcessor := assetprocess.NewProcessor(note.Type, m.bizz)
	taskId, err := assetProcessor.Process(ctx, note)
	if err != nil {
		return "", xerror.Wrapf(err, "procedure manager execute failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return taskId, nil
}

// Confirm 确认流程
// 回填 taskId
func (m *ProcedureManager) Confirm(ctx context.Context, noteId int64, taskId string, protype model.ProcedureType) error {
	err := m.noteProcedureBiz.UpdateTaskId(ctx, noteId, protype, taskId)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager confirm failed").
			WithExtras("note_id", noteId, "taskId", taskId).
			WithCtx(ctx)
	}

	return nil
}

// Complete 完成流程
// 处理回调结果，标记成功
func (m *ProcedureManager) Complete(ctx context.Context, noteId int64, taskId string) error {
	record, err := m.noteProcedureBiz.GetRecord(ctx, noteId, model.ProcedureTypeAssetProcess)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager get record failed").
			WithExtra("taskId", taskId).
			WithCtx(ctx)
	}

	err = m.bizz.Tx(ctx, func(ctx context.Context) error {
		// 笔记状态标记处理完成
		err := m.noteCreatorBiz.SetNoteStateProcessed(ctx, record.NoteId)
		if err != nil {
			return xerror.Wrapf(err, "procedure manager set note state processed failed").
				WithExtra("noteId", record.NoteId).
				WithCtx(ctx)
		}

		// TODO 可能也有处理失败的情况 需要一并处理

		// 任务状态设置成功
		err = m.noteProcedureBiz.MarkSuccess(ctx, record.NoteId, record.Protype)
		if err != nil {
			return xerror.Wrapf(err, "procedure manager mark record success failed").
				WithExtra("taskId", taskId).
				WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		// 失败仅打日志 + 后台重试
		xlog.Msg("procedure manager tx failed").
			Err(err).
			Extras("taskId", taskId).
			Errorx(ctx)

		return nil
	}

	xlog.Msgf("procedure manager tx success").
		Extras("taskId", taskId, "noteId", record.NoteId).
		Infox(ctx)

	// TODO 异步进入下一流程 (审核)

	return nil
}

// StartRetryLoop 启动后台重试循环
func (m *ProcedureManager) StartRetryLoop(ctx context.Context) {
	err := m.shardMgr.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("shard manager start failed: %w", err))
	}

	m.wg.Add(1)
	concurrent.SafeGo2(
		ctx,
		concurrent.SafeGo2Opt{
			Name:             "note.procedure.bg.task_register_retry",
			InheritCtxCancel: true,
			Job: func(ctx context.Context) error {
				defer m.wg.Done()
				m.retryLoop(ctx)
				return nil
			},
		},
	)
}

// StopRetryLoop 停止后台重试循环
func (m *ProcedureManager) StopRetryLoop() {
	xlog.Msg("procedure manager stop retry loop").Info()
	close(m.quitCh)
	m.wg.Wait()
	m.shardMgr.Stop()

	// release all pools
	for _, pool := range m.pools {
		pool.Release()
	}
}

// retryLoop 重试循环
func (m *ProcedureManager) retryLoop(ctx context.Context) {
	// 随机扫描间隔
	interval := config.Conf.RetryConfig.ProcedureRetry.TaskRegister.ScanInterval
	ticker := time.NewTicker(time.Millisecond) // 第一次马上执行一遍扫描操作
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.quitCh:
			return
		case <-ticker.C:
			m.scanAndRetry(ctx)
			// 随机扫描间隔
			ticker.Reset(interval + time.Duration(rand.Intn(int(interval.Seconds())))*time.Second)
		}
	}
}

// scanAndRetry 扫描并重试
func (m *ProcedureManager) scanAndRetry(ctx context.Context) {
	var (
		slotGapSec = max(
			m.c.RetryConfig.ProcedureRetry.TaskRegister.SlotGapSec,
			minSlotGapSec,
		)

		futureSec = min(
			int64(m.c.RetryConfig.ProcedureRetry.TaskRegister.FutureInterval.Seconds()),
			maxFutureSec,
		)

		targetStatus = model.ProcessStatusProcessing
	)

	// 获取最早的一条记录
	record, err := m.noteProcedureBiz.GetEarliestScannedRecord(
		ctx,
		targetStatus,
	)
	if err != nil {
		xlog.Msg("procedure manager get earliest record failed").
			Err(err).
			Extras("targetStatus", targetStatus).
			Errorx(ctx)
		return
	}

	if record == nil {
		return
	}

	timeStart := record.NextCheckTime
	// 取整到能被slotGapSec整除
	timeStart = timeStart - timeStart%int64(slotGapSec)
	timeEnd := timeStart + futureSec

	// assign slot
	for slot := timeStart; slot < timeEnd; slot += int64(slotGapSec) {
		m.wg.Add(1)
		concurrent.SafeGo2(
			ctx,
			concurrent.SafeGo2Opt{
				Name:             "note.procedure.bg.assign_retry_slot",
				InheritCtxCancel: true,
				Job: func(ctx context.Context) error {
					defer m.wg.Done()
					m.retrySlot(ctx, slot, slot+int64(slotGapSec))
					return nil
				},
			},
		)
	}
}

func (m *ProcedureManager) retrySlot(ctx context.Context, start, end int64) {
	shard := m.shardMgr.GetShard()
	// 抢不到分片 无法执行 可能已经被其它节点执行
	if !shard.Active {
		return
	}

	var (
		limit        = m.c.RetryConfig.ProcedureRetry.TaskRegister.Limit
		lastId int64 = 0

		totalRecords []*biz.ProcedureRecord
	)

	for {
		exit := m.shouldExit(ctx)
		if exit {
			return
		}

		// 拿出当前分片的所有待检查任务
		req := &biz.ListRangeScannedRecordsReq{
			Status:     model.ProcessStatusProcessing,
			RangeStart: start,
			RangeEnd:   end,
			OffsetId:   lastId,
			Count:      limit,
			ShardIdx:   shard.Index,
			TotalShard: shard.Total,
		}

		records, err := m.noteProcedureBiz.ListRangeScannedRecords(ctx, req)
		if err != nil {
			xlog.Msg("procedure manager list range records failed").
				Err(err).
				Extras("req", req).
				Errorx(ctx)
			return
		}

		if len(records) == 0 {
			break
		}

		lastId = records[len(records)-1].Id
		totalRecords = append(totalRecords, records...)
	}

	if len(totalRecords) == 0 {
		return
	}

	// 按照protype分组执行
	for _, record := range totalRecords {
		r := record // capture
		m.wg.Add(1)
		err := m.selectPool(r.Protype).Submit(recovery.DoV3(func() {
			defer m.wg.Done()
			m.retryRecord(ctx, r)
		}))
		if err != nil {
			xlog.Msg("procedure manager submit retry task failed").
				Err(err).
				Extras("record", r).
				Errorx(ctx)
		}
	}
}

func (m *ProcedureManager) retryRecord(ctx context.Context, record *biz.ProcedureRecord) {
	exit := m.shouldExit(ctx)
	if exit {
		return
	}

	now := time.Now()
	nextCheckTime := time.Unix(record.NextCheckTime, 0)
	diff := nextCheckTime.Sub(now)

	if diff <= 0 {
		goto exec
	}

	// 还没到 需要等待diff
	select {
	case <-ctx.Done():
		return
	case <-m.quitCh:
		return
	case <-time.After(diff):
	}

exec:
	// 上锁 + 检查是否可以执行
	lockKey := record.GetLockKey()
	locker := redis.NewRedisLock(infra.Cache(), lockKey)
	expire := time.Minute * 5
	locker.SetExpire(int(expire.Seconds()))
	newCtx, cancel := context.WithTimeout(ctx, expire)
	defer cancel()
	held, err := locker.AcquireCtx(newCtx)
	if err != nil {
		xlog.Msgf("procedure manager acquire lock failed").
			Err(err).
			Extra("lock_key", lockKey).
			Extras("record_id", record.Id).
			Errorx(ctx)
		return
	}

	if !held {
		return
	}
	defer locker.ReleaseCtx(newCtx)
	m.doRetry(newCtx, record)
}

func (m *ProcedureManager) shouldExit(ctx context.Context) (exit bool) {
	select {
	case <-ctx.Done():
		exit = true
		return
	case <-m.quitCh:
		exit = true
		return
	default:
	}
	return
}

// doRetry 执行重试
// record存在可能有两种情况：
// 1. 存在taskId表明注册成功 此时等待回调即可无需重新注册
// 2. 不存在taskId 认为注册失败 此时按照重试次数重新注册
func (m *ProcedureManager) doRetry(ctx context.Context, record *biz.ProcedureRecord) {
	exit := m.shouldExit(ctx)
	if exit {
		return
	}

	switch record.Protype {
	case model.ProcedureTypeAssetProcess:
		m.retryAssetProcess(ctx, record)
	default:
	}
}

func (m *ProcedureManager) retryAssetProcess(ctx context.Context, record *biz.ProcedureRecord) {
	noteType, err := m.noteBiz.GetNoteType(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("procedure manager get note type failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}

	processor := assetprocess.NewProcessor(noteType, m.bizz)

	// 存在taskId表明注册成功，主动轮询结果
	if record.TaskId != "" {
		m.pollTaskResult(ctx, record, processor)
		return
	}

	// 检查是否允许重试
	if record.CurRetry >= record.MaxRetryCnt {
		m.markRetryExhausted(ctx, record)
		return
	}

	// 不存在taskId，重新执行任务
	m.reExecuteTask(ctx, record, processor)
}

// pollTaskResult 轮询任务结果
func (m *ProcedureManager) pollTaskResult(
	ctx context.Context,
	record *biz.ProcedureRecord,
	processor assetprocess.Processor,
) {
	_, ok, err := processor.GetTaskResult(ctx, record.TaskId)
	if err != nil {
		xlog.Msg("procedure manager get task result failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}
	if !ok {
		return
	}

	// 执行成功，完成流程
	err = m.Complete(ctx, record.NoteId, record.TaskId)
	if err != nil {
		xlog.Msg("procedure manager complete failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
	}
}

// markRetryExhausted 标记重试耗尽
func (m *ProcedureManager) markRetryExhausted(ctx context.Context, record *biz.ProcedureRecord) {
	// 发送通知 发布流程中间有一环失败了 发布失败
	err := m.noteProcedureBiz.MarkFailed(ctx, record.NoteId, record.Protype)
	if err != nil {
		xlog.Msg("procedure manager mark failed failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
	}
}

// reExecuteTask 重新执行任务
func (m *ProcedureManager) reExecuteTask(
	ctx context.Context,
	record *biz.ProcedureRecord,
	processor assetprocess.Processor,
) {
	note, err := m.noteBiz.GetNoteWithoutCache(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("procedure manager get note failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}

	nextCheckTime := time.Now().Add(config.Conf.RetryConfig.ProcedureRetry.TaskRegister.RetryInterval)
	taskId, err := processor.Process(ctx, note)
	if err != nil {
		xlog.Msg("procedure manager retry process failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)

		// 重试次数增加 并等待下一次重试
		err = m.noteProcedureBiz.UpdateRetry(
			ctx, record.NoteId, record.Protype, nextCheckTime.Unix(),
		)
		if err != nil {
			xlog.Msg("procedure manager update retry failed").
				Err(err).
				Extras("record", record).
				Errorx(ctx)
		}

		return
	}

	// 重试成功更新记录
	record.TaskId = taskId
	record.NextCheckTime = nextCheckTime.Unix()
	record.CurRetry++
	err = m.noteProcedureBiz.UpdateTaskIdRetryNextCheckTime(ctx, record)
	if err != nil {
		xlog.Msg("procedure manager update record failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
	}
}
