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

	defaultRetry = 3

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

	retryHelper *retryHelper
	txHelper    *txHelper

	// 协程池组，按 protype hash 选择
	pools []*ants.Pool

	// 笔记处理的流水线
	pipeline *pipeline // 标准发布流程流水线

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
	registry.Register(NewAuditProcedure(bizz))
	registry.Register(NewPublishProcedure(bizz))

	txHelper := newTxHelper(bizz)
	retryHelper := newRetryHelper2(bizz, txHelper)
	m := &Manager{
		c: c,

		registry: registry,
		shardMgr: shardMgr,

		bizz:             bizz,
		noteBiz:          bizz.Note,
		noteProcedureBiz: bizz.Procedure,
		noteCreatorBiz:   bizz.Creator,
		retryHelper:      retryHelper,
		txHelper:         txHelper,

		pools:  pools,
		quitCh: make(chan struct{}),
	}

	// 初始化正常流水线
	standardPipeline, err := innerStandardPipeline(m)
	if err != nil {
		return nil, fmt.Errorf("init standard pipeline failed: %w", err)
	}

	m.pipeline = standardPipeline

	return m, nil
}

func (m *Manager) selectPool(protype model.ProcedureType) *ants.Pool {
	h := fnv.New32a()
	h.Write([]byte(protype))
	idx := h.Sum32() % uint32(len(m.pools))
	return m.pools[idx]
}

type RunPipelineParam struct {
	Note       *model.Note
	StartStage pipelineStage
	Extra      any
}

// 从某个流程节点开始运行流水线
//
// 返回值:
//
//	proceed: 外部需要调用此函数来继续后续流程, 包含Execute+Confirm,
//	如果流程中设置了自动成功 则自动调用OnSuccess
func (m *Manager) RunPipeline(
	ctx context.Context,
	param *RunPipelineParam,
) (proceed func(ctx context.Context) bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = xerror.Wrapf(xerror.ErrInternalPanic, "%v", r)
		}
	}()
	ppl := m.pipeline
	procType := ppl.startAt(param.StartStage)
	// 本地执行记录
	_, procParam, err := m.Create(ctx, param, procType, defaultRetry)
	if err != nil {
		return nil, xerror.Wrapf(err, "procedure manager run pipeline failed").
			WithExtras("note_id", param.Note.NoteId, "ppltype", param.StartStage).WithCtx(ctx)
	}

	return m.pipelineProceed(procParam, procType), nil
}

// 外部需要调用此函数来继续执行后续流程
func (m *Manager) pipelineProceed(
	param *ProcedureParam,
	targetProcType model.ProcedureType,
) func(context.Context) bool {
	return func(ctx context.Context) bool {
		// 任务开始执行 一般涉及对外调用 错误仅打日志 后续有重试机制
		newTaskId, err := m.Execute(ctx, param, targetProcType)
		logExtras := []any{"note_id", param.Note.NoteId, "protype", targetProcType, "task_id", newTaskId}
		if err != nil {
			xlog.Msg("procedure manager execute failed").
				Err(err).
				Extras(logExtras...).
				Errorx(ctx)
			return false
		}

		if newTaskId != "" {
			// 确认任务创建成功既可（回填taskId）错误仅打日志 后续有重试机制
			err = m.Confirm(ctx, param.Note.NoteId, targetProcType, newTaskId)
			if err != nil {
				xlog.Msg("procedure manager confirm failed").
					Err(err).
					Extras(logExtras...).
					Errorx(ctx)
				return false
			}

			xlog.Msgf("procedure manager confirm success").
				Extras(logExtras...).
				Infox(ctx)
		}

		xlog.Msgf("procedure manager execute success").
			Extras(logExtras...).
			Infox(ctx)

		m.autoCompleteIfNeeded(ctx, param, targetProcType, newTaskId)
		return true
	}
}

