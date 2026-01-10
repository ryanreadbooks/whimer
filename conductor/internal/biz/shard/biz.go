package shard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/conductor/internal/global"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra/etcd"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Biz 分片管理
type Biz struct {
	conf       *config.Config
	cli        *clientv3.Client
	instanceId string
	strategy   Strategy

	// 组件
	registry *Registry
	claimer  *Claimer

	// 当前持有的分片范围
	mu         sync.RWMutex
	shardRange *Range

	// 退出信号（外部 → goroutine）
	quitCh chan struct{}
	// 完成通知信号（goroutine → 外部）
	doneCh chan struct{}
}

func NewBiz(conf *config.Config, etcdcli *etcd.Client) *Biz {
	cli := etcdcli.GetClient()
	instanceId := global.GetIpAndPort()
	quitCh := make(chan struct{})

	registry := NewRegistry(cli, instanceId, quitCh)
	claimer := NewClaimer(cli, instanceId, registry)

	return &Biz{
		conf:       conf,
		cli:        cli,
		instanceId: instanceId,
		// 默认使用均匀分片策略
		strategy: &EvenStrategy{},
		registry: registry,
		claimer:  claimer,
		quitCh:   quitCh,
		doneCh:   make(chan struct{}),
	}
}

// SetStrategy 设置分片策略
func (b *Biz) SetStrategy(strategy Strategy) {
	b.strategy = strategy
}

func (b *Biz) Run(ctx context.Context) {
	b.bgRun(ctx)
}

func (b *Biz) Stop() {
	close(b.quitCh)
	<-b.doneCh

	// 注销实例
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	b.unregister(ctx)
}

// GetShardRange 获取当前持有的分片范围
func (b *Biz) GetShardRange() Range {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.shardRange == nil {
		return Range{}
	}

	return *b.shardRange
}

// HasShard 判断是否持有分片
func (b *Biz) HasShard() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.shardRange != nil
}

// InRange 判断值是否在当前分片范围内
func (b *Biz) InRange(val int) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.shardRange == nil {
		return false
	}
	return val >= b.shardRange.Start && val < b.shardRange.End
}

func (b *Biz) bgRun(ctx context.Context) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "conductor.shard.biz.run",
		Job: func(ctx context.Context) error {
			defer close(b.doneCh)

			// 1. 注册实例
			if err := b.registry.Register(ctx); err != nil {
				xlog.Msg("register instance failed").
					Extras("instance", b.instanceId).
					Err(err).
					Errorx(ctx)
				return err
			}

			// 2. 首次分配分片
			if err := b.tryClaimShard(ctx); err != nil {
				xlog.Msg("initial claim shard failed").
					Extras("instance", b.instanceId).
					Err(err).
					Errorx(ctx)
			}

			// 3. 监听实例变化
			b.watchAndRebalance(ctx)

			return nil
		},
	})
}

func (b *Biz) unregister(ctx context.Context) {
	// 释放当前持有的分片
	b.mu.Lock()
	shard := b.shardRange
	b.shardRange = nil
	b.mu.Unlock()

	b.claimer.Release(ctx, shard)
	b.registry.Unregister(ctx)
}

// calculateShards 根据实例数量计算分片划分
func (b *Biz) calculateShards(instanceCount int) []Range {
	return b.strategy.Calculate(sizeOfShard, instanceCount)
}

// tryClaimShard 尝试抢占分片
func (b *Biz) tryClaimShard(ctx context.Context) error {
	// 获取实例数量
	instanceCount, err := b.registry.GetInstanceCount(ctx)
	if err != nil {
		return fmt.Errorf("get instance count failed: %w", err)
	}

	if instanceCount == 0 {
		return fmt.Errorf("no instances found")
	}

	// 计算分片划分
	shards := b.calculateShards(instanceCount)
	xlog.Msg("calculated shards").
		Extras("instance", b.instanceId,
			"instanceCount", instanceCount,
			"shards", shards).
		Infox(ctx)

	// 先释放当前持有的分片
	b.mu.Lock()
	oldShard := b.shardRange
	b.shardRange = nil
	b.mu.Unlock()
	b.claimer.Release(ctx, oldShard)

	// 清理过期的分片
	b.claimer.CleanupStale(ctx, instanceCount)

	// 尝试抢占每个分片
	for _, shard := range shards {
		claimed, err := b.claimer.Claim(ctx, shard)
		if err != nil {
			xlog.Msg("claim shard failed").
				Extras("instance", b.instanceId, "shardId", shard.ShardId).
				Err(err).
				Errorx(ctx)
			continue
		}
		if claimed {
			b.mu.Lock()
			b.shardRange = &shard
			b.mu.Unlock()
			xlog.Msg("claimed shard").
				Extras("instance", b.instanceId,
					"shard", shard.String(),
					"start", shard.Start,
					"end", shard.End,
					"shardId", shard.ShardId).
				Infox(ctx)
			return nil
		}
	}

	xlog.Msg("failed to claim any shard, will retry").
		Extras("instance", b.instanceId).
		Infox(ctx)
	return nil
}

