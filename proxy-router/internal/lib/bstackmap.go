package lib

import (
	"bytes"
	"fmt"
	"sync"
)

type BoundStackMap[T any] struct {
	capacity    int
	dataMap     map[string]T
	orderedKeys []string
	cc          int // current capacity
	m           sync.RWMutex
}

// NewBoundStackMap creates a new map with the limited capacity, where
// new records overwrite the oldest ones. The map is thread-safe
func NewBoundStackMap[T any](size int) *BoundStackMap[T] {
	return &BoundStackMap[T]{
		capacity:    size,
		dataMap:     make(map[string]T, size),
		orderedKeys: make([]string, 0, size),
		m:           sync.RWMutex{},
	}
}

func (bs *BoundStackMap[T]) Push(key string, item T) {
	bs.m.Lock()
	defer bs.m.Unlock()

	if bs.cc == bs.capacity {
		delete(bs.dataMap, bs.orderedKeys[0])
		bs.orderedKeys = bs.orderedKeys[1:]
	} else {
		bs.cc++
	}
	bs.orderedKeys = append(bs.orderedKeys, key)
	bs.dataMap[key] = item
}

func (bs *BoundStackMap[T]) Get(key string) (T, bool) {
	bs.m.RLock()
	defer bs.m.RUnlock()

	item, ok := bs.dataMap[key]
	return item, ok
}
func (bs *BoundStackMap[T]) At(index int) (T, bool) {
	bs.m.RLock()
	defer bs.m.RUnlock()
	// adjustment for negative index values to be counted from the end
	if index < 0 {
		index = bs.cc + index
	}
	// check if index is out of bounds
	if index < 0 || index > (bs.cc-1) {
		var nilValue T
		return nilValue, false
	}
	return bs.dataMap[bs.orderedKeys[index]], true
}

func (bs *BoundStackMap[T]) Clear() {
	bs.m.Lock()
	defer bs.m.Unlock()

	bs.cc = 0
	for k := range bs.dataMap {
		delete(bs.dataMap, k)
	}
	bs.orderedKeys = bs.orderedKeys[:0]
}

func (bs *BoundStackMap[T]) Count() int {
	bs.m.RLock()
	defer bs.m.RUnlock()

	return bs.cc
}

func (bs *BoundStackMap[T]) Capacity() int {
	bs.m.RLock()
	defer bs.m.RUnlock()

	return bs.capacity
}

func (bs *BoundStackMap[T]) Range(f func(key string, value T) bool) {
	bs.m.Lock()
	defer bs.m.Unlock()

	for key, value := range bs.dataMap {
		if !f(key, value) {
			return
		}
	}
}

func (bs *BoundStackMap[T]) Filter(f func(key string, value T) bool) {
	bs.m.Lock()
	defer bs.m.Unlock()

	for i := 0; i < len(bs.orderedKeys); i++ {
		key := bs.orderedKeys[i]
		value := bs.dataMap[key]
		if !f(key, value) {
			delete(bs.dataMap, key)
			bs.orderedKeys = append(bs.orderedKeys[:i], bs.orderedKeys[i+1:]...)
			bs.cc--
		}
	}
}

func (bs *BoundStackMap[T]) String() string {
	bs.m.RLock()
	defer bs.m.RUnlock()

	b := new(bytes.Buffer)
	for index, key := range bs.orderedKeys {
		fmt.Fprintf(b, "(%d) %s: %v\n", index, key, bs.dataMap[key])
	}
	return b.String()
}
