package procedure

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

	"github.com/panjf2000/ants/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	minSlotGapSec = 5       // 5s
	maxFutureSec  = 30 * 60 // 30min

	defaultPoolCount    = 4
	defaultPoolCapacity = 150
)

// Manager 流程管理器
// 负责流程的创建、执行、回调和后台重试
// 通过 Registry 支持多种流程类型的扩展
type Manager struct {
	c *config.Config

	// 流程注册
	registry *Registry

	// 分片管理
	shardMgr *shard.Manager

	// 业务依赖
	bizz             *biz.Biz
	noteBiz          *biz.NoteBiz
	noteProcedureBiz *biz.NoteProcedureBiz
	noteCreatorBiz   *biz.NoteCreatorBiz

	// 协程池组，按 protype hash 选择
	pools []*ants.Pool

	// 笔记处理的流水线
	standardPipeline *pipeline // 标准发布流程流水线

	wg     sync.WaitGroup
	quitCh chan struct{}
}

func NewManager(c *config.Config, bizz *biz.Biz) (*Manager, error) {
	shardMgr := shard.NewManager(infra.Etcd().GetClient(), c.Etcd.Key, global.GetHostname())
	pools := make([]*ants.Pool, defaultPoolCount)
	for i := range pools {
		pool, _ := ants.NewPool(defaultPoolCapacity)
		pools[i] = pool
	}

	// 流程实现注册 后续有新增的流程都需要注册在这里
	registry := NewRegistry()
	registry.Register(NewAssetProcedure(bizz))
	registry.Register(NewPublishProcedure())

	m := &Manager{
		c: c,

		registry: registry,
		shardMgr: shardMgr,

		bizz:             bizz,
		noteBiz:          bizz.Note,
		noteProcedureBiz: bizz.Procedure,
		noteCreatorBiz:   bizz.Creator,

		pools:  pools,
		quitCh: make(chan struct{}),
	}

	// 初始化正常流水线
	standardPipeline, err := innerStandardPipeline(m)
	if err != nil {
		return nil, fmt.Errorf("init standard pipeline failed: %w", err)
	}

	m.standardPipeline = standardPipeline

	return m, nil
}

func (m *Manager) selectPool(protype model.ProcedureType) *ants.Pool {
	h := fnv.New32a()
	h.Write([]byte(protype))
	idx := h.Sum32() % uint32(len(m.pools))
	return m.pools[idx]
}

// 开始笔记流程处理
//
// 返回值:
// 1. 继续执行后续流程的函数
// 2. 错误
func (m *Manager) BeginPipeline(
	ctx context.Context,
	note *model.Note,
	startAt PipelineStage,
) (func() bool, error) {
	ppl := m.standardPipeline
	targetProc := ppl.startAt(startAt)
	err := m.Create(ctx, note, targetProc, 3)
	if err != nil {
		return nil, xerror.Wrapf(err, "procedure manager begin pipeline failed").
			WithExtras("note_id", note.NoteId, "pipeline_type", startAt.String()).WithCtx(ctx)
	}

	// 外部需要调用此函数来继续执行后续流程
	proceed := func() bool {
		// 任务开始执行 一般涉及对外调用 错误仅打日志 后续有重试机制
		newTaskId, err := m.Execute(ctx, note, targetProc)
		if err != nil {
			xlog.Msg("procedure manager execute failed").
				Err(err).
				Extras("note_id", note.NoteId, "protype", targetProc).
				Errorx(ctx)
			return false
		}

		// 确认任务创建成功既可（回填taskId）错误仅打日志 后续有重试机制
		err = m.Confirm(ctx, note.NoteId, newTaskId, targetProc)
		if err != nil {
			xlog.Msg("procedure manager confirm failed").
				Err(err).
				Extras("note_id", note.NoteId, "protype", targetProc).
				Errorx(ctx)
			return false
		}

		return true
	}

	return proceed, nil
}

