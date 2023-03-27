package synccollections

import "sync"

// https://juejin.cn/post/7011355673069879304
// https://zhuanlan.zhihu.com/p/388444013
// 这2个测试结果应该还是比较好的  所以写多的情况下(简单来说是写多的情况)应该用RWMutexMap而不是SyncMap

// Map
// go文档写 sync.Map优化了这两种情况 1很少写很多读 2读写覆盖多个不同key
// The Map type is optimized for two common use cases: (1) when the entry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a Map may significantly reduce lock
// contention compared to a Go map paired with a separate Mutex or RWMutex.
type Map[K comparable, V any] struct {
	mp sync.Map
}

func (m *Map[K, V]) Load(key K) (value V, ok bool) {
	load, ok := m.mp.Load(key)
	if !ok {
		return
	}
	return load.(V), true
}

func (m *Map[K, V]) Store(key K, value V) {
	m.mp.Store(key, value)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	a, loaded := m.mp.LoadOrStore(key, value)
	actual = a.(V)
	return
}

func (m *Map[K, V]) Delete(key K) {
	m.mp.Delete(key)
}

func (m *Map[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	m.mp.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

// RWMutexMap is an implementation of mapInterface using a sync.RWMutex.
// 来自go标准库里面的一个文件 src/sync/map_reference_test.go   用来读少写少的情况下使用这个还是不错的(比syncmap好)，不知道为啥没有加入标准库

type RWMutexMap[K comparable, V any] struct {
	mu    sync.RWMutex
	dirty map[K]V
}

func (m *RWMutexMap[K, V]) Load(key K) (value V, ok bool) {
	m.mu.RLock()
	value, ok = m.dirty[key]
	m.mu.RUnlock()
	return
}

func (m *RWMutexMap[K, V]) Store(key K, value V) {
	m.mu.Lock()
	if m.dirty == nil {
		m.dirty = make(map[K]V)
	}
	m.dirty[key] = value
	m.mu.Unlock()
}

func (m *RWMutexMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	m.mu.Lock()
	actual, loaded = m.dirty[key]
	if !loaded {
		actual = value
		if m.dirty == nil {
			m.dirty = make(map[K]V)
		}
		m.dirty[key] = value
	}
	m.mu.Unlock()
	return actual, loaded
}

func (m *RWMutexMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	m.mu.Lock()
	value, loaded = m.dirty[key]
	if !loaded {
		m.mu.Unlock()
		var v V
		return v, false
	}
	delete(m.dirty, key)
	m.mu.Unlock()
	return value, loaded
}

func (m *RWMutexMap[K, V]) Delete(key K) {
	m.mu.Lock()
	delete(m.dirty, key)
	m.mu.Unlock()
}

func (m *RWMutexMap[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	m.mu.RLock()
	keys := make([]K, 0, len(m.dirty))
	for k := range m.dirty {
		keys = append(keys, k)
	}
	m.mu.RUnlock()

	for _, k := range keys {
		v, ok := m.Load(k)
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}
