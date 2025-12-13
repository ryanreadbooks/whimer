package shard

import (
	"context"
	"fmt"
	"testing"
	"time"

	etcdv3 "go.etcd.io/etcd/client/v3"
)

// 测试前需要启动本地 etcd: etcd --listen-client-urls http://localhost:2379
const testEtcdEndpoint = "localhost:2379"

func newTestClient(t *testing.T) *etcdv3.Client {
	cli, err := etcdv3.New(etcdv3.Config{
		Endpoints:   []string{testEtcdEndpoint},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		t.Skipf("skip test: cannot connect to etcd: %v", err)
	}
	return cli
}

func cleanupPrefix(t *testing.T, cli *etcdv3.Client, prefix string) {
	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	_, err := cli.Delete(ctx, prefix, etcdv3.WithPrefix())
	if err != nil {
		t.Logf("cleanup prefix %s failed: %v", prefix, err)
	}
}

// 测试单个节点启动和停止
func TestManager_SingleNode(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	prefix := "/test/shard/single"
	cleanupPrefix(t, cli, prefix)
	defer cleanupPrefix(t, cli, prefix)

	ctx := t.Context()

	m := NewManager(cli, prefix, "node1", WithTTL(5))

	err := m.Start(ctx)
	if err != nil {
		t.Fatalf("start manager failed: %v", err)
	}

	// 等待分片信息更新
	time.Sleep(100 * time.Millisecond)

	shard := m.GetShard()
	t.Logf("shard: %+v", shard)

	if !shard.Active {
		t.Error("shard should be active")
	}
	if shard.Total != 1 {
		t.Errorf("shard.Total = %d, want 1", shard.Total)
	}
	if shard.Index != 0 {
		t.Errorf("shard.Index = %d, want 0", shard.Index)
	}

	m.Stop()
}

// 测试多节点分片
func TestManager_MultipleNodes(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	prefix := "/test/shard/multi"
	cleanupPrefix(t, cli, prefix)
	defer cleanupPrefix(t, cli, prefix)

	ctx := t.Context()

	nodeCount := 3
	managers := make([]*Manager, nodeCount)

	// 依次启动节点
	for i := 0; i < nodeCount; i++ {
		nodeId := fmt.Sprintf("node%d", i)
		m := NewManager(cli, prefix, nodeId, WithTTL(10))
		err := m.Start(ctx)
		if err != nil {
			t.Fatalf("start manager %s failed: %v", nodeId, err)
		}
		managers[i] = m

		// 等待 watch 事件传播
		time.Sleep(200 * time.Millisecond)
	}

	// 等待所有节点的分片信息同步
	time.Sleep(500 * time.Millisecond)

	// 验证每个节点的分片信息
	for i, m := range managers {
		shard := m.GetShard()
		t.Logf("node%d shard: %+v", i, shard)

		if !shard.Active {
			t.Errorf("node%d should be active", i)
		}
		if shard.Total != nodeCount {
			t.Errorf("node%d shard.Total = %d, want %d", i, shard.Total, nodeCount)
		}
		// 由于 key 按字典序排序，node0 < node1 < node2
		if shard.Index != i {
			t.Errorf("node%d shard.Index = %d, want %d", i, shard.Index, i)
		}
	}

	// 停止所有节点
	for _, m := range managers {
		m.Stop()
	}
}

// 测试节点退出后其他节点感知
func TestManager_NodeLeave(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	prefix := "/test/shard/leave"
	cleanupPrefix(t, cli, prefix)
	defer cleanupPrefix(t, cli, prefix)

	ctx := t.Context()

	// 启动两个节点
	m1 := NewManager(cli, prefix, "node1", WithTTL(5))
	m2 := NewManager(cli, prefix, "node2", WithTTL(5))

	err := m1.Start(ctx)
	if err != nil {
		t.Fatalf("start m1 failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	err = m2.Start(ctx)
	if err != nil {
		t.Fatalf("start m2 failed: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 验证两个节点都感知到对方
	shard1 := m1.GetShard()
	shard2 := m2.GetShard()
	t.Logf("before leave - m1: %+v, m2: %+v", shard1, shard2)

	if shard1.Total != 2 || shard2.Total != 2 {
		t.Errorf("both nodes should see 2 total, got m1=%d, m2=%d", shard1.Total, shard2.Total)
	}

	// 停止 m2
	m2.Stop()
	t.Log("m2 stopped")

	// 等待 m1 感知到 m2 退出（lease 过期或 watch 到删除事件）
	time.Sleep(1 * time.Second)

	shard1After := m1.GetShard()
	t.Logf("after m2 leave - m1: %+v", shard1After)

	if shard1After.Total != 1 {
		t.Errorf("m1 should see 1 total after m2 leave, got %d", shard1After.Total)
	}
	if shard1After.Index != 0 {
		t.Errorf("m1 index should be 0 after m2 leave, got %d", shard1After.Index)
	}

	m1.Stop()
}

// 测试节点动态加入
func TestManager_NodeJoin(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	prefix := "/test/shard/join"
	cleanupPrefix(t, cli, prefix)
	defer cleanupPrefix(t, cli, prefix)

	ctx := t.Context()

	// 先启动一个节点
	m1 := NewManager(cli, prefix, "node1", WithTTL(10))
	err := m1.Start(ctx)
	if err != nil {
		t.Fatalf("start m1 failed: %v", err)
	}
	time.Sleep(200 * time.Millisecond)

	shard1 := m1.GetShard()
	t.Logf("m1 initial: %+v", shard1)
	if shard1.Total != 1 {
		t.Errorf("m1 should see 1 total, got %d", shard1.Total)
	}

	// 加入新节点
	m2 := NewManager(cli, prefix, "node2", WithTTL(10))
	err = m2.Start(ctx)
	if err != nil {
		t.Fatalf("start m2 failed: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 验证 m1 感知到 m2 加入
	shard1After := m1.GetShard()
	shard2 := m2.GetShard()
	t.Logf("after m2 join - m1: %+v, m2: %+v", shard1After, shard2)

	if shard1After.Total != 2 {
		t.Errorf("m1 should see 2 total, got %d", shard1After.Total)
	}
	if shard2.Total != 2 {
		t.Errorf("m2 should see 2 total, got %d", shard2.Total)
	}

	m1.Stop()
	m2.Stop()
}

// 测试分片计算正确性
func TestManager_ShardCalculation(t *testing.T) {
	cli := newTestClient(t)
	defer cli.Close()

	prefix := "/test/shard/calc"
	cleanupPrefix(t, cli, prefix)
	defer cleanupPrefix(t, cli, prefix)

	ctx := t.Context()

	// 模拟 5 个节点
	nodeIds := []string{"a", "b", "c", "d", "e"} // 按字典序
	managers := make([]*Manager, len(nodeIds))

	for i, id := range nodeIds {
		m := NewManager(cli, prefix, id, WithTTL(10))
		err := m.Start(ctx)
		if err != nil {
			t.Fatalf("start manager %s failed: %v", id, err)
		}
		managers[i] = m
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	// 验证分片索引
	for i, m := range managers {
		shard := m.GetShard()
		t.Logf("node %s: index=%d, total=%d, keys=%v", nodeIds[i], shard.Index, shard.Total, shard.Keys)

		if shard.Index != i {
			t.Errorf("node %s index = %d, want %d", nodeIds[i], shard.Index, i)
		}
		if shard.Total != len(nodeIds) {
			t.Errorf("node %s total = %d, want %d", nodeIds[i], shard.Total, len(nodeIds))
		}
	}

	// 清理
	for _, m := range managers {
		m.Stop()
	}
}
