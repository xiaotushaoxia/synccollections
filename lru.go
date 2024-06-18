package synccollections

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// 但是我为了让它的零值也是可以直接使用的，就只能这样凑合一下了
const defaultLruCap int32 = 100

// Deprecated: bad implement
type LRU[K comparable, V any] struct {
	cache map[K]*list.Element
	list  list.List
	cap   atomic.Int32
	m     sync.RWMutex
	once  sync.Once
}

type node[K comparable, V any] struct {
	k K
	v V
}

func MakeLRU[K comparable, V any](cap int) *LRU[K, V] {
	a := &LRU[K, V]{cache: map[K]*list.Element{}}
	if cap < 1 {
		cap = 1
	}
	a.cap.Store(int32(cap))
	return a
}

func (l *LRU[K, V]) Set(key K, value V) {
	l.once.Do(l.lazyInit)
	l.m.Lock()
	defer l.m.Unlock()
	if e, ok := l.cache[key]; ok {
		e.Value.(*node[K, V]).v = value
		l.list.MoveToFront(e)
		return
	}

	l.cache[key] = l.list.PushFront(&node[K, V]{k: key, v: value})
	l.tryRemove()
}

func (l *LRU[K, V]) Get(key K) (v V, ok bool) {
	l.m.Lock() // Get会修改list  所以要用写锁
	defer l.m.Unlock()
	element, ok := l.cache[key]
	if ok {
		l.list.MoveToFront(element)
		return element.Value.(*node[K, V]).v, true
	}
	return
}

func (l *LRU[K, V]) Delete(key K) {
	l.m.Lock()
	element, ok := l.cache[key]
	if ok {
		l.list.Remove(element)
		delete(l.cache, key)
	}
	l.m.Unlock()
}

func (l *LRU[K, V]) Range(f func(key K, value V) (shouldContinue bool)) {
	l.m.RLock()
	defer l.m.RUnlock()
	for e := l.list.Back(); e != nil; e = e.Prev() {
		n := e.Value.(*node[K, V])
		if !f(n.k, n.v) {
			return
		}
	}
}

func (l *LRU[K, V]) Keys() []K {
	l.m.RLock()
	defer l.m.RUnlock()
	var ks []K
	for e := l.list.Back(); e != nil; e = e.Prev() {
		n := e.Value.(*node[K, V])
		ks = append(ks, n.k)
	}
	return ks
}

func (l *LRU[K, V]) Contains(key K) bool {
	l.m.RLock()
	_, ok := l.cache[key]
	l.m.RUnlock()
	return ok
}

func (l *LRU[K, V]) Peek(key K) (v V, ok bool) {
	l.m.RLock()
	defer l.m.RUnlock()
	element, ok := l.cache[key]
	if ok {
		return element.Value.(*node[K, V]).v, true
	}
	return
}

func (l *LRU[K, V]) GetSize() int {
	l.m.RLock()
	defer l.m.RUnlock()
	return len(l.cache)
}

func (l *LRU[K, V]) SetCap(cap int) {
	if cap < 1 {
		cap = 1
	}
	l.m.Lock()
	l.cap.Store(int32(cap))
	for l.tryRemove() {
	}
	l.m.Unlock()
}

func (l *LRU[K, V]) GetCap() int {
	return int(l.cap.Load())
}

func (l *LRU[K, V]) tryRemove() (removed bool) {
	if l.list.Len() > int(l.cap.Load()) {
		back := l.list.Back()
		l.list.Remove(back)
		delete(l.cache, back.Value.(*node[K, V]).k)
		return true
	}
	return false
}

func (l *LRU[K, V]) lazyInit() {
	if cap_ := l.cap.Load(); cap_ == 0 {
		l.cap.Store(defaultLruCap)
	}
	if l.cache == nil {
		l.cache = map[K]*list.Element{}
	}
}