// Create 初始化流程
// 
// 创建流程记录并标记笔记状态为处理中, 应该作为本地事务的一部分
func (m *Manager) Create(
	ctx context.Context,
	note *model.Note,
	protype model.ProcedureType,
	maxRetryCnt int,
) error {
	proc, ok := m.registry.Get(protype)
	if !ok {
		return xerror.Wrap(ErrProcedureNotRegistered).WithExtra("protype", protype).WithCtx(ctx)
	}

	// taskId 先留空后续再填充
	err := m.noteProcedureBiz.CreateRecord(ctx, &biz.CreateProcedureRecordReq{
		NoteId:      note.NoteId,
		Protype:     protype,
		TaskId:      "",
		MaxRetryCnt: maxRetryCnt,
	})
	if err != nil {
		return xerror.Wrapf(err, "procedure manager create record failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// 流程初始化
	err = proc.PreStart(ctx, note)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager pre start failed").
			WithExtras("note_id", note.NoteId, "protype", protype).
			WithCtx(ctx)
	}

	return nil
}

// Execute 执行流程
// 
// 根据流程类型调度处理任务 可以执行远程调用或者本地数据库操作
func (m *Manager) Execute(
	ctx context.Context,
	note *model.Note,
	protype model.ProcedureType,
) (string, error) {
	proc, ok := m.registry.Get(protype)
	if !ok {
		return "", xerror.Wrap(ErrProcedureNotRegistered).WithExtra("protype", protype).WithCtx(ctx)
	}

	taskId, err := proc.Execute(ctx, note)
	if err != nil {
		return "", xerror.Wrapf(err, "procedure manager execute failed").
			WithExtras("note_id", note.NoteId, "protype", protype).
			WithCtx(ctx)
	}

	return taskId, nil
}

// Confirm 确认流程 用于回填 taskId
func (m *Manager) Confirm(
	ctx context.Context,
	noteId int64,
	taskId string,
	protype model.ProcedureType,
) error {
	err := m.noteProcedureBiz.UpdateTaskId(ctx, noteId, protype, taskId)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager confirm failed").
			WithExtras("note_id", noteId, "taskId", taskId).
			WithCtx(ctx)
	}

	return nil
}

// Complete 完成流程（成功或失败）
func (m *Manager) Complete(
	ctx context.Context,
	noteId int64,
	taskId string,
	protype model.ProcedureType,
	success bool,
) error {
	proc, ok := m.registry.Get(protype)
	if !ok {
		return xerror.Wrap(ErrProcedureNotRegistered).WithExtra("protype", protype).WithCtx(ctx)
	}

	if success {
		// 成功后需要将流水线流转到下一个流程
		return proc.OnSuccess(ctx, noteId, taskId)
	}

	// 失败就不处理流水线流转
	return proc.OnFailure(ctx, noteId, taskId)
}

func (m *Manager) CompleteAssetSuccess(ctx context.Context, noteId int64, taskId string) error {
	return m.Complete(ctx, noteId, taskId, model.ProcedureTypeAssetProcess, true)
}

func (m *Manager) CompleteAssetFailure(ctx context.Context, noteId int64, taskId string) error {
	return m.Complete(ctx, noteId, taskId, model.ProcedureTypeAssetProcess, false)
}

// StartRetryLoop 启动后台重试循环
func (m *Manager) StartRetryLoop(ctx context.Context) {
	err := m.shardMgr.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("shard manager start failed: %w", err))
	}

	m.wg.Add(1)
	concurrent.SafeGo2(
		ctx,
		concurrent.SafeGo2Opt{
			Name:             "note.procedure.bg.retry_loop",
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
func (m *Manager) StopRetryLoop() {
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
func (m *Manager) retryLoop(ctx context.Context) {
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
func (m *Manager) scanAndRetry(ctx context.Context) {
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

func (m *Manager) retrySlot(ctx context.Context, start, end int64) {
	shard := m.shardMgr.GetShard()
	// 抢不到分片 无法执行 可能已经被其它节点执行
	if !shard.Active {
		return
	}

	var (
		limit        = m.c.RetryConfig.ProcedureRetry.TaskRegister.ScanLimit
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

func (m *Manager) retryRecord(ctx context.Context, record *biz.ProcedureRecord) {
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

func (m *Manager) shouldExit(ctx context.Context) (exit bool) {
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
// 根据流程类型分发到对应的流程处理器
func (m *Manager) doRetry(ctx context.Context, record *biz.ProcedureRecord) {
	exit := m.shouldExit(ctx)
	if exit {
		return
	}

	proc, ok := m.registry.Get(record.Protype)
	if !ok {
		xlog.Msg("procedure manager retry unknown protype").
			Extras("protype", record.Protype, "record_id", record.Id).
			Errorx(ctx)
		return
	}

	if err := proc.Retry(ctx, record); err != nil {
		xlog.Msg("procedure manager retry failed").
			Err(err).
			Extras("protype", record.Protype, "record_id", record.Id).
			Errorx(ctx)
	}
}
