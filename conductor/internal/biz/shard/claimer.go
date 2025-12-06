package shard

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// etcd key 前缀
	shardKeyPrefix = "/conductor/shards/"
)

// Claimer 分片抢占器
type Claimer struct {
	cli        *clientv3.Client
	instanceId string
	registry   *Registry
}

func NewClaimer(cli *clientv3.Client, instanceId string, registry *Registry) *Claimer {
	return &Claimer{
		cli:        cli,
		instanceId: instanceId,
		registry:   registry,
	}
}

// holder 获取当前实例的 holder 标识（使用注册的 key）
func (c *Claimer) holder() string {
	return c.registry.InstanceKey()
}

// Claim 抢占单个分片
func (c *Claimer) Claim(ctx context.Context, shard Range) (bool, error) {
	key := fmt.Sprintf("%s%d", shardKeyPrefix, shard.ShardId)

	// 构建 value，holder 使用实例注册的 key
	val := &Value{
		Holder: c.holder(),
		Range:  shard,
	}

	// 使用事务进行原子抢占
	// 只有当 key 不存在时才能创建
	txn := c.cli.Txn(ctx)
	txn = txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0))
	txn = txn.Then(clientv3.OpPut(key, val.Marshal(), clientv3.WithLease(c.registry.LeaseId())))
	txn = txn.Else(clientv3.OpGet(key))

	resp, err := txn.Commit()
	if err != nil {
		return false, err
	}

	if resp.Succeeded {
		return true, nil
	}

	// 检查是否已经被自己持有
	if len(resp.Responses) > 0 && len(resp.Responses[0].GetResponseRange().Kvs) > 0 {
		data := resp.Responses[0].GetResponseRange().Kvs[0].Value
		sv, err := unmarshalValue(data)
		if err == nil && sv.Holder == c.holder() {
			return true, nil
		}
	}

	return false, nil
}

// Release 释放分片
func (c *Claimer) Release(ctx context.Context, shard *Range) {
	if shard == nil {
		return
	}

	key := fmt.Sprintf("%s%d", shardKeyPrefix, shard.ShardId)

	// 获取当前值检查是否是自己持有
	resp, err := c.cli.Get(ctx, key)
	if err != nil || len(resp.Kvs) == 0 {
		return
	}

	sv, err := unmarshalValue(resp.Kvs[0].Value)
	if err != nil || sv.Holder != c.holder() {
		return
	}

	// 只删除自己持有的
	_, err = c.cli.Delete(ctx, key)
	if err != nil {
		xlog.Msg("release shard failed").
			Extras("instance", c.instanceId, "shardId", shard.ShardId).
			Err(err).
			Errorx(ctx)
		return
	}

	xlog.Msg("shard released").
		Extras("instance", c.instanceId, "shard", shard.String()).
		Infox(ctx)
}

// Update 更新分片值
func (c *Claimer) Update(ctx context.Context, shard Range) {
	key := fmt.Sprintf("%s%d", shardKeyPrefix, shard.ShardId)
	val := &Value{
		Holder: c.holder(),
		Range:  shard,
	}

	_, err := c.cli.Put(ctx, key, val.Marshal(), clientv3.WithLease(c.registry.LeaseId()))
	if err != nil {
		xlog.Msg("update shard value failed").
			Extras("instance", c.instanceId, "shardId", shard.ShardId).
			Err(err).
			Errorx(ctx)
	}
}

// CleanupStale 清理过期的分片
func (c *Claimer) CleanupStale(ctx context.Context, currentInstanceCount int) {
	// 获取所有分片
	resp, err := c.cli.Get(ctx, shardKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		xlog.Msg("get shards failed").
			Extras("instance", c.instanceId).
			Err(err).
			Errorx(ctx)
		return
	}

	// 获取所有活跃实例（key 集合）
	activeInstances, err := c.registry.GetAllInstances(ctx)
	if err != nil {
		xlog.Msg("get instances failed").
			Extras("instance", c.instanceId).
			Err(err).
			Errorx(ctx)
		return
	}

	// 删除不属于任何活跃实例的分片
	for _, kv := range resp.Kvs {
		sv, err := unmarshalValue(kv.Value)
		if err != nil {
			continue
		}
		if _, ok := activeInstances[sv.Holder]; !ok {
			_, _ = c.cli.Delete(ctx, string(kv.Key))
			xlog.Msg("cleaned up stale shard").
				Extras("instance", c.instanceId, "key", string(kv.Key), "holder", sv.Holder).
				Infox(ctx)
		}
	}

	// 删除超出当前实例数量的分片
	for _, kv := range resp.Kvs {
		var shardId int
		fmt.Sscanf(string(kv.Key), shardKeyPrefix+"%d", &shardId)
		if shardId >= currentInstanceCount {
			_, _ = c.cli.Delete(ctx, string(kv.Key))
			xlog.Msg("cleaned up excess shard").
				Extras("instance", c.instanceId, "key", string(kv.Key)).
				Infox(ctx)
		}
	}
}

// GetShardCount 获取当前分片数量
func (c *Claimer) GetShardCount(ctx context.Context) (int, error) {
	resp, err := c.cli.Get(ctx, shardKeyPrefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

// Watch 监听分片变化
func (c *Claimer) Watch(ctx context.Context) clientv3.WatchChan {
	return c.cli.Watch(ctx, shardKeyPrefix, clientv3.WithPrefix())
}