// watchAndRebalance 监听实例和分片变化并重新平衡
func (b *Biz) watchAndRebalance(ctx context.Context) {
	// 同时 watch 实例和分片变化
	instanceWatchCh := b.registry.Watch(ctx)
	shardWatchCh := b.claimer.Watch(ctx)

	retryTimer := time.NewTimer(b.conf.Shard.GetClaimRetryInterval())
	defer retryTimer.Stop()

	for {
		select {
		case <-b.quitCh:
			return
		case <-ctx.Done():
			return

		case watchResp, ok := <-instanceWatchCh:
			if !ok {
				xlog.Msg("instance watch channel closed, reconnecting").
					Extras("instance", b.instanceId).
					Infox(ctx)
				time.Sleep(time.Second)
				instanceWatchCh = b.registry.Watch(ctx)
				continue
			}

			if watchResp.Err() != nil {
				xlog.Msg("instance watch error").
					Extras("instance", b.instanceId).
					Err(watchResp.Err()).
					Errorx(ctx)
				continue
			}

			// 实例变化，记录日志
			for _, ev := range watchResp.Events {
				xlog.Msg("instance event").
					Extras("instance", b.instanceId,
						"eventType", ev.Type.String(),
						"key", string(ev.Kv.Key),
						"leaseId", ev.Kv.Lease).
					Infox(ctx)
			}

			// 延迟一点再重新分配，避免频繁变动
			retryTimer.Reset(b.conf.Shard.GetClaimRetryInterval())

		case watchResp, ok := <-shardWatchCh:
			if !ok {
				xlog.Msg("shard watch channel closed, reconnecting").
					Extras("instance", b.instanceId).
					Infox(ctx)
				time.Sleep(time.Second)
				shardWatchCh = b.claimer.Watch(ctx)
				continue
			}

			if watchResp.Err() != nil {
				xlog.Msg("shard watch error").
					Extras("instance", b.instanceId).
					Err(watchResp.Err()).
					Errorx(ctx)
				continue
			}

			// 分片变化，记录日志
			for _, ev := range watchResp.Events {
				xlog.Msg("shard event").
					Extras("instance", b.instanceId,
						"eventType", ev.Type.String(),
						"key", string(ev.Kv.Key),
						"leaseId", ev.Kv.Lease).
					Infox(ctx)
			}

			// 分片变化时也需要触发重新分配（处理 lease 过期导致的分片删除）
			retryTimer.Reset(b.conf.Shard.GetClaimRetryInterval())

		case <-retryTimer.C:
			if !b.HasShard() {
				// 没有分片，尝试抢占
				if err := b.tryClaimShard(ctx); err != nil {
					xlog.Msg("retry claim shard failed").
						Extras("instance", b.instanceId).
						Err(err).
						Errorx(ctx)
				}
				// 没有分片时持续重试
				retryTimer.Reset(b.conf.Shard.GetClaimRetryInterval())
			} else {
				// 已有分片，检查是否需要重新分配
				b.checkAndRebalance(ctx)
				// 定期检查，确保最终一致性
				retryTimer.Reset(b.conf.Shard.GetCheckInterval())
			}
		}
	}
}

// checkAndRebalance 检查并重新平衡分片
func (b *Biz) checkAndRebalance(ctx context.Context) {
	instanceCount, err := b.registry.GetInstanceCount(ctx)
	if err != nil {
		xlog.Msg("get instance count failed").
			Extras("instance", b.instanceId).
			Err(err).
			Errorx(ctx)
		return
	}

	// 获取当前分片数量
	shardCount, err := b.claimer.GetShardCount(ctx)
	if err != nil {
		xlog.Msg("get shard count failed").
			Extras("instance", b.instanceId).
			Err(err).
			Errorx(ctx)
		return
	}

	// 检查分片数量是否和实例数量一致
	if shardCount != instanceCount {
		xlog.Msg("shard count mismatch, rebalancing").
			Extras("instance", b.instanceId,
				"shardCount", shardCount,
				"instanceCount", instanceCount).
			Infox(ctx)
		if err := b.tryClaimShard(ctx); err != nil {
			xlog.Msg("rebalance failed").
				Extras("instance", b.instanceId).
				Err(err).
				Errorx(ctx)
		}
		return
	}

	// 检查当前分片范围是否正确
	b.mu.RLock()
	currentShard := b.shardRange
	b.mu.RUnlock()

	if currentShard == nil {
		xlog.Msg("no current shard, try claim").
			Extras("instance", b.instanceId).
			Infox(ctx)
		if err := b.tryClaimShard(ctx); err != nil {
			xlog.Msg("claim shard failed").
				Extras("instance", b.instanceId).
				Err(err).
				Errorx(ctx)
		}
		return
	}

	shards := b.calculateShards(instanceCount)
	if currentShard.ShardId >= len(shards) {
		// 当前 shardId 超出范围，需要重新分配
		xlog.Msg("current shardId out of range, rebalancing").
			Extras("instance", b.instanceId,
				"currentShardId", currentShard.ShardId,
				"maxShardId", len(shards)-1).
			Infox(ctx)
		if err := b.tryClaimShard(ctx); err != nil {
			xlog.Msg("rebalance failed").
				Extras("instance", b.instanceId).
				Err(err).
				Errorx(ctx)
		}
		return
	}

	expectedShard := shards[currentShard.ShardId]
	if currentShard.Start != expectedShard.Start || currentShard.End != expectedShard.End {
		xlog.Msg("shard range changed, rebalancing").
			Extras("instance", b.instanceId,
				"oldShard", currentShard.String(),
				"newShard", expectedShard.String()).
			Infox(ctx)

		// 范围变化意味着实例数量变化，需要重新分配
		if err := b.tryClaimShard(ctx); err != nil {
			xlog.Msg("rebalance failed").
				Extras("instance", b.instanceId).
				Err(err).
				Errorx(ctx)
		}
	}
}
