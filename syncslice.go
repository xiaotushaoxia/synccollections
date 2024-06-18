package synccollections

import "sync"

// Deprecated: bad implement
type Slice[T any] struct {
	ts []T
	m  sync.RWMutex
}

func MakeSlice[T any](lengthCap ...int) *Slice[T] {
	var ts []T
	if len(lengthCap) == 1 {
		ts = make([]T, lengthCap[0])
	} else if len(lengthCap) == 2 {
		ts = make([]T, lengthCap[0], lengthCap[1])
	}
	a := &Slice[T]{ts: ts}
	return a
}

func (ss *Slice[T]) Copy() []T {
	ss.m.RLock()
	defer ss.m.RUnlock()
	if ss.ts == nil {
		return nil
	}
	var k = make([]T, len(ss.ts))
	copy(k, ss.ts)
	return k
}

// Replace 把ss.ts替换成ts
func (ss *Slice[T]) Replace(ts []T) {
	ss.m.Lock()
	defer ss.m.Unlock()
	ss.ts = ts
}

func (ss *Slice[T]) Get(i int) T {
	ss.m.RLock()
	defer ss.m.RUnlock()
	return ss.ts[i]
}

func (ss *Slice[T]) Len() int {
	ss.m.RLock()
	defer ss.m.RUnlock()
	return len(ss.ts)
}

func (ss *Slice[T]) Cap() int {
	ss.m.RLock()
	defer ss.m.RUnlock()
	return cap(ss.ts)
}

func (ss *Slice[T]) Set(i int, v T) {
	ss.m.Lock()
	defer ss.m.Unlock()
	ss.ts[i] = v
}

func (ss *Slice[T]) Append(ts ...T) *Slice[T] {
	ss.m.Lock()
	defer ss.m.Unlock()

	ss.ts = append(ss.ts, ts...)
	return ss
}

func (ss *Slice[T]) Slice(begin, end int) *Slice[T] {
	ss.m.RLock()
	defer ss.m.RUnlock()

	r := MakeSlice[T](end - begin)
	copy(r.ts, ss.ts[begin:end])
	return r
}

func (ss *Slice[T]) CopySlice(begin, end int) []T {
	ss.m.RLock()
	defer ss.m.RUnlock()
	var k = make([]T, end-begin)
	copy(k, ss.ts[begin:end])
	return k
}

// Foreach 对每个元素执行f
func (ss *Slice[T]) Foreach(f func(T)) {
	ss.m.RLock()
	defer ss.m.RUnlock()
	for _, t := range ss.ts {
		f(t)
	}
}

func (ss *Slice[T]) Range(f func(index int, value T) (shouldContinue bool)) {
	ss.m.RLock()
	defer ss.m.RUnlock()

	for i, t := range ss.ts {
		if !f(i, t) {
			return
		}
	}
}

func (ss *Slice[T]) FindIf(predicate func(T) bool) (i int, found bool) {
	ss.m.RLock()
	defer ss.m.RUnlock()
	return ss.findIf(predicate)
}

// DeleteOneIf 删除一个满足predicate的元素
func (ss *Slice[T]) DeleteOneIf(predicate func(T) bool) (deleted bool) {
	ss.m.Lock()
	defer ss.m.Unlock()
	i, found := ss.findIf(predicate)
	if found {
		ss.deleteIndex(i)
		return true
	}
	return
}

// DeleteAllIf 删除全部满足predicate的元素
func (ss *Slice[T]) DeleteAllIf(predicate func(T) bool) (deleted int) {
	ss.m.Lock()
	defer ss.m.Unlock()
	first, found := ss.findIf(predicate)
	if !found {
		return 0
	}
	for i := first; i < len(ss.ts); i++ {
		if !predicate(ss.ts[i]) {
			ss.ts[first] = ss.ts[i]
			first += 1
		}
	}
	deleted = len(ss.ts) - first
	ss.ts = ss.ts[:first]
	return
}

func (ss *Slice[T]) DeleteIndex(i int) {
	ss.m.Lock()
	defer ss.m.Unlock()
	ss.deleteIndex(i)
}

func (ss *Slice[T]) deleteIndex(i int) {
	ss.ts = append(ss.ts[:i], ss.ts[i+1:]...)
}

func (ss *Slice[T]) findIf(predicate func(T) bool) (i int, found bool) {
	for idx, item := range ss.ts {
		if predicate(item) {
			return idx, true
		}
	}
	return
}
