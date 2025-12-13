package srv

import (
	"context"
	"fmt"
	"hash/fnv"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
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
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	minSlotGapSec = 5       // 5s
	maxFutureSec  = 30 * 60 // 30min

	defaultPoolCount    = 4
	defaultPoolCapacity = 150
)

// 负责笔记状态流转
type NoteProcedureSrv struct {
	c *config.Config

	// 分片管理
	shardMgr *shard.Manager

	// 各业务逻辑
	bizz             biz.Biz
	noteProcedureBiz biz.NoteProcedureBiz
	noteCreatorBiz   biz.NoteCreatorBiz

	// 协程池组，按 protype hash 选择
	pools []*ants.Pool

	wg     sync.WaitGroup
	quitCh chan struct{}
}

func NewNoteProcedureSrv(c *config.Config, biz biz.Biz) *NoteProcedureSrv {
	mgr := shard.NewManager(infra.Etcd().GetClient(), c.Etcd.Key, global.GetHostname())

	pools := make([]*ants.Pool, defaultPoolCount)
	for i := range pools {
		pool, _ := ants.NewPool(defaultPoolCapacity)
		pools[i] = pool
	}

	srv := &NoteProcedureSrv{
		c: c,

		shardMgr:         mgr,
		bizz:             biz,
		noteProcedureBiz: biz.Procedure,
		noteCreatorBiz:   biz.Creator,

		pools:  pools,
		quitCh: make(chan struct{}),
	}

	return srv
}

func (s *NoteProcedureSrv) selectPool(protype model.ProcedureType) *ants.Pool {
	h := fnv.New32a()
	h.Write([]byte(protype))
	idx := h.Sum32() % uint32(len(s.pools))
	return s.pools[idx]
}

type HandleAssetProcessResultReq struct {
	NoteId int64
	TaskId string
}

// 调度任务完成后回调处理逻辑
//
// 此处需要将状态标记为已成功并且进入下一流程
func (s *NoteProcedureSrv) HandleAssetProcessResult(
	ctx context.Context,
	req *HandleAssetProcessResultReq,
) error {
	record, err := s.noteProcedureBiz.GetRecord(ctx, req.NoteId, model.ProcedureTypeAssetProcess)
	if err != nil {
		return xerror.Wrapf(err, "process biz get record failed").
			WithExtra("taskId", req.TaskId).
			WithCtx(ctx)
	}

	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		// 笔记状态标记处理完成
		err := s.noteCreatorBiz.SetNoteStateProcessed(ctx, record.NoteId)
		if err != nil {
			return xerror.Wrapf(err, "creator set note state processed failed").
				WithExtra("noteId", record.NoteId).
				WithCtx(ctx)
		}

		// TODO 可能也有处理失败的情况 需要一并处理

		// 任务状态设置成功
		err = s.noteProcedureBiz.MarkSuccess(ctx, record.NoteId, record.Protype)
		if err != nil {
			return xerror.Wrapf(err, "process biz mark record success failed").
				WithExtra("taskId", req.TaskId).
				WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		// 失败仅打日志 + 后台重试
		xlog.Msg("process biz tx failed").
			Err(err).
			Extras("taskId", req.TaskId).
			Errorx(ctx)

		return nil
	}

	xlog.Msgf("process biz tx success").
		Extras("taskId", req.TaskId, "noteId", record.NoteId).
		Infox(ctx)

	// TODO 异步进入下一流程 (审核)

	return nil
}

// 处理错误情况
func (s *NoteProcedureSrv) goStartBackgroundHandle(ctx context.Context) {
	err := s.shardMgr.Start(ctx)
	if err != nil {
		panic(fmt.Errorf("shard manager start failed: %w", err))
	}

	s.wg.Add(1)
	concurrent.SafeGo2(
		ctx,
		concurrent.SafeGo2Opt{
			Name:             "note.procedure.bg.task_register_retry",
			InheritCtxCancel: true,
			Job: func(ctx context.Context) error {
				defer s.wg.Done()
				s.handlePendingTasks(ctx)
				return nil
			},
		},
	)
}