// Create 初始化流程
//
// 创建流程记录并标记笔记状态为处理中, 应该作为本地事务的一部分
func (m *Manager) Create(
	ctx context.Context,
	param *RunPipelineParam,
	protype model.ProcedureType,
	maxRetryCnt int,
) (Procedure, *ProcedureParam, error) {
	proc, ok := m.registry.Get(protype)
	if !ok {
		return nil, nil, xerror.Wrap(ErrProcedureNotRegistered).WithExtra("protype", protype).WithCtx(ctx)
	}

	procParam := &ProcedureParam{
		Note:  param.Note,
		Extra: param.Extra,
	}

	// 流程初始化
	doRecord, err := proc.BeforeExecute(ctx, procParam)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "procedure manager pre start failed").
			WithExtras("note_id", param.Note.NoteId, "protype", protype).
			WithCtx(ctx)
	}

	if !doRecord {
		return proc, procParam, nil
	}

	var params []byte
	if paramProvider, ok := any(proc).(ProcedureParamProvider); ok {
		params = paramProvider.Provide(procParam)
	}

	// taskId 先留空后续再填充
	err = m.noteProcedureBiz.CreateRecord(ctx, &biz.CreateProcedureRecordReq{
		NoteId:      param.Note.NoteId,
		Protype:     protype,
		TaskId:      "",
		MaxRetryCnt: maxRetryCnt,
		Params:      params,
	})
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "procedure manager create record failed").
			WithExtra("note_id", param.Note.NoteId).
			WithCtx(ctx)
	}

	return proc, procParam, nil
}

// Execute 执行流程
//
// 根据流程类型调度处理任务 可以执行远程调用或者本地数据库操作
func (m *Manager) Execute(
	ctx context.Context,
	param *ProcedureParam,
	proctype model.ProcedureType,
) (string, error) {
	proc, ok := m.registry.Get(proctype)
	if !ok {
		return "", xerror.Wrap(ErrProcedureNotRegistered).WithExtra("proctype", proctype).WithCtx(ctx)
	}

	taskId, err := proc.Execute(ctx, param)
	if err != nil {
		return "", xerror.Wrapf(err, "procedure manager execute failed").
			WithExtras("note_id", param.Note.NoteId, "protype", proctype).
			WithCtx(ctx)
	}

	return taskId, nil
}

// Confirm 确认流程 用于回填 taskId
func (m *Manager) Confirm(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	taskId string,
) error {
	err := m.noteProcedureBiz.UpdateTaskId(ctx, noteId, protype, taskId)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager confirm failed").
			WithExtras("note_id", noteId, "taskId", taskId).
			WithCtx(ctx)
	}

	return nil
}

// 终止某个流程
func (m *Manager) Abort(ctx context.Context,
	note *model.Note,
	proctype model.ProcedureType,
	taskId string,
) error {
	proc, ok := m.registry.Get(proctype)
	if !ok {
		return xerror.Wrap(ErrProcedureNotRegistered).WithExtra("proctype", proctype).WithCtx(ctx)
	}

	err := proc.OnAbort(ctx, note, taskId)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager abort failed").
			WithExtras(
				"note_id", note.NoteId,
				"proctype", proctype,
				"task_id", taskId).
			WithCtx(ctx)
	}

	return nil
}

func (m *Manager) GetTask(
	ctx context.Context,
	noteId int64,
	proctype model.ProcedureType,
) (*biz.ProcedureRecord, error) {
	record, err := m.noteProcedureBiz.GetRecord(ctx, noteId, proctype)
	if err != nil {
		return nil, xerror.Wrapf(err, "procedure manager get record failed").
			WithExtras("note_id", noteId, "proctype", proctype).
			WithCtx(ctx)
	}

	return record, nil
}

// 取消任务 标记为失败
func (m *Manager) CancelTask(ctx context.Context, noteId int64, proctype model.ProcedureType) error {
	err := m.noteProcedureBiz.MarkFailed(ctx, noteId, proctype)
	if err != nil {
		return xerror.Wrapf(err, "procedure manager cancel task failed").
			WithExtras("note_id", noteId, "proctype", proctype).
			WithCtx(ctx)
	}

	return nil
}

// CompleteResult 完成流程的入参
type CompleteResult struct {
	NoteId  int64
	Protype model.ProcedureType
	TaskId  string
	Success bool
	Arg     any
}

// Complete 完成流程（成功或失败）
func (m *Manager) Complete(ctx context.Context, result *CompleteResult) error {
	proc, ok := m.registry.Get(result.Protype)
	if !ok {
		return xerror.Wrap(ErrProcedureNotRegistered).WithExtra("protype", result.Protype).WithCtx(ctx)
	}

	if result.Success {
		return m.handleSuccess(ctx, result, proc)
	}

	// 失败就不处理流水线流转
	return m.handleFailure(ctx, result, proc)
}

func (m *Manager) handleFailure(
	ctx context.Context,
	result *CompleteResult,
	proc Procedure,
) error {
	return m.txHelper.txHandleFailure(ctx, result, proc.OnFailure)
}

