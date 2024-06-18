package synccollections

import "sync"

// https://juejin.cn/post/7011355673069879304
// https://zhuanlan.zhihu.com/p/388444013
// 这2个测试结果应该还是比较好的  所以写多的情况下(简单来说是写多的情况)应该用RWMutexMap而不是SyncMap

// Map
//
//	20230328	/src/sync/map_bench_test.go   根据这测试Map意外的好  看来几乎没有情况要用rwmutexmap了
//	                          name   copy on write map  rwmutex map          map
//
// 0                 LoadMostlyHits       6.463 ns/op  43.40 ns/op  6.134 ns/op
// 1               LoadMostlyMisses       5.512 ns/op  53.78 ns/op  6.652 ns/op
// 2            LoadOrStoreBalanced          19 ns/op   85.6 ns/op   28.0 ns/op
// 3              LoadOrStoreUnique          47 ns/op   42.3 ns/op   21.2 ns/op
// 4           LoadOrStoreCollision       2.991 ns/op  68.37 ns/op  3.124 ns/op
// 5          LoadAndDeleteBalanced        34.5 ns/op  70.39 ns/op  6.087 ns/op
// 6            LoadAndDeleteUnique        43.7 ns/op  76.12 ns/op  6.494 ns/op
// 7         LoadAndDeleteCollision        82.0 ns/op  68.77 ns/op  1.901 ns/op
// 8                          Range          29 ns/op     16 ns/op     85 ns/op
// 9               AdversarialAlloc          86 ns/op  64.77 ns/op   95.9 ns/op
// 10             AdversarialDelete        84.4 ns/op  82.19 ns/op  52.48 ns/op
// 11               DeleteCollision        17.7 ns/op  62.97 ns/op  3.525 ns/op
// 12                 SwapCollision        14.8 ns/op  98.14 ns/op   17.4 ns/op
// 13                SwapMostlyHits          76 ns/op   54.9 ns/op  24.95 ns/op
// 14              SwapMostlyMisses          35 ns/op   38.6 ns/op   69.7 ns/op  x
// 15       CompareAndSwapCollision       3.642 ns/op  82.75 ns/op  7.159 ns/op
// 16   CompareAndSwapNoExistingKey       5.608 ns/op  73.59 ns/op  6.180 ns/op
// 17   CompareAndSwapValueNotEqual       3.391 ns/op  73.54 ns/op  3.793 ns/op
// 18      CompareAndSwapMostlyHits       78247 ns/op  181.2 ns/op  28.31 ns/op
// 19    CompareAndSwapMostlyMisses       14.22 ns/op  117.4 ns/op  13.08 ns/op
// 20     CompareAndDeleteCollision       1.900 ns/op  62.05 ns/op  4.706 ns/op
// 21    CompareAndDeleteMostlyHits      147829 ns/op  297.1 ns/op  36.65 ns/op
// 22  CompareAndDeleteMostlyMisses       10.15 ns/op  90.41 ns/op  9.881 ns/op
//
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
	if a != nil {
		actual = a.(V)
	}
	return
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	a, loaded := m.mp.LoadAndDelete(key)
	if a != nil {
		value = a.(V)
	}
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

func (m *Map[K, V]) RangeAll(f func(key K, value V)) {
	// useful for me . but make Map[K,V] different with sync.Map
	m.Range(func(key K, value V) bool {
		f(key, value)
		return true
	})
}

func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	p, loaded := m.mp.Swap(key, value)
	if loaded {
		return p.(V), loaded
	}
	return
}

func (m *Map[K, V]) CompareAndSwap(key K, old, newV V) (swapped bool) {
	return m.mp.CompareAndSwap(key, old, newV)
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.mp.CompareAndDelete(key, old)
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
	return
}

func (m *RWMutexMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	m.mu.Lock()
	value, loaded = m.dirty[key]
	if !loaded {
		m.mu.Unlock()
		return
	}
	delete(m.dirty, key)
	m.mu.Unlock()
	return
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

func (m *RWMutexMap[K, V]) RangeAll(f func(key K, value V)) {
	// useful for me . but make Map[K,V] different with sync.Map
	m.Range(func(key K, value V) bool {
		f(key, value)
		return true
	})
}

func (m *RWMutexMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	m.mu.Lock()
	if m.dirty == nil {
		m.dirty = make(map[K]V)
	}

	previous, loaded = m.dirty[key]
	m.dirty[key] = value
	m.mu.Unlock()
	return
}

func (m *RWMutexMap[K, V]) CompareAndSwap(key K, old, newV V) (swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dirty == nil {
		return false
	}

	value, loaded := m.dirty[key]
	if loaded && eq(value, old) {
		m.dirty[key] = newV
		return true
	}
	return false
}

func (m *RWMutexMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dirty == nil {
		return false
	}

	value, loaded := m.dirty[key]
	if loaded && eq(value, old) {
		delete(m.dirty, key)
		return true
	}
	return false
}

func eq(a, b any) bool {
	return a == b
}
