package shard

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// etcd key 前缀
	instanceKeyPrefix = "/conductor/instances/"

	// lease TTL (秒)
	leaseTTLSec = 10
)

// Registry 实例注册管理器
type Registry struct {
	cli        *clientv3.Client
	instanceId string
	leaseId    clientv3.LeaseID
	quitCh     chan struct{}
}

func NewRegistry(cli *clientv3.Client, instanceId string, quitCh chan struct{}) *Registry {
	return &Registry{
		cli:        cli,
		instanceId: instanceId,
		quitCh:     quitCh,
	}
}

// InstanceKey 生成实例注册的 key
func (r *Registry) InstanceKey() string {
	return fmt.Sprintf("%s%d", instanceKeyPrefix, r.leaseId)
}

// Register 注册实例到 etcd
func (r *Registry) Register(ctx context.Context) error {
	// 创建 lease
	resp, err := r.cli.Grant(ctx, leaseTTLSec)
	if err != nil {
		return fmt.Errorf("grant lease failed: %w", err)
	}
	r.leaseId = resp.ID

	// 注册实例，key 使用 leaseId，value 存储 instanceId
	key := r.InstanceKey()
	_, err = r.cli.Put(ctx, key, r.instanceId, clientv3.WithLease(r.leaseId))
	if err != nil {
		return fmt.Errorf("register instance failed: %w", err)
	}

	// 保活
	ch, err := r.cli.KeepAlive(ctx, r.leaseId)
	if err != nil {
		return fmt.Errorf("keepalive failed: %w", err)
	}

	// 消费 keepalive 响应
	go func() {
		defer func() {
			if p := recover(); p != nil {
				xlog.Msg("keepalive goroutine panic").
					Extras("instance", r.instanceId, "leaseId", r.leaseId, "panic", p).
					Errorx(ctx)
			}
		}()

		for {
			select {
			case <-r.quitCh:
				return
			case _, ok := <-ch:
				if !ok {
					xlog.Msg("keepalive channel closed").
						Extras("instance", r.instanceId, "leaseId", r.leaseId).
						Errorx(ctx)
					return
				}
			}
		}
	}()

	xlog.Msg("instance registered").
		Extras("instance", r.instanceId, "leaseId", r.leaseId, "key", key).
		Infox(ctx)

	return nil
}

// Unregister 注销实例
func (r *Registry) Unregister(ctx context.Context) {
	// 删除实例注册
	key := r.InstanceKey()
	_, _ = r.cli.Delete(ctx, key)

	// 撤销 lease
	if r.leaseId != 0 {
		_, _ = r.cli.Revoke(ctx, r.leaseId)
	}

	xlog.Msg("instance unregistered").
		Extras("instance", r.instanceId, "leaseId", r.leaseId).
		Infox(ctx)
}

// GetInstanceCount 获取当前实例数量
func (r *Registry) GetInstanceCount(ctx context.Context) (int, error) {
	resp, err := r.cli.Get(ctx, instanceKeyPrefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, err
	}
	return int(resp.Count), nil
}

// GetAllInstances 获取所有活跃实例（返回实例的 key，即 /conductor/instances/{leaseId}）
func (r *Registry) GetAllInstances(ctx context.Context) (map[string]struct{}, error) {
	resp, err := r.cli.Get(ctx, instanceKeyPrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	instances := make(map[string]struct{}, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		// 使用 key 作为实例标识
		instances[string(kv.Key)] = struct{}{}
	}
	return instances, nil
}

// LeaseId 获取 lease ID
func (r *Registry) LeaseId() clientv3.LeaseID {
	return r.leaseId
}

// Watch 监听实例变化
func (r *Registry) Watch(ctx context.Context) clientv3.WatchChan {
	return r.cli.Watch(ctx, instanceKeyPrefix, clientv3.WithPrefix())
}
