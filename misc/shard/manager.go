package shard

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

// 注册当前实例 + 监听其它实例 并更新本地分片信息
//
// 创建租约 -> 续约租约
// 监听key -> 更新本地分片信息
type Manager struct {
	keyPrefix string
	keyId     string
	fullKey   string // keyPrefix + keyId
	opt       *option
	wg        sync.WaitGroup
	quitCh    chan struct{}

	cli *etcdv3.Client

	leaseMu sync.RWMutex
	lease   etcdv3.Lease
	leaseId etcdv3.LeaseID

	// 记录当前数据的版本号
	revisionMu sync.RWMutex
	revision   int64

	// cur shard info (atomic for lock-free read)
	shard atomic.Pointer[Shard]
}

func registeredKey(prefix, instance string) string {
	return prefix + "/" + instance
}

func NewManager(
	cli *etcdv3.Client,
	keyPrefix string,
	keyId string,
	opts ...Option,
) *Manager {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	m := &Manager{
		keyPrefix: keyPrefix,
		keyId:     keyId,
		fullKey:   registeredKey(keyPrefix, keyId),
		opt:       opt,
		cli:       cli,
		quitCh:    make(chan struct{}),
	}
	m.shard.Store(&Shard{})
	return m
}

func (m *Manager) logger(msg string) *xlog.LogItem {
	m.revisionMu.RLock()
	rev := m.revision
	m.revisionMu.RUnlock()
	return xlog.Msg(msg).
		Extra("key_prefix", m.keyPrefix).
		Extra("key_id", m.keyId).
		Extra("cur_revision", rev)
}

func (m *Manager) Start(ctx context.Context) error {
	err := m.createLease(ctx)
	if err != nil {
		return fmt.Errorf("create lease failed: %w", err)
	}

	// register myself to etcd
	err = m.registerSelf(ctx)
	if err != nil {
		return fmt.Errorf("register myself to etcd failed: %w", err)
	}

	// keepalive
	m.goKeepAlive(ctx)

	// watch key prefix with revision
	m.goWatch(ctx)

	return nil
}

func (m *Manager) Stop() {
	close(m.quitCh)
	m.wg.Wait()
}

func (m *Manager) createLease(ctx context.Context) error {
	m.leaseMu.Lock()
	defer m.leaseMu.Unlock()

	lease := etcdv3.NewLease(m.cli)
	leaseResp, err := lease.Grant(ctx, m.opt.ttlSec)
	if err != nil {
		return fmt.Errorf("etcd grant lease failed: %w", err)
	}

	if m.lease != nil {
		m.lease.Close()
	}

	m.lease = lease
	m.leaseId = leaseResp.ID

	return nil
}

// 注册当前实例到特定租约下
func (m *Manager) registerSelf(ctx context.Context) error {
	_, err := m.cli.Put(ctx, m.fullKey, m.keyId, etcdv3.WithLease(m.leaseId))
	if err != nil {
		return fmt.Errorf("etcd put failed: %w", err)
	}

	// 启动时首先拿一次shard
	m.fetchAndUpdateShard(ctx)

	return nil
}

func (m *Manager) goKeepAlive(ctx context.Context) {
	m.wg.Add(1)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:             "misc.shard_manager.keepalive",
		InheritCtxCancel: true,
		Job: func(ctx context.Context) error {
			defer m.wg.Done()
			m.doKeepAlive(ctx)
			return nil
		},
	})
}

func (m *Manager) doKeepAlive(ctx context.Context) {
	var (
		quit         bool
		ctxErr       error
		shouldReturn bool
	)

	defer func() {
		if quit || ctxErr != nil {
			m.logger("shard manager keepalive quit").
				Extra("quit", quit).
				Extra("ctx_err", ctxErr).
				Infox(ctx)
			m.revokeLease()
		}
	}()

	for {
		quit, ctxErr, shouldReturn = m.selectQuit(ctx)
		if shouldReturn {
			break
		}

		// 阻塞keepalive
		// 退出表明keepalive失败 需要重建租约
		m.keepAliveLoop(ctx)

		// keepalive 断开后需要重建租约
		select {
		case <-m.quitCh:
			quit = true
			return
		case <-ctx.Done():
			ctxErr = ctx.Err()
			return
		case <-time.After(time.Duration(m.opt.keepaliveRetryWaitSec) * time.Second): // 重试等待
		}

		m.logger("shard manager recreating lease").Infox(ctx)
		if err := m.createLease(ctx); err != nil {
			m.logger("shard manager recreate lease failed").Err(err).Errorx(ctx)
			continue
		}
		if err := m.registerSelf(ctx); err != nil {
			m.logger("shard manager re-register failed").Err(err).Errorx(ctx)
			continue
		}
		m.logger("shard manager lease recreated").Infox(ctx)
	}
}