func (m *Manager) handleSuccess(ctx context.Context, result *CompleteResult, proc Procedure) error {
	err := m.txHelper.txHandleSuccess(ctx, result, proc.OnSuccess)
	if err != nil {
		return xerror.Wrap(err)
	}

	// 成功后需要将流水线流转到下一个流程
	nextProcType := m.pipeline.nextOf(result.Protype)
	if nextProcType != "" {
		// 启动流水线的下一个流程
		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name:             "note.procedure.complete.success.next_proc",
			InheritCtxCancel: false,
			LogOnError:       true,
			Job: func(ctx context.Context) error {
				curNote, err := m.noteBiz.GetNoteCoreWithoutCache(ctx, result.NoteId)
				if err != nil {
					return xerror.Wrapf(err, "procedure manager get note without cache failed").
						WithExtra("note_id", result.NoteId).
						WithCtx(ctx)
				}
				proceed, err := m.RunPipeline(ctx, &RunPipelineParam{
					Note:       curNote,
					StartStage: nextProcType,
				})
				if err != nil {
					return xerror.Wrapf(err, "procedure manager run pipeline failed").
						WithExtra("note_id", result.NoteId).
						WithExtra("next_stage", nextProcType).
						WithCtx(ctx)
				}
				_ = proceed(ctx) // 继续后续流程

				return nil
			},
		})
	}

	return nil
}

func (m *Manager) CompleteAssetSuccess(ctx context.Context, noteId int64, taskId string, arg any) error {
	return m.Complete(ctx, &CompleteResult{
		NoteId:  noteId,
		Protype: model.ProcedureTypeAssetProcess,
		TaskId:  taskId,
		Success: true,
		Arg:     arg,
	})
}

func (m *Manager) CompleteAssetFailure(ctx context.Context, noteId int64, taskId string, arg any) error {
	return m.Complete(ctx, &CompleteResult{
		NoteId:  noteId,
		Protype: model.ProcedureTypeAssetProcess,
		TaskId:  taskId,
		Success: false,
		Arg:     arg,
	})
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

		targetStatus = model.ProcedureStatusProcessing
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

		// 拿出当前分片的所有待检查任务 只检查status=Processing的任务
		req := &biz.ListRangeScannedRecordsReq{
			Status:     model.ProcedureStatusProcessing,
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

	xlog.Msgf("procedure manager retry slot, range=[%d, %d), total=%d", start, end, len(totalRecords)).Infox(ctx)

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

// 真正重试逻辑
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
	// 拿到锁后再检查一遍记录状态
	curRecord, err := m.noteProcedureBiz.GetRecord(ctx, record.NoteId, record.Protype)
	if err != nil {
		xlog.Msg("procedure manager get latest record failed, will use old record to retry").
			Err(err).
			Extras("note_id", record.NoteId, "protype", record.Protype, "record_id", record.Id).
			Errorx(ctx)
	}

	if curRecord.Status != model.ProcedureStatusProcessing {
		xlog.Msg("procedure manager skip retry, record already handled").
			Extras("note_id", record.NoteId, "protype", record.Protype, "status", curRecord.Status).
			Infox(ctx)
		return
	}

	if curRecord == nil {
		// 拿不到最新的newRecord就用旧的重试
		curRecord = record
	}

	m.doRetry(newCtx, curRecord)
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

	if err := m.retryHelper.retry(
		ctx,
		record,
		proc.PollResult,
		proc.Retry,
		proc.OnSuccess,
		proc.OnFailure,
	); err != nil {
		xlog.Msg("procedure manager retry failed").
			Err(err).
			Extras("protype", record.Protype, "record_id", record.Id).
			Errorx(ctx)
	}
}

func (m *Manager) autoCompleteIfNeeded(
	ctx context.Context,
	param *ProcedureParam,
	protype model.ProcedureType,
	taskId string,
) {
	proc, ok := m.registry.Get(protype)
	if !ok {
		return
	}
	if proc, ok := proc.(AutoCompleter); ok {
		success, autoComplete, arg := proc.AutoComplete(ctx, param, taskId)
		if autoComplete {
			m.Complete(ctx, &CompleteResult{
				NoteId:  param.Note.NoteId,
				Protype: protype,
				TaskId:  taskId,
				Success: success,
				Arg:     arg,
			})
		}
	}
}