func (s *NoteProcedureSrv) StopBackgroundHandle() {
	// stop all background tasks
	xlog.Msg("note procedure srv stop background handle").Info()
	close(s.quitCh)
	s.wg.Wait()
	s.shardMgr.Stop()

	// release all pools
	for _, pool := range s.pools {
		pool.Release()
	}
}

// 需要重新注册的情况包括：
//
// 1. 笔记落库成功但任务注册失败
func (s *NoteProcedureSrv) handlePendingTasks(ctx context.Context) {
	// TODO 马上扫描一次 + 随机扫描间隔 防止多pod同时扫相同内容
	ticker := time.NewTicker(config.Conf.RetryConfig.ProcedureRetry.TaskRegister.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.quitCh:
			return
		case <-ticker.C:
			s.doRetryPendingTasks(ctx)
		}
	}
}

// 任务注册重试逻辑
func (s *NoteProcedureSrv) doRetryPendingTasks(ctx context.Context) {
	var (
		slotGapSec = max(
			s.c.RetryConfig.ProcedureRetry.TaskRegister.SlotGapSec,
			minSlotGapSec,
		)

		futureSec = min(
			int64(s.c.RetryConfig.ProcedureRetry.TaskRegister.FutureInterval.Seconds()),
			maxFutureSec,
		)

		targetStatus = model.ProcessStatusProcessing
	)

	// 获取最早的一条记录
	record, err := s.noteProcedureBiz.GetEarliestScannedRecord(
		ctx,
		targetStatus,
	)
	if err != nil {
		xlog.Msg("note procedure biz get earliest next check time record failed").
			Err(err).
			Extras("targetStatus", targetStatus).
			Errorx(ctx)
		return
	}

	if record == nil {
		return
	}

	timeStart := record.NextCheckTime // 开始扫描时间
	// 取整到能被slotGapSec整除
	timeStart = timeStart - timeStart%int64(slotGapSec)
	timeEnd := timeStart + futureSec

	// assign slot
	for slot := timeStart; slot < timeEnd; slot += int64(slotGapSec) {
		s.wg.Add(1)
		concurrent.SafeGo2(
			ctx,
			concurrent.SafeGo2Opt{
				Name:             "note.procedure.bg.assign_retry_task_registration_slot",
				InheritCtxCancel: true,
				Job: func(ctx context.Context) error {
					defer s.wg.Done()
					s.assignRetryPendingTasksSlot(ctx, slot, slot+int64(slotGapSec))
					return nil
				},
			},
		)
	}
}

func (s *NoteProcedureSrv) assignRetryPendingTasksSlot(ctx context.Context, start, end int64) {
	shard := s.shardMgr.GetShard()
	// 抢不到分片 无法执行 可能已经被其它节点执行
	if !shard.Active {
		return
	}

	var (
		limit        = s.c.RetryConfig.ProcedureRetry.TaskRegister.Limit
		lastId int64 = 0

		totalRecords []*biz.ProcedureRecord
	)

	for {
		exit := s.selectQuitCtx(ctx)
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

		records, err := s.noteProcedureBiz.ListRangeScannedRecords(ctx, req)
		if err != nil {
			xlog.Msg("note procedure biz list range scanned records failed").
				Err(err).
				Extras("req", req).
				Errorx(ctx)
			return
		}

		if len(records) == 0 {
			break
		}

		lastId = records[len(records)-1].Id
		// append to records
		totalRecords = append(totalRecords, records...)
	}

	if len(totalRecords) == 0 {
		// no need to handle
		return
	}

	// 按照protype分组执行
	// totalRecords的next_retry_time都是升序排列
	for _, record := range totalRecords {
		r := record // capture
		s.wg.Add(1)
		err := s.selectPool(r.Protype).Submit(recovery.DoV3(func() {
			defer s.wg.Done()
			s.execRetryPendingTask(ctx, r)
		}))
		if err != nil {
			xlog.Msg("note procedure biz submit retry pending task to pool failed").
				Err(err).
				Extras("record", r).
				Errorx(ctx)
		}
	}
}