func (m *Manager) keepAliveLoop(ctx context.Context) {
	m.leaseMu.RLock()
	leaseId := m.leaseId
	m.leaseMu.RUnlock()

	ch, err := m.cli.KeepAlive(ctx, leaseId)
	if err != nil {
		m.logger("shard manager start keepalive failed").Err(err).Errorx(ctx)
		return
	}

	m.logger("shard manager keepalive started").Extra("lease_id", leaseId).Infox(ctx)

	// block here forever
	for {
		select {
		case <-m.quitCh:
			return
		case <-ctx.Done():
			return
		case resp, ok := <-ch:
			if !ok {
				m.logger("shard manager keepalive channel closed").Infox(ctx)
				return
			}
			if resp == nil {
				m.logger("shard manager lease expired cuz resp is nil").Infox(ctx)
				return
			}
			// 续约成功
		}
	}
}

// 撤回租约 用于退出时使用
func (m *Manager) revokeLease() {
	m.leaseMu.RLock()
	leaseId := m.leaseId
	m.leaseMu.RUnlock()

	if leaseId == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.cli.Revoke(ctx, leaseId)
	if err != nil {
		m.logger("shard manager revoke lease failed").Err(err).Error()
		return
	}
	m.logger("shard manager lease revoked").Extra("lease_id", leaseId).Info()
}

func (m *Manager) goWatch(ctx context.Context) {
	m.wg.Add(1)

	// 监听其它实例变化 并且更新自身instance
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:             "misc.shard_manager.watch",
		InheritCtxCancel: true,
		Job: func(ctx context.Context) error {
			defer m.wg.Done()

			m.doWatch(ctx)
			return nil
		},
	})
}

func (m *Manager) doWatch(ctx context.Context) {
	var (
		quit         bool
		ctxErr       error
		shouldReturn bool
	)

	defer func() {
		if quit || ctxErr != nil {
			m.logger("shard manager watch quit").
				Extra("quit", quit).
				Extra("ctx_err", ctxErr).
				Infox(ctx)
		}
	}()

	// 包装ctx 防止被某个etcd分片阻塞
	watchCtx := etcdv3.WithRequireLeader(ctx)
	for {
		quit, ctxErr, shouldReturn = m.selectQuit(ctx)
		if shouldReturn {
			break
		}

		m.revisionMu.RLock()
		revision := m.revision
		m.revisionMu.RUnlock()
		ch := m.cli.Watch(watchCtx, m.keyPrefix, etcdv3.WithPrefix(), etcdv3.WithRev(revision+1))
		for resp := range ch { // will block here until error or quit
			quit, ctxErr, shouldReturn = m.selectQuit(ctx)
			if shouldReturn {
				break
			}

			if err := resp.Err(); err != nil {
				m.logger("shard manager watch error").
					Err(err).
					Extra("withrev", revision+1).
					Errorx(ctx)
				// 出错重新watch
				break
			}

			if resp.Header.Revision > revision {
				m.fetchAndUpdateShard(ctx)
			}
		}

		// watch 断开后等待再重试
		select {
		case <-m.quitCh:
			quit = true
			return
		case <-ctx.Done():
			ctxErr = ctx.Err()
			return
		case <-time.After(time.Second):
		}
	}
}

// 去etcd Get最新的实例分片
func (m *Manager) fetchAndUpdateShard(ctx context.Context) {
	// 拿实例前缀所有key
	resp, err := m.cli.Get(ctx, m.keyPrefix,
		etcdv3.WithPrefix(),
		etcdv3.WithSort(etcdv3.SortByKey, etcdv3.SortAscend)) // 升序排序key
	if err != nil {
		m.logger("shard manager get shard failed").Err(err).Errorx(ctx)
		return
	}

	// 更新当前实例的shard信息 主要是更新total和index即可
	keys := make([]string, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		keys = append(keys, string(kv.Key))
	}

	newShard := &Shard{
		Keys:         keys,
		LastUpdateAt: time.Now(),
	}

	selfExists := false
	for idx, k := range keys {
		if k == m.fullKey {
			newShard.Index = idx
			newShard.Total = len(keys)
			newShard.Active = true
			selfExists = true
			break
		}
	}
	if !selfExists {
		m.logger("shard manager not found and considered as offline").Infox(ctx)
		newShard.Active = false
	}

	if !m.shard.Load().Equal(newShard) {
		m.logger("shard manager shard changed").Extra("new_shard", newShard).Infox(ctx)
	}

	m.shard.Store(newShard)

	m.revisionMu.Lock()
	m.revision = resp.Header.Revision
	m.revisionMu.Unlock()
}

func (m *Manager) GetShard() Shard {
	if s := m.shard.Load(); s != nil {
		return *s
	}
	return Shard{}
}

func (m *Manager) selectQuit(ctx context.Context) (quit bool, err error, shouldBreak bool) {
	select {
	case <-m.quitCh:
		quit = true
		shouldBreak = true
		return
	case <-ctx.Done():
		err = ctx.Err()
		shouldBreak = true
		return
	default:
	}

	return
}
