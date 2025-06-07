package xmap

import (
	"fmt"
	"hash/fnv"
)

type hasher[K comparable] func(K) uint64

type ShardedMap[K comparable, V any] struct {
	shards    uint64
	stores    []*SyncMap[K, V]
	keyHasher hasher[K]
}

func defaultHasher[K comparable]() hasher[K] {
	return func(k K) uint64 {
		h := fnv.New64a()
		h.Write([]byte(fmt.Sprint(k)))
		return h.Sum64()
	}
}

// shards 指定分片数量
func NewShardedMap[K comparable, V any](shards int) *ShardedMap[K, V] {
	return NewShardedMapWithHasher[K, V](shards, defaultHasher[K]())
}

func NewShardedMapWithHasher[K comparable, V any](shards int, hasher hasher[K]) *ShardedMap[K, V] {
	if shards < 1 {
		shards = 16
	}

	if hasher == nil {
		hasher = defaultHasher[K]()
	}

	shardMap := make([]*SyncMap[K, V], shards)
	for i := range shards {
		shardMap[i] = NewSyncMap[K, V]()
	}

	return &ShardedMap[K, V]{
		shards:    uint64(shards),
		stores:    shardMap,
		keyHasher: hasher,
	}
}

func (sm *ShardedMap[K, V]) Keys() []K {
	var keys []K
	for _, shard := range sm.stores {
		keys = append(keys, shard.Keys()...)
	}
	return keys
}

func (sm *ShardedMap[K, V]) Values() []V {
	var values []V
	for _, shard := range sm.stores {
		values = append(values, shard.Values()...)
	}
	return values
}

func (sm *ShardedMap[K, V]) Get(key K) V {
	return sm.getShard(key).Get(key)
}

func (sm *ShardedMap[K, V]) Put(key K, value V) {
	sm.getShard(key).Put(key, value)
}

func (sm *ShardedMap[K, V]) Remove(key K) V {
	return sm.getShard(key).Remove(key)
}

func (sm *ShardedMap[K, V]) Size() int {
	var size int
	for _, shard := range sm.stores {
		size += shard.Size()
	}
	return size
}

func (sm *ShardedMap[K, V]) Has(key K) bool {
	return sm.getShard(key).Has(key)
}

func (sm *ShardedMap[K, V]) Empty() bool {
	for _, shard := range sm.stores {
		if !shard.Empty() {
			return false
		}
	}
	return true
}

func (sm *ShardedMap[K, V]) Clear() {
	for _, shard := range sm.stores {
		shard.Clear()
	}
}

func (sm *ShardedMap[K, V]) getShard(key K) *SyncMap[K, V] {
	return sm.stores[sm.keyHasher(key)%sm.shards]
}