func (s *NoteProcedureSrv) execRetryPendingTask(
	ctx context.Context,
	record *biz.ProcedureRecord,
) {
	exit := s.selectQuitCtx(ctx)
	if exit {
		return
	}

	now := time.Now()
	nextCheckTime := time.Unix(record.NextCheckTime, 0)
	diff := nextCheckTime.Sub(now)

	if diff <= 0 {
		// 已经过了检查时间 直接执行
		goto exec
	}

	// 还没到 需要等待diff
	select {
	case <-ctx.Done():
		return
	case <-s.quitCh:
		return
	case <-time.After(diff):
		// 等待时间到了 直接执行
		return
	default:
	}

exec:
	// 上锁 + 检查是否可以执行
	lockKey := record.GetLockKey()
	locker := redis.NewRedisLock(infra.Cache(), lockKey)
	expire := time.Minute * 5 // 暂定锁5分钟 并且下面任务执行超时时间也是5min
	locker.SetExpire(int(expire.Seconds()))
	newCtx, cancel := context.WithTimeout(ctx, expire)
	defer cancel()
	held, err := locker.AcquireCtx(newCtx)
	if err != nil {
		xlog.Msgf("note procedure biz acquire lock failed").
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
	s.doExecPendingTask(newCtx, record)
}

func (s *NoteProcedureSrv) selectQuitCtx(ctx context.Context) (exit bool) {
	select {
	case <-ctx.Done():
		exit = true
		return
	case <-s.quitCh:
		exit = true
		return
	default:
	}
	return
}

// record存在可能有两种情况：
// 1. 存在taskId表明注册成功 此时等待回调即可无需重新注册
// 2. 不存在taskId 认为注册失败 此时按照重试次数重新注册
func (s *NoteProcedureSrv) doExecPendingTask(ctx context.Context, record *biz.ProcedureRecord) {
	exit := s.selectQuitCtx(ctx)
	if exit {
		return
	}

	switch record.Protype {
	case model.ProcedureTypeAssetProcess:
		s.retryHandleAssetProcess(ctx, record)
	default:

	}
}

func (s *NoteProcedureSrv) retryHandleAssetProcess(
	ctx context.Context, record *biz.ProcedureRecord,
) {
	noteType, err := s.noteCreatorBiz.GetNoteType(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("note procedure biz get note type failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}

	processor := assetprocess.NewProcessor(noteType, s.bizz)
	if record.TaskId != "" {
		// 存在taskId表明注册成功 此时主动查一遍
		_, ok, err := processor.GetTaskResult(ctx, record.TaskId)
		if err != nil {
			xlog.Msg("note procedure biz get task result failed").
				Err(err).
				Extras("record", record).
				Errorx(ctx)
			return
		}
		if !ok {
			return
		}

		// 执行成功
		err = s.HandleAssetProcessResult(ctx, &HandleAssetProcessResultReq{
			NoteId: record.NoteId,
			TaskId: record.TaskId,
		})
		if err != nil {
			xlog.Msg("note procedure biz handle asset process result failed").
				Err(err).
				Extras("record", record).
				Errorx(ctx)
			return
		}

		return
	}

	note, err := s.noteCreatorBiz.GetNoteWithoutCache(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("note procedure biz get note without cache failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}

	// 检查是否允许重试
	if record.CurRetry >= record.MaxRetryCnt {
		// 发送通知 发布流程中间有一环失败了 发布失败
		err = s.noteProcedureBiz.MarkFailed(ctx, record.NoteId, record.Protype)
		if err != nil {
			xlog.Msg("note procedure biz mark failed failed").
				Err(err).
				Extras("record", record).
				Errorx(ctx)
			return
		}
		return
	}

	// 不存在taskId 认为注册失败 此时按照重试次数重新注册
	taskId, err := processor.Process(ctx, note)
	if err != nil {
		xlog.Msg("note procedure biz process asset process failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
		return
	}

	nextCheckTime := time.Now().Add(config.Conf.RetryConfig.ProcedureRetry.TaskRegister.RetryInterval)
	record.TaskId = taskId
	record.NextCheckTime = nextCheckTime.Unix()
	record.CurRetry++
	record.Status = model.ProcessStatusProcessing // 重试成功 重新标记为处理中
	record.TaskId = taskId
	err = s.noteProcedureBiz.UpdateRecord(ctx, record)
	if err != nil {
		xlog.Msg("note procedure biz update record failed").
			Err(err).
			Extras("record", record).
			Errorx(ctx)
	}
}
