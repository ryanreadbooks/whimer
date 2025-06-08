package xmap

import "sync"

type SyncMap[K comparable, V any] struct {
	sync.RWMutex
	store map[K]V
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		store: make(map[K]V),
	}
}

func (sm *SyncMap[K, V]) Keys() []K {
	sm.RLock()
	defer sm.RUnlock()
	keys := make([]K, 0, len(sm.store))
	for key := range sm.store {
		keys = append(keys, key)
	}

	return keys
}

func (sm *SyncMap[K, V]) Values() []V {
	sm.RLock()
	defer sm.RUnlock()
	values := make([]V, 0, len(sm.store))
	for _, value := range sm.store {
		values = append(values, value)
	}

	return values
}

func (sm *SyncMap[K, V]) Get(key K) V {
	sm.RLock()
	defer sm.RUnlock()

	return sm.store[key]
}

func (sm *SyncMap[K, V]) Put(key K, value V) {
	sm.Lock()
	defer sm.Unlock()
	sm.store[key] = value
}

func (sm *SyncMap[K, V]) Remove(key K) V {
	sm.Lock()
	defer sm.Unlock()
	value, ok := sm.store[key]
	if ok {
		delete(sm.store, key)
	}

	return value
}

func (sm *SyncMap[K, V]) Size() int {
	sm.RLock()
	defer sm.RUnlock()

	return len(sm.store)
}

func (sm *SyncMap[K, V]) Empty() bool {
	sm.RLock()
	defer sm.RUnlock()

	return len(sm.store) == 0
}

func (sm *SyncMap[K, V]) Clear() {
	sm.Lock()
	defer sm.Unlock()
	sm.store = make(map[K]V)
}

func (sm *SyncMap[K, V]) Has(key K) bool {
	sm.RLock()
	defer sm.RUnlock()
	_, ok := sm.store[key]
	
	return ok
}
